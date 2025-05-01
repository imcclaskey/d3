package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/imcclaskey/i3/internal/errors"
)

// Init implements an initialization command
type Init struct {
	Force bool // Whether to force re-initialization
}

// NewInit creates a new initialization command
func NewInit(force bool) Init {
	return Init{
		Force: force,
	}
}

// Run implements the Command interface
func (i Init) Run(ctx context.Context, cfg Config) (string, error) {
	// Check for Cursor integration
	cursorRulesDir := filepath.Join(cfg.WorkspaceRoot, ".cursor", "rules")
	if _, err := os.Stat(cursorRulesDir); os.IsNotExist(err) {
		return "", errors.WithSuggestion(errors.ErrCursorIntegration, 
			"This command must be run in a Cursor project")
	}
	
	// Check if i3 is already initialized
	isInitialized := false
	if _, err := os.Stat(cfg.I3Dir); err == nil {
		isInitialized = true
		
		// If already initialized and not force, return error
		if !i.Force {
			return "", errors.WithSuggestion(errors.ErrAlreadyInitialized,
				"Use --force to reinitialize (this will overwrite existing files)")
		}
	}
	
	// Create required directories
	directories := []string{
		cfg.I3Dir,
		cfg.FeaturesDir,
		filepath.Join(cursorRulesDir, "i3"),
	}
	
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed to create directory %s", dir))
		}
	}
	
	// Create required files
	if err := createRequiredFiles(cfg.I3Dir, i.Force); err != nil {
		return "", errors.Wrap(err, "failed to create required files")
	}
	
	// Create i3 .gitignore files
	if err := createGitignoreFiles(cfg.I3Dir, cursorRulesDir, i.Force); err != nil {
		return "", errors.Wrap(err, "failed to create .gitignore files")
	}
	
	// Initialize session
	if err := cfg.Session.Stop(); err != nil {
		return "", errors.Wrap(err, "failed to initialize session")
	}
	
	// Return success message with appropriate text
	message := "i3 has been initialized successfully"
	if isInitialized {
		message = "i3 has been reinitialized successfully"
	}
	
	return message, nil
}

// createRequiredFiles creates the required files with empty content
func createRequiredFiles(i3Dir string, force bool) error {
	// Create project.md (empty)
	projectMD := filepath.Join(i3Dir, "project.md")
	if err := createFileIfNotExists(projectMD, "", force); err != nil {
		return err
	}
	
	// Create tech.md (empty)
	techMD := filepath.Join(i3Dir, "tech.md")
	if err := createFileIfNotExists(techMD, "", force); err != nil {
		return err
	}
	
	return nil
}

// createGitignoreFiles creates .gitignore files in appropriate directories
func createGitignoreFiles(i3Dir, cursorRulesDir string, force bool) error {
	// .gitignore content (same for both locations)
	gitignoreContent := `# i3 Files
session.json
/gen/*.mdc`
	
	// Create .gitignore in .i3 directory
	i3Gitignore := filepath.Join(i3Dir, ".gitignore")
	if err := createFileIfNotExists(i3Gitignore, gitignoreContent, force); err != nil {
		return err
	}
	
	// Create .gitignore in .cursor/rules/i3 directory
	cursorGitignore := filepath.Join(cursorRulesDir, "i3", ".gitignore")
	if err := createFileIfNotExists(cursorGitignore, gitignoreContent, force); err != nil {
		return err
	}
	
	return nil
}

// createFileIfNotExists creates a file with given content if it doesn't exist or force is true
func createFileIfNotExists(filePath, content string, force bool) error {
	// Check if file exists
	if _, err := os.Stat(filePath); err == nil && !force {
		return errors.WithDetails(errors.ErrAlreadyInitialized, 
			fmt.Sprintf("file already exists: %s", filePath))
	}
	
	// Create or overwrite the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to write to %s", filePath))
	}
	
	return nil
} 