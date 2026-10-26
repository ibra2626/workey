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
				Kind:        "nope",       // invalid enum
				Items: []Item{
					{Book: "ZZ", ChapterNo: 0, ChapterTitle: "", PageFrom: 58, PageTo: 41, SoruFrom: ptr(5)},
				},
			},
		},
	}
	codes := codesFor(t, req)

	want := map[string]string{
		"to":                                 CodeRangeOrder,
		"sessions[0].date":                   CodeOutOfRange,
		"sessions[0].start_time":             CodeInvalid,
		"sessions[0].duration_min":           CodeInvalid,
		"sessions[0].kind":                   CodeInvalid,
		"sessions[0].items[0].book":          CodeInvalid,
		"sessions[0].items[0].chapter_no":    CodeInvalid,
		"sessions[0].items[0].chapter_title": CodeRequired,
		"sessions[0].items[0].page_to":       CodeRangeOrder,
		"sessions[0].items[0].soru_from":     CodeInvalid, // soru_to missing
	}
	for path, code := range want {
		if codes[path] != code {
			t.Errorf("path %s: got code %q, want %q", path, codes[path], code)
		}
	}
}

func TestValidateSavePlanDuplicateSlot(t *testing.T) {
	req := SavePlanRequest{
		From: "2026-10-26",
		To:   "2026-10-26",
		Sessions: []SessionInput{
			{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{Book: "M1", ChapterNo: 1, ChapterTitle: "A", PageFrom: 1, PageTo: 2}}},
			{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{Book: "M1", ChapterNo: 2, ChapterTitle: "B", PageFrom: 3, PageTo: 4}}},
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
		Sessions: []SessionInput{{Date: "2026-10-26", StartTime: "20:30", DurationMin: 90, Kind: KindAksam, Items: []Item{{Book: "M1", ChapterNo: 1, ChapterTitle: "A", PageFrom: 1, PageTo: 22}}}},
	}
	if err := validateSavePlan(req); err != nil {
		t.Errorf("expected valid request, got %v", err)
	}
}
