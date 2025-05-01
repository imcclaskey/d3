package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/rules"
)

// Enter sets the current working feature context.
type Enter struct {
	Feature string // Name of the feature to enter
}

// NewEnter creates a new Enter command
func NewEnter(feature string) Enter {
	return Enter{Feature: feature}
}

// Run implements the Command interface
func (e Enter) Run(ctx context.Context, cfg Config) (Result, error) {
	// Validate feature name
	if e.Feature == "" {
		return Result{}, fmt.Errorf("feature name is required")
	}

	// Validate feature exists
	featureDir := filepath.Join(cfg.FeaturesDir, e.Feature)
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return Result{}, fmt.Errorf("feature '%s' does not exist", e.Feature)
	} else if err != nil {
		return Result{}, fmt.Errorf("checking feature '%s': %w", e.Feature, err)
	}

	// Create new context - set default phase to ideation
	newCtx := i3context.Context{
		Feature: e.Feature,
		Phase:   "ideation", // Default to ideation phase
	}

	// Save the context
	if err := i3context.SaveContext(cfg.I3Dir, newCtx); err != nil {
		return Result{}, fmt.Errorf("saving context: %w", err)
	}

	// Ensure rule files are generated
	ruleGenerator := rules.NewRuleFileGenerator("", cfg.CursorRulesDir)
	if err := ruleGenerator.EnsureRuleFiles(newCtx.Phase, newCtx.Feature); err != nil {
		return Result{}, fmt.Errorf("ensuring rule files: %w", err)
	}

	message := fmt.Sprintf("Entered feature context: %s (phase: %s)", e.Feature, newCtx.Phase)
	return NewResult(message, newCtx, nil), nil
} 