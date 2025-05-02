// Package files provides file management for i3 project files
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/imcclaskey/i3/internal/common"
	"github.com/imcclaskey/i3/internal/core/rules"
	"gopkg.in/yaml.v3"
)

// Paths contains the standard paths for an i3 project
type Paths struct {
	WorkspaceRoot  string
	I3Dir          string
	FeaturesDir    string
	CursorRulesDir string
}

// GetPaths returns the standard paths for an i3 project
func GetPaths() (Paths, error) {
	workspaceRoot, err := common.GetWorkspaceRoot()
	if err != nil {
		return Paths{}, fmt.Errorf("failed to get workspace root: %w", err)
	}

	i3Dir := filepath.Join(workspaceRoot, ".i3")
	featuresDir := filepath.Join(i3Dir, "features")
	cursorRulesDir := filepath.Join(workspaceRoot, ".cursor", "rules")

	return Paths{
		WorkspaceRoot:  workspaceRoot,
		I3Dir:          i3Dir,
		FeaturesDir:    featuresDir,
		CursorRulesDir: cursorRulesDir,
	}, nil
}

// Service provides file management operations
type Service struct {
	paths Paths
}

// NewService creates a new file management service
func NewService(workspaceRoot string) *Service {
	// Construct paths directly using the provided workspaceRoot
	i3Dir := filepath.Join(workspaceRoot, ".i3")
	featuresDir := filepath.Join(i3Dir, "features")
	cursorRulesDir := filepath.Join(workspaceRoot, ".cursor", "rules")

	paths := Paths{
		WorkspaceRoot:  workspaceRoot,
		I3Dir:          i3Dir,
		FeaturesDir:    featuresDir,
		CursorRulesDir: cursorRulesDir,
	}

	return &Service{
		paths: paths,
	}
}

// GetWorkspaceRoot returns the configured workspace root path for the service.
func (s *Service) GetWorkspaceRoot() (string, error) {
	if s == nil {
		return "", fmt.Errorf("files.Service is nil")
	}
	// We could potentially re-verify with common.GetWorkspaceRoot() here,
	// but for diagnostics, returning the path the service *was initialized with* is more useful.
	return s.paths.WorkspaceRoot, nil
}

// GetI3DirPath returns the calculated .i3 directory path for the service.
func (s *Service) GetI3DirPath() string {
	if s == nil {
		return ""
	}
	return s.paths.I3Dir
}

// GetFeaturesPath returns the calculated features directory path for the service.
func (s *Service) GetFeaturesPath() string {
	if s == nil {
		return ""
	}
	return s.paths.FeaturesDir
}

// InitWorkspace initializes the i3 workspace
// Returns a message describing the operation and any newly created files
func (s *Service) InitWorkspace(clean bool) (string, []string, error) {
	// If clean, remove existing i3 directory
	if clean {
		if err := os.RemoveAll(s.paths.I3Dir); err != nil {
			return "", nil, fmt.Errorf("failed to clean i3 directory: %w", err)
		}
	}

	// Create directory structure
	if err := os.MkdirAll(s.paths.FeaturesDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create directories: %w", err)
	}

	// Create basic files
	newlyCreated, err := s.EnsureProjectFiles(clean)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create project files: %w", err)
	}

	// Create empty session file
	sessionFile := filepath.Join(s.paths.I3Dir, "session.yaml")
	sessionData := struct {
		Version      string    `yaml:"version"`
		LastModified time.Time `yaml:"last_modified"`
	}{
		Version:      "1.0",
		LastModified: time.Now(),
	}

	// Serialize to YAML
	sessionYaml, err := yaml.Marshal(sessionData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session data: %w", err)
	}

	// Write session file
	if err := os.WriteFile(sessionFile, sessionYaml, 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write session file: %w", err)
	}

	// Generate core rule file
	if err := s.GenerateCoreRuleFile("", ""); err != nil {
		return "", nil, fmt.Errorf("failed to create core rule file: %w", err)
	}

	// Prepare result message
	message := "Initialized i3 in current workspace"
	if clean {
		message = "Clean initialized i3 in current workspace"
	}

	return message, newlyCreated, nil
}

// GenerateCoreRuleFile generates the core rule file with the given feature and phase
// The rule generator handles empty feature/phase strings by using "Ready" as the prefix
func (s *Service) GenerateCoreRuleFile(feature, phase string) error {
	// Create rule generator
	ruleGen := rules.NewRuleGenerator()

	// Generate core rule content
	coreContent, err := ruleGen.GenerateCoreContent(feature, phase)
	if err != nil {
		return fmt.Errorf("failed to generate core rule: %w", err)
	}

	// Ensure i3 directory exists
	i3Dir := filepath.Join(s.paths.CursorRulesDir, "i3")
	if err := os.MkdirAll(i3Dir, 0755); err != nil {
		return fmt.Errorf("failed to create rule directory: %w", err)
	}

	// Write core rule file
	corePath := filepath.Join(i3Dir, "core.gen.mdc")
	if err := os.WriteFile(corePath, []byte(coreContent), 0644); err != nil {
		return fmt.Errorf("failed to write core rule file: %w", err)
	}

	return nil
}

// EnsureProjectFiles ensures the base project files exist in the i3 root
// Returns a list of newly created files
func (s *Service) EnsureProjectFiles(clean bool) ([]string, error) {
	var newlyCreated []string
	projectFiles := []string{"project.md", "tech.md"}

	// Create each base file
	for _, filename := range projectFiles {
		filePath := filepath.Join(s.paths.I3Dir, filename)

		// Skip if file exists and we're not doing a clean init
		if !clean {
			if _, err := os.Stat(filePath); err == nil {
				continue
			}
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}

		// Write empty file
		if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filename, err)
		}

		newlyCreated = append(newlyCreated, filename)
	}

	return newlyCreated, nil
}

// EnsurePhaseFiles creates phase-specific files for a feature
// Returns a list of newly created files
func (s *Service) EnsurePhaseFiles(feature, phase string) ([]string, error) {
	if feature == "" {
		return nil, fmt.Errorf("feature name must be specified")
	}

	// Map phase to its single file
	var filename string
	switch phase {
	case "ideation":
		filename = "problem.md"
	case "instruction":
		filename = "plan.md"
	case "implementation":
		filename = "progress.yaml"
	default:
		return nil, nil // No file for this phase
	}

	// Create phase directory
	phaseDir := filepath.Join(s.paths.FeaturesDir, feature, phase)
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create phase directory: %w", err)
	}

	var newlyCreated []string
	filePath := filepath.Join(phaseDir, filename)

	// Skip if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return nil, nil
	}

	// Write empty file
	if err := os.WriteFile(filePath, []byte{}, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", filename, err)
	}

	newlyCreated = append(newlyCreated, filename)
	return newlyCreated, nil
}

// HasPhaseFiles checks if a phase directory has files
func (s *Service) HasPhaseFiles(feature, phase string) bool {
	if feature == "" || phase == "" {
		return false
	}

	// Check for phase directory
	phaseDir := filepath.Join(s.paths.FeaturesDir, feature, phase)
	if _, err := os.Stat(phaseDir); os.IsNotExist(err) {
		return false
	}

	// Check if directory has any files
	entries, err := os.ReadDir(phaseDir)
	if err != nil || len(entries) == 0 {
		return false
	}

	return true
}
