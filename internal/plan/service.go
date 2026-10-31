package plan

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ibra2626/workey/internal/clock"
	"github.com/ibra2626/workey/internal/store"
)

// timestampLayout is ISO-8601 UTC with second precision, used for created_at,
// updated_at and audit times.
const timestampLayout = "2006-01-02T15:04:05Z"

// Audit actions.
const (
	actionCreate = "create"
	actionUpdate = "update"
	actionDelete = "delete"
)

// Service is the single entry for writing the plan. It validates, freezes the
// past, de-duplicates by idempotency and audits every change, delegating
// persistence to store and "today" to an injectable clock.
type Service struct {
	st  *store.Store
	clk clock.Clock
}

// New builds a Service over the given store and clock.
func New(st *store.Store, clk clock.Clock) *Service {
	return &Service{st: st, clk: clk}
}

// nowUTC is the current instant as an ISO-8601 UTC string.
func (s *Service) nowUTC() string {
	return s.clk.Now().UTC().Format(timestampLayout)
}

// SavePlan declaratively replaces the open sessions for the dates in
// [req.From, req.To]: for each such date that is today or later, the open
// (uncompleted) sessions become exactly the given set, matched by
// (date, start_time). An incoming session keeps the id of the one it matches,
// a changed one is updated in place, and an open session absent from the
// request is deleted unless it has a study record. Started sessions retain
// their slot and content: equal content is unchanged, changes are skipped with
// reason started, and omitted started sessions are kept.
// Sessions dated before today and completed sessions are
// frozen: they are never modified or deleted, and an incoming session that
// targets one is returned in Skipped.
//
// The call is idempotent: the same payload applied twice leaves the same state
// and writes no audit rows the second time. With req.IdempotencyKey the stored
// response is replayed for a repeat, and a different payload under the same key
// is rejected with ErrIdempotencyConflict.
func (s *Service) SavePlan(ctx context.Context, source Source, req SavePlanRequest) (*SavePlanResult, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	if err := validateSavePlan(req); err != nil {
		return nil, err
	}

	hash, err := hashPayload(source, "save_plan", savePlanPayload{From: req.From, To: req.To, Sessions: req.Sessions})
	if err != nil {
		return nil, err
	}
	today := clock.Today(s.clk).String()

	var result *SavePlanResult
	err = s.st.WithTx(ctx, func(tx *sql.Tx) error {
		if replay, done, err := s.replay(ctx, tx, req.IdempotencyKey, hash); err != nil {
			return err
		} else if done {
			result = &SavePlanResult{}
			if err := json.Unmarshal([]byte(replay), result); err != nil {
				return fmt.Errorf("plan: decode stored response: %w", err)
			}
			return nil
		}

		res, err := s.applyPlan(ctx, tx, source, today, hash, req)
		if err != nil {
			return err
		}
		result = res
		return s.remember(ctx, tx, req.IdempotencyKey, hash, res)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// applyPlan performs the range-replace against the store inside tx.
func (s *Service) applyPlan(ctx context.Context, tx *sql.Tx, source Source, today, hash string, req SavePlanRequest) (*SavePlanResult, error) {
	now := s.nowUTC()
	result := &SavePlanResult{}

	existing, err := s.st.SessionsByDateRange(ctx, tx, req.From, req.To)
	if err != nil {
		return nil, err
	}
	existingBySlot := map[SlotRef]store.Session{}
	for _, e := range existing {
		existingBySlot[SlotRef{Date: e.Date, StartTime: e.StartTime}] = e
	}
	started, err := s.st.StartedSessionIDs(ctx, tx, req.From, req.To)
	if err != nil {
		return nil, err
	}

	// Classify each incoming session and resolve store-dependent references
	// (topic group, makeup target) before any write, collecting typed errors.
	type applyItem struct {
		input    SessionInput
		existing *store.Session // non-nil when an open session already holds the slot
		tgID     *int64
		makeupID *int64
	}
	var toApply []applyItem
	var verr errBuilder
	seen := map[SlotRef]bool{}

	for i, si := range req.Sessions {
		slot := si.slot()
		if si.Date < today {
			result.Skipped = append(result.Skipped, SkippedSession{Date: si.Date, StartTime: si.StartTime, Reason: SkipPastDate})
			continue
		}
		ex, ok := existingBySlot[slot]
		if ok && ex.CompletedAt != nil {
			result.Skipped = append(result.Skipped, SkippedSession{Date: si.Date, StartTime: si.StartTime, Reason: SkipCompleted})
			continue
		}

		p := fmt.Sprintf("sessions[%d]", i)
		tgID, err := s.resolveTopicGroup(ctx, tx, &verr, p, si.TopicGroup)
		if err != nil {
			return nil, err
		}
		makeupID, err := s.resolveMakeup(ctx, tx, &verr, p, today, si.MakeupOf)
		if err != nil {
			return nil, err
		}

		item := applyItem{input: si, tgID: tgID, makeupID: makeupID}
		if ok {
			exCopy := ex
			item.existing = &exCopy
		}
		toApply = append(toApply, item)
		seen[slot] = true
	}

	if ve := verr.result(); ve != nil {
		return nil, ve
	}

	for _, a := range toApply {
		items := toStoreItems(a.input.Items)
		if a.existing != nil {
			if contentEqual(*a.existing, a.input.DurationMin, a.input.Kind, a.tgID, a.makeupID, items) {
				result.Unchanged = append(result.Unchanged, refOf(*a.existing))
				continue
			}
			if started[a.existing.ID] {
				result.Skipped = append(result.Skipped, SkippedSession{Date: a.input.Date, StartTime: a.input.StartTime, Reason: SkipStarted})
				continue
			}
			before, err := snapshotJSON(*a.existing)
			if err != nil {
				return nil, err
			}
			updated := *a.existing
			updated.DurationMin = a.input.DurationMin
			updated.Kind = a.input.Kind
			updated.TopicGroupID = a.tgID
			updated.MakeupOf = a.makeupID
			updated.UpdatedAt = now
			if err := s.st.UpdateSession(ctx, tx, &updated); err != nil {
				return nil, err
			}
			if err := s.st.DeleteSessionItems(ctx, tx, updated.ID); err != nil {
				return nil, err
			}
			if err := s.insertItems(ctx, tx, updated.ID, items); err != nil {
				return nil, err
			}
			updated.Items = withSessionID(items, updated.ID)
			after, err := snapshotJSON(updated)
			if err != nil {
				return nil, err
			}
			if err := s.audit(ctx, tx, source, actionUpdate, updated.ID, &before, &after, hash, now); err != nil {
				return nil, err
			}
			result.Updated = append(result.Updated, refOf(updated))
			continue
		}

		created := store.Session{
			Date:         a.input.Date,
			StartTime:    a.input.StartTime,
			DurationMin:  a.input.DurationMin,
			Kind:         a.input.Kind,
			TopicGroupID: a.tgID,
			MakeupOf:     a.makeupID,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := s.st.InsertSession(ctx, tx, &created); err != nil {
			return nil, err
		}
		if err := s.insertItems(ctx, tx, created.ID, items); err != nil {
			return nil, err
		}
		created.Items = withSessionID(items, created.ID)
		after, err := snapshotJSON(created)
		if err != nil {
			return nil, err
		}
		if err := s.audit(ctx, tx, source, actionCreate, created.ID, nil, &after, hash, now); err != nil {
			return nil, err
		}
		result.Created = append(result.Created, refOf(created))
	}

	// Delete the open future sessions no longer in the request. Frozen
	// sessions (completed, dated before today, or started) are kept.
	for slot, ex := range existingBySlot {
		if seen[slot] || ex.CompletedAt != nil || ex.Date < today || started[ex.ID] {
			continue
		}
		before, err := snapshotJSON(ex)
		if err != nil {
			return nil, err
		}
		if err := s.st.DeleteSession(ctx, tx, ex.ID); err != nil {
			return nil, err
		}
		if err := s.audit(ctx, tx, source, actionDelete, ex.ID, &before, nil, hash, now); err != nil {
			return nil, err
		}
		result.Deleted = append(result.Deleted, refOf(ex))
	}

	sortRefs(result.Created)
	sortRefs(result.Updated)
	sortRefs(result.Deleted)
	sortRefs(result.Unchanged)
	sort.Slice(result.Skipped, func(i, j int) bool {
		if result.Skipped[i].Date != result.Skipped[j].Date {
			return result.Skipped[i].Date < result.Skipped[j].Date
		}
		return result.Skipped[i].StartTime < result.Skipped[j].StartTime
	})
	return result, nil
}

// sortRefs orders session refs by (date, start_time) for a stable response.
func sortRefs(refs []SessionRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Date != refs[j].Date {
			return refs[i].Date < refs[j].Date
		}
		return refs[i].StartTime < refs[j].StartTime
	})
}

// resolveTopicGroup maps a topic-group key to its id, recording a field error
// when the key is set but unknown. An empty key yields a nil id and no error.
func (s *Service) resolveTopicGroup(ctx context.Context, tx *sql.Tx, verr *errBuilder, path, key string) (*int64, error) {
	if key == "" {
		return nil, nil
	}
	g, err := s.st.TopicGroupByKey(ctx, tx, key)
	if err != nil {
		return nil, err
	}
	if g == nil {
		verr.add(path+".topic_group", CodeNotFound, fmt.Sprintf("topic group %q does not exist", key))
		return nil, nil
	}
	return &g.ID, nil
}

// resolveMakeup resolves a makeup target slot to its session id, requiring that
// it names a missed session: one that exists, is dated before today and is not
// completed. A nil ref yields a nil id.
func (s *Service) resolveMakeup(ctx context.Context, tx *sql.Tx, verr *errBuilder, path, today string, ref *SlotRef) (*int64, error) {
	if ref == nil {
		return nil, nil
	}
	target, err := s.st.SessionByDateAndTime(ctx, tx, ref.Date, ref.StartTime)
	if err != nil {
		return nil, err
	}
	if target == nil {
		verr.add(path+".makeup_of", CodeNotFound, "makeup_of does not reference an existing session")
		return nil, nil
	}
	if target.Date >= today || target.CompletedAt != nil {
		verr.add(path+".makeup_of", CodeMakeupTarget, "makeup_of must reference a past, uncompleted session")
		return nil, nil
	}
	return &target.ID, nil
}

// insertItems inserts items under sessionID, numbering ordinals from zero.
func (s *Service) insertItems(ctx context.Context, tx *sql.Tx, sessionID int64, items []store.SessionItem) error {
	for i := range items {
		it := items[i]
		it.SessionID = sessionID
		it.Ordinal = i
		if err := s.st.InsertSessionItem(ctx, tx, &it); err != nil {
			return err
		}
	}
	return nil
}

// audit appends one audit row for a session change.
func (s *Service) audit(ctx context.Context, tx *sql.Tx, source Source, action string, id int64, before, after *string, hash, now string) error {
	src := string(source)
	entry := store.AuditEntry{
		AtUTC:       now,
		Source:      src,
		Action:      action,
		Entity:      "session",
		EntityID:    &id,
		BeforeJSON:  before,
		AfterJSON:   after,
		RequestHash: &hash,
	}
	return s.st.InsertAudit(ctx, tx, &entry)
}

// replay returns the stored response for key when it holds the same request
// hash (done=true). A different hash under the same key is a conflict. An empty
// key never replays.
func (s *Service) replay(ctx context.Context, tx *sql.Tx, key, hash string) (resp string, done bool, err error) {
	if key == "" {
		return "", false, nil
	}
	rec, err := s.st.GetIdempotency(ctx, tx, key)
	if err != nil {
		return "", false, err
	}
	if rec == nil {
		return "", false, nil
	}
	if rec.RequestHash != hash {
		return "", false, ErrIdempotencyConflict
	}
	return rec.ResponseJSON, true, nil
}

// remember stores the response under key for later replay. An empty key is a
// no-op.
func (s *Service) remember(ctx context.Context, tx *sql.Tx, key, hash string, result any) error {
	if key == "" {
		return nil
	}
	buf, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("plan: encode response: %w", err)
	}
	return s.st.PutIdempotency(ctx, tx, store.IdempotencyRecord{
		Key:          key,
		RequestHash:  hash,
		ResponseJSON: string(buf),
		CreatedAt:    s.nowUTC(),
	})
}

// --- helpers on plain values (no store/clock) ---

// toStoreItems converts request items to store items (ids and ordinals set on
// insert).
func toStoreItems(items []Item) []store.SessionItem {
	out := make([]store.SessionItem, 0, len(items))
	for i, it := range items {
		out = append(out, store.SessionItem{
			Ordinal:       i,
			Book:          it.Book,
			ChapterNo:     it.ChapterNo,
			ChapterTitle:  it.ChapterTitle,
			PageFrom:      it.PageFrom,
			PageTo:        it.PageTo,
			KonuFrom:      it.KonuFrom,
			KonuTo:        it.KonuTo,
			SoruFrom:      it.SoruFrom,
			SoruTo:        it.SoruTo,
			AntrenmanFrom: it.AntrenmanFrom,
			AntrenmanTo:   it.AntrenmanTo,
		})
	}
	return out
}

// withSessionID returns a copy of items bound to sessionID with sequential
// ordinals, for building an accurate after-snapshot.
func withSessionID(items []store.SessionItem, sessionID int64) []store.SessionItem {
	out := make([]store.SessionItem, len(items))
	copy(out, items)
	for i := range out {
		out[i].SessionID = sessionID
		out[i].Ordinal = i
	}
	return out
}

// refOf summarises a session as a SessionRef.
func refOf(s store.Session) SessionRef {
	return SessionRef{ID: s.ID, Date: s.Date, StartTime: s.StartTime}
}

// contentEqual reports whether an existing session already has the given
// content, so no update (and no audit row) is needed.
func contentEqual(ex store.Session, dur int, kind string, tg, makeup *int64, items []store.SessionItem) bool {
	if ex.DurationMin != dur || ex.Kind != kind {
		return false
	}
	if !int64PtrEqual(ex.TopicGroupID, tg) || !int64PtrEqual(ex.MakeupOf, makeup) {
		return false
	}
	if len(ex.Items) != len(items) {
		return false
	}
	for i := range items {
		if !itemContentEqual(ex.Items[i], items[i]) {
			return false
		}
	}
	return true
}

// itemContentEqual compares two items ignoring ids and session ids.
func itemContentEqual(a, b store.SessionItem) bool {
	return a.Book == b.Book &&
		a.ChapterNo == b.ChapterNo &&
		a.ChapterTitle == b.ChapterTitle &&
		a.PageFrom == b.PageFrom &&
		a.PageTo == b.PageTo &&
		intPtrEqual(a.KonuFrom, b.KonuFrom) &&
		intPtrEqual(a.KonuTo, b.KonuTo) &&
		intPtrEqual(a.SoruFrom, b.SoruFrom) &&
		intPtrEqual(a.SoruTo, b.SoruTo) &&
		intPtrEqual(a.AntrenmanFrom, b.AntrenmanFrom) &&
		intPtrEqual(a.AntrenmanTo, b.AntrenmanTo)
}

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// snapshotJSON renders a session's content as the JSON kept in audit before/after.
func snapshotJSON(s store.Session) (string, error) {
	type snapItem struct {
		Book          string `json:"book"`
		ChapterNo     int    `json:"chapter_no"`
		ChapterTitle  string `json:"chapter_title"`
		PageFrom      int    `json:"page_from"`
		PageTo        int    `json:"page_to"`
		KonuFrom      *int   `json:"konu_from,omitempty"`
		KonuTo        *int   `json:"konu_to,omitempty"`
		SoruFrom      *int   `json:"soru_from,omitempty"`
		SoruTo        *int   `json:"soru_to,omitempty"`
		AntrenmanFrom *int   `json:"antrenman_from,omitempty"`
		AntrenmanTo   *int   `json:"antrenman_to,omitempty"`
	}
	type snap struct {
		ID           int64      `json:"id"`
		Date         string     `json:"date"`
		StartTime    string     `json:"start_time"`
		DurationMin  int        `json:"duration_min"`
		Kind         string     `json:"kind"`
		TopicGroupID *int64     `json:"topic_group_id,omitempty"`
		MakeupOf     *int64     `json:"makeup_of,omitempty"`
		CompletedAt  *string    `json:"completed_at,omitempty"`
		Items        []snapItem `json:"items"`
	}
	sn := snap{
		ID:           s.ID,
		Date:         s.Date,
		StartTime:    s.StartTime,
		DurationMin:  s.DurationMin,
		Kind:         s.Kind,
		TopicGroupID: s.TopicGroupID,
		MakeupOf:     s.MakeupOf,
		CompletedAt:  s.CompletedAt,
	}
	for _, it := range s.Items {
		sn.Items = append(sn.Items, snapItem{
			Book: it.Book, ChapterNo: it.ChapterNo, ChapterTitle: it.ChapterTitle,
			PageFrom: it.PageFrom, PageTo: it.PageTo,
			KonuFrom: it.KonuFrom, KonuTo: it.KonuTo,
			SoruFrom: it.SoruFrom, SoruTo: it.SoruTo,
			AntrenmanFrom: it.AntrenmanFrom, AntrenmanTo: it.AntrenmanTo,
		})
	}
	buf, err := json.Marshal(sn)
	if err != nil {
		return "", fmt.Errorf("plan: encode snapshot: %w", err)
	}
	return string(buf), nil
}

// savePlanPayload is the hashed view of a SavePlan request (the idempotency key
// itself is excluded, so a repeat under the same key is recognised).
type savePlanPayload struct {
	From     string         `json:"from"`
	To       string         `json:"to"`
	Sessions []SessionInput `json:"sessions"`
}

// hashPayload produces a stable sha256 hex digest of a request payload under an
// operation name and source.
func hashPayload(source Source, op string, payload any) (string, error) {
	envelope := struct {
		Source  Source `json:"source"`
		Op      string `json:"op"`
		Payload any    `json:"payload"`
	}{source, op, payload}
	buf, err := json.Marshal(envelope)
	if err != nil {
		return "", fmt.Errorf("plan: hash payload: %w", err)
	}
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:]), nil
}
