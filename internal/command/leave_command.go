package command

import (
	"context"
	"fmt"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/rules"
)

// Leave clears the current working feature context.
type Leave struct{}

// NewLeave creates a new Leave command
func NewLeave() Leave {
	return Leave{}
}

// Run implements the Command interface
func (l Leave) Run(ctx context.Context, cfg Config) (Result, error) {
	// Get current context first (to check if we need to clean up)
	currentCtx, err := i3context.LoadContext(cfg.I3Dir)
	if err != nil {
		return Result{}, fmt.Errorf("loading current context: %w", err)
	}

	// Clear rule files if they exist
	if currentCtx.Feature != "" || currentCtx.Phase != "" {
		ruleGenerator := rules.NewRuleFileGenerator("", cfg.CursorRulesDir)
		if err := ruleGenerator.ClearRuleFiles(); err != nil {
			return Result{}, fmt.Errorf("clearing rule files: %w", err)
		}
	}

	// Create an empty context
	emptyCtx := i3context.Context{}

	// Save the empty context
	if err := i3context.SaveContext(cfg.I3Dir, emptyCtx); err != nil {
		return Result{}, fmt.Errorf("clearing context: %w", err)
	}

	return NewResult("Cleared active feature context and removed rule files.", nil, nil), nil
} 