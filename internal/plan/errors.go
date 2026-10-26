package plan

import (
	"errors"
	"fmt"
	"strings"
)

// FieldError is one typed validation failure: a field path into the request, a
// stable machine code and an English message. REST and MCP render these without
// re-checking the rules.
type FieldError struct {
	Path    string `json:"path"`    // e.g. "sessions[0].items[1].page_to"
	Code    string `json:"code"`    // e.g. "range_order"
	Message string `json:"message"` // English, human-readable
}

// ValidationError collects every FieldError found in one request, so a caller
// sees all problems at once rather than one at a time.
type ValidationError struct {
	Errors []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fmt.Sprintf("%s: %s (%s)", fe.Path, fe.Message, fe.Code))
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// AsValidationError returns the *ValidationError in err's chain, if any.
func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	if errors.As(err, &ve) {
		return ve, true
	}
	return nil, false
}

// ErrIdempotencyConflict is returned when an idempotency key is reused with a
// different payload.
var ErrIdempotencyConflict = errors.New("plan: idempotency key reused with a different request")

// Validation field-error codes.
const (
	CodeRequired     = "required"      // a required field is empty or missing
	CodeInvalid      = "invalid"       // a malformed value (bad date, time, enum)
	CodeRangeOrder   = "range_order"   // from is greater than to
	CodeOutOfRange   = "out_of_range"  // a date falls outside [from, to]
	CodeDuplicate    = "duplicate"     // two entries collide on a unique key
	CodeNotFound     = "not_found"     // a referenced entity does not exist
	CodeMakeupTarget = "makeup_target" // makeup_of does not point at a missed session
	CodeReferenced   = "referenced"    // deleting an entity still referenced by sessions
	CodeFuture       = "future"        // completing or undoing a session dated after today
)

// fieldError builds a *ValidationError carrying a single field error, for the
// store-dependent checks that fail on one target (completion, undo).
func fieldError(path, code, message string) *ValidationError {
	return &ValidationError{Errors: []FieldError{{Path: path, Code: code, Message: message}}}
}

// errBuilder accumulates FieldErrors during validation.
type errBuilder struct {
	errs []FieldError
}

// add records a field error.
func (b *errBuilder) add(path, code, message string) {
	b.errs = append(b.errs, FieldError{Path: path, Code: code, Message: message})
}

// result returns a *ValidationError if any errors were added, else nil.
func (b *errBuilder) result() error {
	if len(b.errs) == 0 {
		return nil
	}
	return &ValidationError{Errors: b.errs}
}
