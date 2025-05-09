// Package feature implements core feature operations
package feature

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -package=mocks -destination=mocks/feature_mock.go . FeatureServicer

// FeatureServicer defines the interface for feature operations.
// This allows for mocking the feature service in tests.
type FeatureServicer interface {
	CreateFeature(ctx context.Context, featureName string) (*FeatureInfo, error)
	GetFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error)
	SetFeaturePhase(ctx context.Context, featureName string, p phase.Phase) error
	FeatureExists(featureName string) bool
	GetFeaturePath(featureName string) string
	ListFeatures(ctx context.Context) ([]FeatureInfo, error)
	DeleteFeature(ctx context.Context, featureName string) (activeContextCleared bool, err error)
	GetActiveFeature() (string, error)
	SetActiveFeature(featureName string) error
	ClearActiveFeature() error
}

// featureStateData defines the structure for a feature's state.yaml file
type featureStateData struct {
	LastActivePhase phase.Phase `yaml:"active_phase"`
}

const stateFileName = "state.yaml"

// Service provides feature management operations
type Service struct {
	projectRoot           string
	featuresDir           string
	d3Dir                 string
	activeFeatureFilePath string
	fs                    ports.FileSystem
}

// NewService creates a new feature service
func NewService(projectRoot, featuresDir, d3Dir string, fs ports.FileSystem) *Service {
	return &Service{
		projectRoot:           projectRoot,
		featuresDir:           featuresDir,
		d3Dir:                 d3Dir,
		activeFeatureFilePath: filepath.Join(d3Dir, ".active_feature"),
		fs:                    fs,
	}
}

// FeatureInfo contains basic information about a feature
type FeatureInfo struct {
	Name string
	Path string
}

// CreateFeature creates a new feature directory and its initial state.yaml file
func (s *Service) CreateFeature(ctx context.Context, featureName string) (*FeatureInfo, error) {
	featurePath := filepath.Join(s.featuresDir, featureName)

	// Check if feature already exists
	if _, err := s.fs.Stat(featurePath); err == nil {
		return nil, fmt.Errorf("feature %s already exists", featureName)
	} else if !os.IsNotExist(err) {
		// If it's an error other than NotExist, return it
		return nil, fmt.Errorf("failed to check if feature %s exists: %w", featureName, err)
	}

	// Create feature directory
	if err := s.fs.MkdirAll(featurePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Create initial state.yaml for the feature
	initialState := featureStateData{LastActivePhase: phase.Define} // Default to Define phase
	stateFilePath := filepath.Join(featurePath, stateFileName)
	data, err := yaml.Marshal(&initialState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal initial feature state for %s: %w", featureName, err)
	}
	if err := s.fs.WriteFile(stateFilePath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write initial state.yaml for feature %s: %w", featureName, err)
	}

	return &FeatureInfo{
		Name: featureName,
		Path: featurePath,
	}, nil
}

// GetFeaturePhase reads the last active phase from a feature's state.yaml file.
// If state.yaml doesn't exist, it creates it with a default phase (Define) and returns that.
func (s *Service) GetFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error) {
	featurePath := filepath.Join(s.featuresDir, featureName)
	stateFilePath := filepath.Join(featurePath, stateFileName)

	data, err := s.fs.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// state.yaml does not exist, create it with default Define phase
			initialState := featureStateData{LastActivePhase: phase.Define}
			writeData, marshalErr := yaml.Marshal(&initialState)
			if marshalErr != nil {
				return phase.None, fmt.Errorf("failed to marshal default state for %s: %w", featureName, marshalErr)
			}
			// Ensure feature directory exists before writing state file (important for bare features)
			if errMkdir := s.fs.MkdirAll(featurePath, 0755); errMkdir != nil {
				return phase.None, fmt.Errorf("failed to create directory for feature %s to write state.yaml: %w", featureName, errMkdir)
			}
			if writeErr := s.fs.WriteFile(stateFilePath, writeData, 0644); writeErr != nil {
				return phase.None, fmt.Errorf("failed to write default state.yaml for %s: %w", featureName, writeErr)
			}
			return phase.Define, nil // Return default phase after creation
		}
		// Other error reading file
		return phase.None, fmt.Errorf("failed to read state.yaml for feature %s: %w", featureName, err)
	}

	var state featureStateData
	if err := yaml.Unmarshal(data, &state); err != nil {
		return phase.None, fmt.Errorf("failed to unmarshal state.yaml for feature %s: %w", featureName, err)
	}

	if state.LastActivePhase == "" {
		// If phase is empty string after unmarshal, treat as Define or an error depending on strictness
		// For now, let's default to Define if it's empty for some reason.
		return phase.Define, nil
	}

	return state.LastActivePhase, nil
}

// SetFeaturePhase writes the given phase to a feature's state.yaml file.
func (s *Service) SetFeaturePhase(ctx context.Context, featureName string, p phase.Phase) error {
	featurePath := filepath.Join(s.featuresDir, featureName)
	stateFilePath := filepath.Join(featurePath, stateFileName)

	newState := featureStateData{LastActivePhase: p}
	data, err := yaml.Marshal(&newState)
	if err != nil {
		return fmt.Errorf("failed to marshal feature state for %s: %w", featureName, err)
	}

	// Ensure feature directory exists before writing state file (important for bare features)
	if errMkdir := s.fs.MkdirAll(featurePath, 0755); errMkdir != nil {
		return fmt.Errorf("failed to create directory for feature %s to write state.yaml: %w", featureName, errMkdir)
	}

	if err := s.fs.WriteFile(stateFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state.yaml for feature %s: %w", featureName, err)
	}
	return nil
}

// FeatureExists checks if a feature exists
func (s *Service) FeatureExists(featureName string) bool {
	featurePath := filepath.Join(s.featuresDir, featureName)
	_, err := s.fs.Stat(featurePath)
	return err == nil
}

// GetFeaturePath returns the path to a feature directory
func (s *Service) GetFeaturePath(featureName string) string {
	return filepath.Join(s.featuresDir, featureName)
}

// ListFeatures returns a list of all features
func (s *Service) ListFeatures(ctx context.Context) ([]FeatureInfo, error) {
	// Check if features directory exists
	if _, err := s.fs.Stat(s.featuresDir); os.IsNotExist(err) {
		return []FeatureInfo{}, nil // Empty list, not an error
	} else if err != nil {
		return nil, fmt.Errorf("failed to check features directory: %w", err)
	}

	// Read feature directories
	entries, err := s.fs.ReadDir(s.featuresDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read features directory: %w", err)
	}

	// Build result list
	var features []FeatureInfo
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, FeatureInfo{
				Name: entry.Name(),
				Path: filepath.Join(s.featuresDir, entry.Name()),
			})
		}
	}

	return features, nil
}

// GetActiveFeature reads the active feature name from the active feature file.
// Returns an empty string and nil error if the file is empty or does not exist.
func (s *Service) GetActiveFeature() (string, error) {
	data, err := s.fs.ReadFile(s.activeFeatureFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File not existing means no active feature
		}
		return "", fmt.Errorf("failed to read active feature file %s: %w", s.activeFeatureFilePath, err)
	}
	// Return the content, trimming whitespace
	return strings.TrimSpace(string(data)), nil
}

// SetActiveFeature saves the active feature name to the active feature file.
func (s *Service) SetActiveFeature(featureName string) error {
	// Ensure the base directory exists
	if err := s.fs.MkdirAll(filepath.Dir(s.activeFeatureFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create active feature file directory: %w", err)
	}

	// Write the feature name as plain text
	data := []byte(featureName)
	if err := s.fs.WriteFile(s.activeFeatureFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write active feature file %s: %w", s.activeFeatureFilePath, err)
	}
	return nil
}

// ClearActiveFeature removes the active feature file, effectively clearing the active feature.
func (s *Service) ClearActiveFeature() error {
	err := s.fs.Remove(s.activeFeatureFilePath)
	// Ignore "not exist" error, as it means the state is already cleared
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove active feature file %s: %w", s.activeFeatureFilePath, err)
	}
	return nil
}

// DeleteFeature removes a feature directory and its contents.
// If the deleted feature is the currently active one, it also clears the active feature state.
// Returns true if the active context was cleared as a result of this deletion.
func (s *Service) DeleteFeature(ctx context.Context, featureName string) (bool, error) {
	activeContextCleared := false
	featurePath := filepath.Join(s.featuresDir, featureName)

	currentActiveFeature, err := s.GetActiveFeature()
	if err != nil {
		// Log a warning, but this might not be fatal for the deletion itself if the active feature file is just unreadable.
		// However, we need to know if featureName *was* the active one.
		fmt.Fprintf(os.Stderr, "Warning: could not read active feature state while trying to delete %s: %v\n", featureName, err)
		// If we can't read it, we can't be sure it's not featureName. Proceeding cautiously by not assuming it *isn't* featureName.
	}

	if currentActiveFeature == featureName {
		if clearErr := s.ClearActiveFeature(); clearErr != nil {
			// If we can't clear the active feature, we should probably stop and report this.
			return false, fmt.Errorf("failed to clear active feature state for %s before deletion: %w", featureName, clearErr)
		}
		activeContextCleared = true
	}

	// Check if feature directory exists before attempting to remove
	if _, err := s.fs.Stat(featurePath); os.IsNotExist(err) {
		return activeContextCleared, fmt.Errorf("feature '%s' not found at %s", featureName, featurePath)
	} else if err != nil {
		return activeContextCleared, fmt.Errorf("failed to check feature '%s': %w", featureName, err)
	}

	// Remove the entire feature directory
	if errRemove := s.fs.RemoveAll(featurePath); errRemove != nil {
		return activeContextCleared, fmt.Errorf("failed to delete feature '%s': %w", featureName, errRemove)
	}

	return activeContextCleared, nil
}
