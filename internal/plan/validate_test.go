package plan

import "testing"

// codesFor runs validateSavePlan and returns a map of field path -> code.
func codesFor(t *testing.T, req SavePlanRequest) map[string]string {
	t.Helper()
	err := validateSavePlan(req)
	ve, ok := AsValidationError(err)
	if !ok {
		t.Fatalf("expected a *ValidationError, got %v", err)
	}
	m := map[string]string{}
	for _, fe := range ve.Errors {
		m[fe.Path] = fe.Code
	}
	return m
}

func TestValidateSavePlanStructural(t *testing.T) {
	req := SavePlanRequest{
		From: "2026-11-05",
		To:   "2026-11-01", // before from
		Sessions: []SessionInput{
			{
				Date:        "2026-10-01", // out of range
				StartTime:   "25:00",      // invalid time
				DurationMin: 0,            // invalid
				Kind:        "nope",       // free label now — not an error
				Items: []Item{
					// empty source key; a unit range without a section; and a
					// reversed location range.
					{SourceKey: "", LocFrom: ptr(58), LocTo: ptr(41), UnitFrom: ptr(5)},
				},
			},
		},
	}
	codes := codesFor(t, req)

	want := map[string]string{
		"to":                                   CodeRangeOrder,
		"sessions[0].date":                     CodeOutOfRange,
		"sessions[0].start_time":               CodeInvalid,
		"sessions[0].duration_min":             CodeInvalid,
		"sessions[0].items[0].source_key":      CodeRequired,
		"sessions[0].items[0].section_ordinal": CodeRequired, // unit range needs a section
		"sessions[0].items[0].unit_from":       CodeInvalid,  // unit_to missing
		"sessions[0].items[0].loc_to":          CodeRangeOrder,
	}
	for path, code := range want {
		if codes[path] != code {
			t.Errorf("path %s: got code %q, want %q", path, codes[path], code)
		}
	}
	// kind is a free label since WKY-22; no validation error for an unknown kind.
	if _, ok := codes["sessions[0].kind"]; ok {
		t.Errorf("kind must not be validated as an enum, got %v", codes["sessions[0].kind"])
	}
}

func TestValidateSavePlanDuplicateSlot(t *testing.T) {
	req := SavePlanRequest{
		From: "2026-10-26",
		To:   "2026-10-26",
		Sessions: []SessionInput{
			{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{SourceKey: "M1", SectionOrdinal: ptr(1), LocFrom: ptr(1), LocTo: ptr(2)}}},
			{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{SourceKey: "M1", SectionOrdinal: ptr(2), LocFrom: ptr(3), LocTo: ptr(4)}}},
		},
	}
	codes := codesFor(t, req)
	if codes["sessions[1]"] != CodeDuplicate {
		t.Errorf("expected duplicate-slot error on sessions[1], got %v", codes)
	}
}

func TestValidateSavePlanEmptyItems(t *testing.T) {
	req := SavePlanRequest{
		From: "2026-10-26", To: "2026-10-26",
		Sessions: []SessionInput{{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam}},
	}
	codes := codesFor(t, req)
	if codes["sessions[0].items"] != CodeRequired {
		t.Errorf("expected required-items error, got %v", codes)
	}
}

func TestValidateSavePlanValid(t *testing.T) {
	req := SavePlanRequest{
		From: "2026-10-26", To: "2026-10-26",
		Sessions: []SessionInput{{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{SourceKey: "M1", SectionOrdinal: ptr(1), LocFrom: ptr(1), LocTo: ptr(22)}}}},
	}
	if err := validateSavePlan(req); err != nil {
		t.Errorf("expected valid request, got %v", err)
	}
}
