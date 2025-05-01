package command

import (
	"context"
	"fmt"
	
	"github.com/imcclaskey/i3/internal/errors"
	"github.com/imcclaskey/i3/internal/validation"
)

// Exit terminates the current i3 session
type Exit struct {
	// No parameters needed for this command
}

// NewExit creates an exit command
func NewExit() Exit {
	return Exit{}
}

// Run implements the Command interface
func (e Exit) Run(ctx context.Context, cfg Config) (string, error) {
	// Validate i3 is initialized
	if err := validation.Init(cfg.I3Dir); err != nil {
		return "", err
	}
	
	// Get current session information
	active, feature, phase, err := cfg.Session.Status()
	if err != nil {
		return "", errors.Wrap(err, "failed to get session status")
	}
	
	// If not active, return a message but not an error
	if !active {
		return "No active i3 session to exit", nil
	}
	
	// Update session to inactive
	if err := cfg.Session.Stop(); err != nil {
		return "", errors.Wrap(err, "failed to stop session")
	}
	
	// Format success message
	if feature == "" {
		feature = "unknown"
	}
	
	if phase == "" {
		phase = "unknown"
	}
	
	return fmt.Sprintf("Exited %s phase of feature '%s'", phase, feature), nil
} 