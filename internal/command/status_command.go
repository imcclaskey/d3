package command

import (
	"context"
	"fmt"
	
	"github.com/imcclaskey/i3/internal/validation"
)

// Status displays the current i3 status
type Status struct {
	// No parameters needed for this command
}

// NewStatus creates a new status command
func NewStatus() Status {
	return Status{}
}

// Run implements the Command interface
func (s Status) Run(ctx context.Context, cfg Config) (Result, error) {
	// Validate i3 is initialized
	if err := validation.Init(cfg.I3Dir); err != nil {
		// Return zero Result on error
		return Result{}, err
	}
	
	// Get current session information
	active, feature, phase, err := cfg.Session.Status()
	if err != nil {
		// Return zero Result on error
		return Result{}, fmt.Errorf("failed to get session status: %w", err)
	}
	
	// Format status message
	var message string
	if !active {
		message = "No active i3 session"
	} else {
		if feature == "" {
			if phase == "setup" {
				message = "Currently in setup phase"
			} else {
				message = fmt.Sprintf("Currently in %s phase (no feature set)", phase)
			}
		} else {
			message = fmt.Sprintf("Currently in %s phase of feature '%s'", phase, feature)
		}
	}
	
	// Collect any warnings and append to message
	warnings := validation.ContentWarnings(cfg.I3Dir)
	// Do not append warnings to message anymore
	/*
		if len(warnings) > 0 {
			message += "\n\nWarnings:\n" + strings.Join(warnings, "\n")
		}
	*/
	
	// Return Result struct with message and warnings separated
	return NewResult(message, nil, warnings), nil
} 