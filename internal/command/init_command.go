package command

import (
	"context"
	"fmt"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/workspace"
	"github.com/imcclaskey/i3/internal/validation"
)

// Init implements the i3 initialization command
type Init struct {
	Clean bool // Whether to perform a clean initialization (remove existing files)
}

// NewInit creates a new initialization command
func NewInit(clean bool) Init {
	return Init{Clean: clean}
}

// Run implements the Command interface
func (i Init) Run(ctx context.Context, cfg Config) (Result, error) {
	// If clean flag is set, remove existing i3 directories/files
	if i.Clean {
		if err := workspace.CleanWorkspace(cfg.I3Dir, cfg.CursorRulesDir); err != nil {
			return Result{}, fmt.Errorf("cleaning workspace: %w", err)
		}
	} else {
		// Only check initialization status when not performing a clean init
		initErr := validation.Init(cfg.I3Dir)
		if initErr == nil {
			// Already initialized and not cleaning
			return Result{},
				fmt.Errorf("i3 already initialized in %s (use --clean to reinitialize)", cfg.I3Dir)
		}
	}

	// Ensure base directories exist
	if err := workspace.EnsureDirectories(cfg.I3Dir); err != nil {
		return Result{}, fmt.Errorf("failed to ensure base directories: %w", err)
	}

	// Ensure basic project files exist
	if err := workspace.EnsureBasicFiles(cfg.I3Dir); err != nil {
		return Result{}, fmt.Errorf("failed to ensure basic files: %w", err)
	}
	
	// Ensure gitignore files exist
	if err := workspace.EnsureGitignoreFiles(cfg.I3Dir, cfg.CursorRulesDir); err != nil {
		return Result{}, fmt.Errorf("failed to ensure gitignore files: %w", err)
	}

	// Clear any existing context upon initialization
	emptyCtx := i3context.Context{}
	if err := i3context.SaveContext(cfg.I3Dir, emptyCtx); err != nil {
		return Result{}, fmt.Errorf("clearing context during init: %w", err)
	}

	// Success message with note about clean initialization
	var message string
	if i.Clean {
		message = fmt.Sprintf("i3 initialized with clean workspace in %s", cfg.I3Dir)
	} else {
		message = fmt.Sprintf("i3 initialized successfully in %s", cfg.I3Dir)
	}
	
	return NewResult(message, nil, nil), nil
} 