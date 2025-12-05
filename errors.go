package inrequest

import "fmt"

// ParseError represents an error that occurred during request parsing
type ParseError struct {
	Type    string // Type of request being parsed (form, json, query)
	Message string // Error message
	Err     error  // Underlying error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("inrequest: %s parse error: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("inrequest: %s parse error: %s", e.Type, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// BindError represents an error that occurred during struct binding
type BindError struct {
	Field   string // Field name that caused the error (if known)
	Message string // Error message
	Err     error  // Underlying error
}

func (e *BindError) Error() string {
	if e.Field != "" {
		if e.Err != nil {
			return fmt.Sprintf("inrequest: bind error on field '%s': %s: %v", e.Field, e.Message, e.Err)
		}
		return fmt.Sprintf("inrequest: bind error on field '%s': %s", e.Field, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("inrequest: bind error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("inrequest: bind error: %s", e.Message)
}

func (e *BindError) Unwrap() error {
	return e.Err
}

// NewParseError creates a new ParseError
func NewParseError(reqType, message string, err error) *ParseError {
	return &ParseError{
		Type:    reqType,
		Message: message,
		Err:     err,
	}
}

// NewBindError creates a new BindError
func NewBindError(field, message string, err error) *BindError {
	return &BindError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// IsParseError returns true if the error is a ParseError
func IsParseError(err error) bool {
	_, ok := err.(*ParseError)
	return ok
}

// IsBindError returns true if the error is a BindError
func IsBindError(err error) bool {
	_, ok := err.(*BindError)
	return ok
}
