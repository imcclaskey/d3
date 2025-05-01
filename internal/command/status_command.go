package command

import (
	"context"
	"fmt"
	"strings"
	
	"github.com/imcclaskey/i3/internal/errors"
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
func (s Status) Run(ctx context.Context, cfg Config) (string, error) {
	// Validate i3 is initialized
	if err := validation.Init(cfg.I3Dir); err != nil {
		return "", err
	}
	
	// Get current session information
	active, feature, phase, err := cfg.Session.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get session status")
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
	if len(warnings) > 0 {
		message += "\n\nWarnings:\n" + strings.Join(warnings, "\n")
	}
	
	return message, nil
} 