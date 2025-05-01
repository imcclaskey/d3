package command

import (
	"context"
	"fmt"

	i3context "github.com/imcclaskey/i3/internal/context"
)

// Status displays the current i3 context.
type Status struct{}

// NewStatus creates a new Status command
func NewStatus() Status {
	return Status{}
}

// Run implements the Command interface
func (s Status) Run(ctx context.Context, cfg Config) (Result, error) {
	// Load current context
	currentCtx, err := i3context.LoadContext(cfg.I3Dir)
	if err != nil {
		// If context file is simply missing/empty, that's not an error for status,
		// but other errors (read permission, bad JSON) should be reported.
		// LoadContext already handles ErrNotExist returning empty context and no error.
		return Result{}, fmt.Errorf("loading context: %w", err)
	}

	var message string
	if currentCtx.Feature == "" {
		message = "No active feature context."
	} else {
		featureMsg := fmt.Sprintf("Current Feature: %s", currentCtx.Feature)
		phaseMsg := "(No phase set)"
		if currentCtx.Phase != "" {
			phaseMsg = fmt.Sprintf("Current Phase: %s", currentCtx.Phase)
		}
		message = fmt.Sprintf("%s\n%s", featureMsg, phaseMsg)
	}

	return NewResult(message, currentCtx, nil), nil
} 