// Package project provides project management functionality
package project

// Requirements defines an interface for project operation requirements
// to be used by CLI commands and MCP tools that need to check if a
// project is initialized
type Requirements interface {
	// IsInitialized checks if the project is initialized
	IsInitialized() bool

	// RequiresInitialized ensures the project is initialized
	// Returns ErrNotInitialized if project is not initialized
	RequiresInitialized() error
}
