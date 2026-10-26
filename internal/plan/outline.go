package plan

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ibra2626/workey/internal/store"
)

// SaveOutline declaratively replaces the plan's phases and topic groups, each
// matched by key: a key present in the request is created or, if it already
// exists, updated in place when its fields changed; a key absent from the
// request is deleted. Deleting a topic group that sessions still reference
// fails with a typed validation error. The call is audited and, like SavePlan,
// idempotent (identical payload → same state, no new audit rows; a key replays
// its stored response).
func (s *Service) SaveOutline(ctx context.Context, source Source, req SaveOutlineRequest) (*SaveOutlineResult, error) {
	if !source.valid() {
		return nil, fmt.Errorf("plan: invalid source %q", source)
	}
	if err := validateSaveOutline(req); err != nil {
		return nil, err
	}

	hash, err := hashPayload(source, "save_outline", saveOutlinePayload{Phases: req.Phases, TopicGroups: req.TopicGroups})
	if err != nil {
		return nil, err
	}

	var result *SaveOutlineResult
	err = s.st.WithTx(ctx, func(tx *sql.Tx) error {
		if replay, done, err := s.replay(ctx, tx, req.IdempotencyKey, hash); err != nil {
			return err
		} else if done {
			result = &SaveOutlineResult{}
			if err := json.Unmarshal([]byte(replay), result); err != nil {
				return fmt.Errorf("plan: decode stored response: %w", err)
			}
			return nil
		}

		res, err := s.applyOutline(ctx, tx, source, hash, req)
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

func (s *Service) applyOutline(ctx context.Context, tx *sql.Tx, source Source, hash string, req SaveOutlineRequest) (*SaveOutlineResult, error) {
	now := s.nowUTC()
	result := &SaveOutlineResult{}

	tgResult, err := s.applyTopicGroups(ctx, tx, source, hash, now, req.TopicGroups)
	if err != nil {
		return nil, err
	}
	result.TopicGroups = *tgResult

	phResult, err := s.applyPhases(ctx, tx, source, hash, now, req.Phases)
	if err != nil {
		return nil, err
	}
	result.Phases = *phResult

	return result, nil
}

// applyTopicGroups replaces topic groups by key. A delete is refused (typed
// validation error) while sessions still reference the group.
func (s *Service) applyTopicGroups(ctx context.Context, tx *sql.Tx, source Source, hash, now string, incoming []TopicGroupInput) (*OutlineChange, error) {
	existing, err := s.st.ListTopicGroups(ctx, tx)
	if err != nil {
		return nil, err
	}
	byKey := map[string]store.TopicGroup{}
	for _, g := range existing {
		byKey[g.Key] = g
	}

	change := &OutlineChange{}
	seen := map[string]bool{}
	var verr errBuilder

	for _, in := range incoming {
		seen[in.Key] = true
		desired := store.TopicGroup{
			Key:            in.Key,
			Name:           in.Name,
			Ordinal:        in.Ordinal,
			DGSQuestionMin: in.DGSQuestionMin,
			DGSQuestionMax: in.DGSQuestionMax,
		}
		ex, ok := byKey[in.Key]
		if !ok {
			if err := s.st.InsertTopicGroup(ctx, tx, &desired); err != nil {
				return nil, err
			}
			if err := s.auditOutline(ctx, tx, source, actionCreate, "topic_group", desired.ID, nil, desired, hash, now); err != nil {
				return nil, err
			}
			change.Created = append(change.Created, in.Key)
			continue
		}
		if topicGroupEqual(ex, desired) {
			change.Unchanged = append(change.Unchanged, in.Key)
			continue
		}
		desired.ID = ex.ID
		if err := s.st.UpdateTopicGroupByKey(ctx, tx, desired); err != nil {
			return nil, err
		}
		if err := s.auditOutline(ctx, tx, source, actionUpdate, "topic_group", ex.ID, ex, desired, hash, now); err != nil {
			return nil, err
		}
		change.Updated = append(change.Updated, in.Key)
	}

	for _, ex := range existing {
		if seen[ex.Key] {
			continue
		}
		n, err := s.st.CountSessionsByTopicGroup(ctx, tx, ex.ID)
		if err != nil {
			return nil, err
		}
		if n > 0 {
			verr.add("topic_groups", CodeReferenced,
				fmt.Sprintf("topic group %q is referenced by %d session(s) and cannot be removed", ex.Key, n))
			continue
		}
		if err := s.st.DeleteTopicGroupByKey(ctx, tx, ex.Key); err != nil {
			return nil, err
		}
		if err := s.auditOutline(ctx, tx, source, actionDelete, "topic_group", ex.ID, ex, nil, hash, now); err != nil {
			return nil, err
		}
		change.Deleted = append(change.Deleted, ex.Key)
	}

	if ve := verr.result(); ve != nil {
		return nil, ve
	}
	return change, nil
}

// applyPhases replaces phases by key (phases carry no session references).
func (s *Service) applyPhases(ctx context.Context, tx *sql.Tx, source Source, hash, now string, incoming []PhaseInput) (*OutlineChange, error) {
	existing, err := s.st.ListPhases(ctx, tx)
	if err != nil {
		return nil, err
	}
	byKey := map[string]store.Phase{}
	for _, p := range existing {
		byKey[p.Key] = p
	}

	change := &OutlineChange{}
	seen := map[string]bool{}

	for _, in := range incoming {
		seen[in.Key] = true
		desired := store.Phase{
			Key:       in.Key,
			Ordinal:   in.Ordinal,
			Name:      in.Name,
			StartDate: in.StartDate,
			EndDate:   in.EndDate,
		}
		ex, ok := byKey[in.Key]
		if !ok {
			if err := s.st.InsertPhase(ctx, tx, &desired); err != nil {
				return nil, err
			}
			if err := s.auditOutline(ctx, tx, source, actionCreate, "phase", desired.ID, nil, desired, hash, now); err != nil {
				return nil, err
			}
			change.Created = append(change.Created, in.Key)
			continue
		}
		if phaseEqual(ex, desired) {
			change.Unchanged = append(change.Unchanged, in.Key)
			continue
		}
		desired.ID = ex.ID
		if err := s.st.UpdatePhaseByKey(ctx, tx, desired); err != nil {
			return nil, err
		}
		if err := s.auditOutline(ctx, tx, source, actionUpdate, "phase", ex.ID, ex, desired, hash, now); err != nil {
			return nil, err
		}
		change.Updated = append(change.Updated, in.Key)
	}

	for _, ex := range existing {
		if seen[ex.Key] {
			continue
		}
		if err := s.st.DeletePhaseByKey(ctx, tx, ex.Key); err != nil {
			return nil, err
		}
		if err := s.auditOutline(ctx, tx, source, actionDelete, "phase", ex.ID, ex, nil, hash, now); err != nil {
			return nil, err
		}
		change.Deleted = append(change.Deleted, ex.Key)
	}

	return change, nil
}

// auditOutline appends one audit row for a phase or topic-group change. before
// and after are marshalled to JSON; pass nil for a create's before or a
// delete's after.
func (s *Service) auditOutline(ctx context.Context, tx *sql.Tx, source Source, action, entity string, id int64, before, after any, hash, now string) error {
	beforeJSON, err := optionalJSON(before)
	if err != nil {
		return err
	}
	afterJSON, err := optionalJSON(after)
	if err != nil {
		return err
	}
	src := string(source)
	entry := store.AuditEntry{
		AtUTC:       now,
		Source:      src,
		Action:      action,
		Entity:      entity,
		EntityID:    &id,
		BeforeJSON:  beforeJSON,
		AfterJSON:   afterJSON,
		RequestHash: &hash,
	}
	return s.st.InsertAudit(ctx, tx, &entry)
}

// optionalJSON marshals v to a *string, returning nil for a nil interface.
func optionalJSON(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("plan: encode audit payload: %w", err)
	}
	s := string(buf)
	return &s, nil
}

func topicGroupEqual(a, b store.TopicGroup) bool {
	return a.Name == b.Name &&
		a.Ordinal == b.Ordinal &&
		intPtrEqual(a.DGSQuestionMin, b.DGSQuestionMin) &&
		intPtrEqual(a.DGSQuestionMax, b.DGSQuestionMax)
}

func phaseEqual(a, b store.Phase) bool {
	return a.Ordinal == b.Ordinal &&
		a.Name == b.Name &&
		a.StartDate == b.StartDate &&
		a.EndDate == b.EndDate
}

// saveOutlinePayload is the hashed view of a SaveOutline request.
type saveOutlinePayload struct {
	Phases      []PhaseInput      `json:"phases"`
	TopicGroups []TopicGroupInput `json:"topic_groups"`
}
