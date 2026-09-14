package validate

import "strings"

// FieldError is one field-level validation failure: the dotted/indexed path
// to the field, a message key (validator tag or business rule key), and an
// optional interpolation parameter.
type FieldError struct {
	Field string
	Tag   string
	Param string
}

// AppValidationError carries one or more FieldErrors plus the original
// sentinel error (Cause), so errors.Is/errors.As against the sentinel keeps
// working after wrapping.
type AppValidationError struct {
	Errors []FieldError
	Cause  error
}

func NewFieldError(field, tag string, cause error, param ...string) *AppValidationError {
	p := ""
	if len(param) > 0 {
		p = param[0]
	}
	return &AppValidationError{
		Errors: []FieldError{{Field: field, Tag: tag, Param: p}},
		Cause:  cause,
	}
}

func (e *AppValidationError) Error() string {
	if e.Cause != nil {
		return e.Cause.Error()
	}
	parts := make([]string, 0, len(e.Errors))
	for _, fe := range e.Errors {
		parts = append(parts, fe.Field+": "+fe.Tag)
	}
	return strings.Join(parts, "; ")
}

func (e *AppValidationError) Unwrap() error { return e.Cause }
