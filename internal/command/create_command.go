package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/imcclaskey/i3/internal/errors"
	"github.com/imcclaskey/i3/internal/validation"
)

// Create implements a feature creation command
type Create struct {
	Feature string // Name of the feature to create
}

// NewCreate creates a new feature creation command
func NewCreate(feature string) Create {
	return Create{
		Feature: feature,
	}
}

// Run implements the Command interface
func (c Create) Run(ctx context.Context, cfg Config) (string, error) {
	// Validate i3 is initialized
	if err := validation.Init(cfg.I3Dir); err != nil {
		return "", err
	}
	
	// Validate feature name
	if c.Feature == "" {
		return "", errors.WithSuggestion(
			errors.New("feature name is required"),
			"provide a name for the feature to create",
		)
	}
	
	// Check if the feature already exists
	featureDir := filepath.Join(cfg.FeaturesDir, c.Feature)
	if _, err := os.Stat(featureDir); err == nil {
		return "", errors.WithSuggestion(
			errors.Errorf("feature '%s' already exists", c.Feature),
			"choose a different name or use 'i3 move' to enter it",
		)
	}
	
	// Create feature directory
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create feature directory")
	}
	
	// Create feature files
	if err := createFeatureFiles(featureDir); err != nil {
		return "", errors.Wrap(err, "failed to create feature files")
	}
	
	// Always enter ideation phase
	move := NewMove("ideation", c.Feature, true)
	if _, err := move.Run(ctx, cfg); err != nil {
		return "", errors.Wrap(err, "failed to enter ideation phase")
	}
	
	return fmt.Sprintf("Created feature '%s' and entered ideation phase", c.Feature), nil
}

// createFeatureFiles creates the necessary files for a new feature
func createFeatureFiles(featureDir string) error {
	// Create ideation file (empty)
	ideationMD := filepath.Join(featureDir, "ideation.md")
	if err := os.WriteFile(ideationMD, []byte(""), 0644); err != nil {
		return err
	}
	
	// Create instruction file (empty)
	instructionMD := filepath.Join(featureDir, "instruction.md")
	if err := os.WriteFile(instructionMD, []byte(""), 0644); err != nil {
		return err
	}
	
	// Create implementation JSON file (with empty structure)
	implJSON := filepath.Join(featureDir, "implementation.json")
	implContent := `{
  "files": [],
  "tasks": []
}`

	if err := os.WriteFile(implJSON, []byte(implContent), 0644); err != nil {
		return err
	}
	
	return nil
} 