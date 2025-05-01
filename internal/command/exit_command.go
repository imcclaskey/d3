package command

import (
	"context"
	"fmt"
	
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
func (e Exit) Run(ctx context.Context, cfg Config) (Result, error) {
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
	
	// If not active, return a message but not an error
	if !active {
		// Return Result struct (no error)
		return NewResult("No active i3 session to exit", nil, nil), nil
	}
	
	// Update session to inactive
	if err := cfg.Session.Stop(); err != nil {
		// Return zero Result on error
		return Result{}, fmt.Errorf("failed to stop session: %w", err)
	}
	
	// Format success message
	if feature == "" {
		feature = "unknown"
	}
	
	if phase == "" {
		phase = "unknown"
	}
	
	message := fmt.Sprintf("Exited %s phase of feature '%s'", phase, feature)
	// Return Result struct on success
	return NewResult(message, nil, nil), nil
} 