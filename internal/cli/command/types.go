// Package command implements CLI commands for d3
package command

import (
	"context"
	"path/filepath"
)

// Config holds common configuration used by all commands
type Config struct {
	// WorkspaceRoot is the absolute path to the workspace root
	WorkspaceRoot string
	// D3Dir is the path to the .d3 directory
	D3Dir string
	// FeaturesDir is the path to the features directory
	FeaturesDir string
	// CursorRulesDir is the path to the cursor rules directory
	CursorRulesDir string
}

// NewConfig creates a configuration with all needed dependencies
func NewConfig(workspaceRoot string) Config {
	d3Dir := filepath.Join(workspaceRoot, ".d3")
	featuresDir := filepath.Join(d3Dir, "features")
	cursorRulesDir := filepath.Join(workspaceRoot, ".cursor", "rules")

	return Config{
		WorkspaceRoot:  workspaceRoot,
		D3Dir:          d3Dir,
		FeaturesDir:    featuresDir,
		CursorRulesDir: cursorRulesDir,
	}
}

// Command defines the interface for all d3 commands
type Command interface {
	// Run executes the command and returns a result or an error
	Run(ctx context.Context, cfg Config) (Result, error)
}

// Result holds the result of a command execution
type Result struct {
	Message  string
	Data     interface{}
	Warnings []string
}

// NewResult creates a new result with message, data and warnings
func NewResult(message string, data interface{}, warnings []string) Result {
	if warnings == nil {
		warnings = []string{}
	}

	return Result{
		Message:  message,
		Data:     data,
		Warnings: warnings,
	}
}
