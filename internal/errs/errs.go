package errs

import (
	"errors"
	"fmt"
)

// Severity indicates how prominently an error should be surfaced to the user.
type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Error represents an application error enriched with metadata for user-facing handling.
type Error struct {
	msg      string
	code     string
	severity Severity
	fields   map[string]string
	cause    error
}

// Option mutates optional attributes on the Error during construction.
type Option func(*Error)

// WithSeverity sets the severity for the error.
func WithSeverity(sev Severity) Option {
	return func(e *Error) {
		e.severity = sev
	}
}

// WithFields attaches metadata pairs to the error.
func WithFields(fields map[string]string) Option {
	return func(e *Error) {
		if len(fields) == 0 {
			return
		}
		if e.fields == nil {
			e.fields = make(map[string]string, len(fields))
		}
		for k, v := range fields {
			e.fields[k] = v
		}
	}
}

// New creates a metadata rich error for user facing reporting.
func New(code, msg string, opts ...Option) *Error {
	e := &Error{code: code, msg: msg, severity: SeverityError}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Wrap decorates an existing error with additional metadata. The original error is preserved for unwrapping.
func Wrap(err error, code, msg string, opts ...Option) *Error {
	if err == nil {
		return nil
	}
	e := New(code, msg, opts...)
	e.cause = err
	return e
}

// From attempts to extract an *Error from an error chain.
func From(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.msg != "" {
		return e.msg
	}
	if e.code != "" {
		return e.code
	}
	return "unknown error"
}

// Unwrap returns the underlying cause if any.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// Code returns the identifying code for the error.
func (e *Error) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

// Severity exposes the severity classification.
func (e *Error) Severity() Severity {
	if e == nil || e.severity == "" {
		return SeverityError
	}
	return e.severity
}

// Fields returns a copy of the metadata map.
func (e *Error) Fields() map[string]string {
	if e == nil || len(e.fields) == 0 {
		return nil
	}
	out := make(map[string]string, len(e.fields))
	for k, v := range e.fields {
		out[k] = v
	}
	return out
}

// Format returns a formatted representation combining code and message.
func (e *Error) Format() string {
	if e == nil {
		return ""
	}
	if e.code != "" && e.msg != "" {
		return fmt.Sprintf("%s: %s", e.code, e.msg)
	}
	return e.Error()
}
