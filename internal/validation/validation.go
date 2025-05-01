package validation

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/imcclaskey/i3/internal/errors"
)

// Phase names that are valid in i3
var ValidPhases = []string{"setup", "ideation", "instruction", "implementation"}

// Init checks if i3 is properly initialized in the given directory
func Init(i3Dir string) error {
	paths := []string{
		i3Dir,
		filepath.Join(i3Dir, "features"),
		filepath.Join(i3Dir, "session.json"),
		filepath.Join(i3Dir, "project.md"),
		filepath.Join(i3Dir, "tech.md"),
	}
	
	// Check for existence of core directories and files
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			base := filepath.Base(path)
			suggestion := "run 'i3 init'"
			if base != filepath.Base(i3Dir) {
				suggestion = "run 'i3 init --force' to reinitialize"
			}
			return errors.WithSuggestion(errors.ErrNotInitialized, suggestion)
		}
	}
	
	return nil
}

// Phase validates if a phase name is valid
func Phase(name string) error {
	for _, valid := range ValidPhases {
		if name == valid {
			return nil
		}
	}
	
	validStr := strings.Join(ValidPhases, ", ")
	return errors.WithDetails(errors.ErrInvalidPhase, 
		"must be one of: "+validStr)
}

// Feature checks if a feature exists with all required files
func Feature(featuresDir, name string) error {
	featureDir := filepath.Join(featuresDir, name)
	if _, err := os.Stat(featureDir); os.IsNotExist(err) {
		return errors.WithSuggestion(errors.ErrFeatureNotFound, 
			"use 'i3 create "+name+"' to create it")
	}
	
	// Check for required files
	required := []string{"ideation.md", "instruction.md", "implementation.json"}
	var missing []string
	
	for _, file := range required {
		path := filepath.Join(featureDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, file)
		}
	}
	
	if len(missing) > 0 {
		return errors.WithSuggestion(errors.ErrMissingFile, 
			"missing files: "+strings.Join(missing, ", ")+
			". You may need to recreate the feature")
	}
	
	return nil
}

// ContentWarnings checks project files for potential issues
func ContentWarnings(i3Dir string) []string {
	var warnings []string
	
	// Check files for empty content
	fileChecks := map[string]string{
		filepath.Join(i3Dir, "project.md"): "Project description file is empty or missing content",
		filepath.Join(i3Dir, "tech.md"):    "Technical stack file is empty or missing content",
	}
	
	for file, warning := range fileChecks {
		empty, _ := isFileEmpty(file)
		if empty {
			warnings = append(warnings, warning)
		}
	}
	
	// Check for Cursor integration
	workspaceRoot := filepath.Dir(i3Dir)
	i3MdcFile := filepath.Join(workspaceRoot, ".cursor", "rules", "i3.mdc")
	if _, err := os.Stat(i3MdcFile); os.IsNotExist(err) {
		warnings = append(warnings, "Cursor integration file is missing")
	}
	
	return warnings
}

// isFileEmpty returns true if a file doesn't exist, is empty, or contains only whitespace
func isFileEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	defer f.Close()
	
	// Read a small amount to check if file is empty
	buf := make([]byte, 64)
	n, err := f.Read(buf)
	if err == io.EOF || n == 0 {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	
	// Check if content is only whitespace
	return strings.TrimSpace(string(buf[:n])) == "", nil
} 