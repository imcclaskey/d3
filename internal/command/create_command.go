package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/rules"
	"github.com/imcclaskey/i3/internal/workspace"
	"github.com/imcclaskey/i3/internal/validation"
)

// Create implements a feature creation command
type Create struct {
	Feature string // Name of the feature to create
	// Dependencies injected via interfaces/values
	featureCreator workspace.FeatureCreator
	// phaseMover is removed
}

// NewCreate creates a new feature creation command with default dependencies.
func NewCreate(feature string) Create {
	// Use NewCreateWithDeps internally for consistency
	return NewCreateWithDeps(feature, workspace.NewDefaultCreator())
}

// NewCreateWithDeps creates a new feature creation command with injected dependencies.
func NewCreateWithDeps(feature string, fc workspace.FeatureCreator) Create {
	return Create{
		Feature:        feature,
		featureCreator: fc,
	}
}

// Run implements the Command interface
func (c Create) Run(ctx context.Context, cfg Config) (Result, error) {
	// Validate i3 is initialized (this check might be less critical now)
	if err := validation.Init(cfg.I3Dir); err != nil {
		return Result{}, err
	}

	// Validate feature name
	if c.Feature == "" {
		return Result{}, fmt.Errorf("feature name is required (suggestion: provide a name for the feature to create)")
	}

	// Check if the feature already exists
	featureDir := filepath.Join(cfg.FeaturesDir, c.Feature)
	if _, err := os.Stat(featureDir); err == nil {
		return Result{}, fmt.Errorf("feature '%s' already exists (suggestion: choose a different name or use 'i3 enter %s' to use it)", c.Feature, c.Feature)
	} else if !os.IsNotExist(err) {
		// Handle other stat errors
		return Result{}, fmt.Errorf("checking feature directory %s: %w", featureDir, err)
	}

	// Create feature directory
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return Result{}, fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Ensure feature files using the injected creator
	if err := c.featureCreator.EnsureFeatureFiles(featureDir); err != nil {
		// Attempt cleanup if file creation fails
		_ = os.RemoveAll(featureDir)
		return Result{}, fmt.Errorf("failed to ensure feature files: %w", err)
	}

	// Set the context to the newly created feature and 'ideation' phase
	newCtx := i3context.Context{
		Feature: c.Feature,
		Phase:   "ideation", // Default starting phase
	}
	if err := i3context.SaveContext(cfg.I3Dir, newCtx); err != nil {
		// Attempt cleanup? Maybe not, context file fail is less critical than feature files.
		// Log warning? For now, return error.
		return Result{}, fmt.Errorf("setting context after creation: %w", err)
	}
	
	// Generate rule files
	ruleGenerator := rules.NewRuleFileGenerator("", cfg.CursorRulesDir)
	if err := ruleGenerator.EnsureRuleFiles(newCtx.Phase, newCtx.Feature); err != nil {
		return Result{}, fmt.Errorf("ensuring rule files: %w", err)
	}

	message := fmt.Sprintf("Created feature '%s' and entered context (phase: ideation)", c.Feature)
	return NewResult(message, newCtx, nil), nil
} 