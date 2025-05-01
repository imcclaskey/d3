// Package workspace provides functions to ensure the i3 workspace structure
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// FeatureCreator defines the interface for ensuring feature files exist.
type FeatureCreator interface {
	// EnsureFeatureFiles creates the standard set of files for a new feature
	// within the specified feature directory.
	// It assumes the feature directory itself already exists.
	EnsureFeatureFiles(featureDir string) error
}

// defaultCreator implements the FeatureCreator interface using standard os calls.
type defaultCreator struct{}

// NewDefaultCreator creates a new instance of the default FeatureCreator.
func NewDefaultCreator() FeatureCreator {
	return &defaultCreator{}
}

// EnsureFeatureFiles implements the FeatureCreator interface's method.
func (c *defaultCreator) EnsureFeatureFiles(featureDir string) error {
	// Create ideation file (empty)
	ideationMD := filepath.Join(featureDir, "ideation.md")
	if err := os.WriteFile(ideationMD, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filepath.Base(ideationMD), err)
	}

	// Create instruction file (empty)
	instructionMD := filepath.Join(featureDir, "instruction.md")
	if err := os.WriteFile(instructionMD, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filepath.Base(instructionMD), err)
	}

	// Create implementation JSON file (with empty structure)
	implJSON := filepath.Join(featureDir, "implementation.json")
	implContent := `{ 
  "files": [],
  "tasks": []
}`

	if err := os.WriteFile(implJSON, []byte(implContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filepath.Base(implJSON), err)
	}

	return nil
}

// CleanWorkspace removes all i3 related files and directories for a clean initialization
func CleanWorkspace(i3Dir, cursorRulesDir string) error {
	// Remove i3 directory
	if err := removeIfExists(i3Dir); err != nil {
		return fmt.Errorf("cleaning i3 directory: %w", err)
	}
	
	// Remove cursor rules i3 directory
	cursorI3Dir := filepath.Join(cursorRulesDir, "i3") 
	if err := removeIfExists(cursorI3Dir); err != nil {
		return fmt.Errorf("cleaning cursor rules i3 directory: %w", err)
	}
	
	return nil
}

// removeIfExists removes a file or directory if it exists
func removeIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Path doesn't exist, nothing to do
		return nil
	} else if err != nil {
		// Error checking path
		return fmt.Errorf("checking %s: %w", path, err)
	}
	
	// Path exists, remove it
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("removing %s: %w", path, err)
	}
	
	return nil
}

// EnsureDirectories creates essential directories if they don't exist
func EnsureDirectories(i3Dir string) error {
	featuresDir := filepath.Join(i3Dir, "features")
	
	// Only create the base i3 directory and features directory
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	return nil
}

// EnsureBasicFiles creates the core files in the i3 directory
func EnsureBasicFiles(i3Dir string) error {
	// Create core files if they don't exist
	files := []string{
		filepath.Join(i3Dir, "project.md"),
		filepath.Join(i3Dir, "tech.md"),
	}
	
	for _, file := range files {
		// Create file only if it doesn't exist (idempotent)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if err := os.WriteFile(file, []byte(""), 0644); err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
		}
	}
	
	return nil
}

// EnsureGitignoreFiles creates .gitignore files in appropriate directories
func EnsureGitignoreFiles(i3Dir, cursorRulesDir string) error {
	// .gitignore content (same for both locations)
	gitignoreContent := `# i3 Files
context.json
/gen/*.mdc`
	
	// Create .gitignore in .i3 directory
	i3Gitignore := filepath.Join(i3Dir, ".gitignore")
	if err := createFileIfNotExists(i3Gitignore, gitignoreContent); err != nil {
		return fmt.Errorf("creating .gitignore in i3 directory: %w", err)
	}
	
	// Ensure the cursor rules i3 directory exists
	cursorI3Dir := filepath.Join(cursorRulesDir, "i3")
	if err := os.MkdirAll(cursorI3Dir, 0755); err != nil {
		return fmt.Errorf("creating cursor rules i3 directory: %w", err)
	}
	
	// Create .gitignore in .cursor/rules/i3 directory
	cursorGitignore := filepath.Join(cursorI3Dir, ".gitignore")
	if err := createFileIfNotExists(cursorGitignore, gitignoreContent); err != nil {
		return fmt.Errorf("creating .gitignore in cursor rules directory: %w", err)
	}
	
	return nil
}

// createFileIfNotExists creates a file with given content if it doesn't exist
func createFileIfNotExists(filePath, content string) error {
	// Check if file exists - only create if missing (idempotent)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Ensure directory exists
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
		}
		
		// Create or overwrite the file
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write to %s: %w", filePath, err)
		}
	}
	
	return nil
}

// EnsurePhaseFiles creates the file(s) for a specific phase
func EnsurePhaseFiles(featuresDir, feature, phase string) error {
	featureDir := filepath.Join(featuresDir, feature)
	
	// Create feature directory
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		return fmt.Errorf("failed to create feature directory: %w", err)
	}
	
	// Define files needed for each phase
	var files []string
	
	switch phase {
	case "setup":
		// Setup doesn't need any feature files
		return nil
	case "ideation":
		files = []string{"ideation.md"}
	case "instruction":
		files = []string{"instruction.md"}
	case "implementation":
		files = []string{"implementation.json"}
	}
	
	// Create the files for this phase if they don't exist
	for _, filename := range files {
		path := filepath.Join(featureDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(""), 0644); err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
		}
	}
	
	return nil
} 