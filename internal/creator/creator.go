// Package creator provides functions to create i3 resources
package creator

import (
	"fmt"
	"os"
	"path/filepath"
)

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