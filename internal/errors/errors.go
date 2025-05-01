package errors

import (
	"errors"
	"fmt"
)

// Standard errors for the i3 application
var (
	// ErrNotInitialized indicates the i3 workspace is not initialized
	ErrNotInitialized = errors.New("i3 is not initialized")
	
	// ErrAlreadyInitialized indicates i3 is already initialized
	ErrAlreadyInitialized = errors.New("i3 is already initialized")
	
	// ErrMissingFile indicates a required file is missing
	ErrMissingFile = errors.New("required file is missing")
	
	// ErrInvalidPhase indicates an invalid phase name was provided
	ErrInvalidPhase = errors.New("invalid phase name")
	
	// ErrFeatureNotFound indicates a specified feature could not be found
	ErrFeatureNotFound = errors.New("feature not found")
	
	// ErrNoActiveSession indicates no session is currently active
	ErrNoActiveSession = errors.New("no active session")
	
	// ErrPermission indicates a file permission issue
	ErrPermission = errors.New("permission denied")
	
	// ErrCursorIntegration indicates cursor integration is missing or invalid
	ErrCursorIntegration = errors.New("cursor integration issue")
)

// withDetails adds details to an error without wrapping
func withDetails(err error, details string) error {
	return fmt.Errorf("%v: %s", err, details)
}

// WithDetails adds context details to a standard error
func WithDetails(err error, details string) error {
	return withDetails(err, details)
}

// withSuggestion adds a suggestion to an error message
func withSuggestion(err error, suggestion string) error {
	return fmt.Errorf("%v (suggestion: %s)", err, suggestion)
}

// WithSuggestion adds a recommendation to an error
func WithSuggestion(err error, suggestion string) error {
	return withSuggestion(err, suggestion)
}

// Is reports whether any error in err's tree matches target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's tree that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// New creates a new error with the given text
func New(text string) error {
	return errors.New(text)
}

// Errorf creates a new error with formatting
func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
} 