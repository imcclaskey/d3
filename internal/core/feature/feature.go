// Package feature implements core feature operations
package feature

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/ports"
	"github.com/imcclaskey/d3/internal/core/session"
	"gopkg.in/yaml.v3"
)

// featureStateData defines the structure for a feature's state.yaml file
type featureStateData struct {
	LastActivePhase session.Phase `yaml:"active_phase"`
}

const stateFileName = "state.yaml"

// Service provides feature management operations
type Service struct {
	projectRoot string
	featuresDir string
	d3Dir       string
	fs          ports.FileSystem
}

// NewService creates a new feature service
func NewService(projectRoot, featuresDir, d3Dir string, fs ports.FileSystem) *Service {
	return &Service{
		projectRoot: projectRoot,
		featuresDir: featuresDir,
		d3Dir:       d3Dir,
		fs:          fs,
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
	initialState := featureStateData{LastActivePhase: session.Define} // Default to Define phase
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
func (s *Service) GetFeaturePhase(ctx context.Context, featureName string) (session.Phase, error) {
	featurePath := filepath.Join(s.featuresDir, featureName)
	stateFilePath := filepath.Join(featurePath, stateFileName)

	data, err := s.fs.ReadFile(stateFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// state.yaml does not exist, create it with default Define phase
			initialState := featureStateData{LastActivePhase: session.Define}
			writeData, marshalErr := yaml.Marshal(&initialState)
			if marshalErr != nil {
				return session.None, fmt.Errorf("failed to marshal default state for %s: %w", featureName, marshalErr)
			}
			// Ensure feature directory exists before writing state file (important for bare features)
			if errMkdir := s.fs.MkdirAll(featurePath, 0755); errMkdir != nil {
				return session.None, fmt.Errorf("failed to create directory for feature %s to write state.yaml: %w", featureName, errMkdir)
			}
			if writeErr := s.fs.WriteFile(stateFilePath, writeData, 0644); writeErr != nil {
				return session.None, fmt.Errorf("failed to write default state.yaml for %s: %w", featureName, writeErr)
			}
			return session.Define, nil // Return default phase after creation
		}
		// Other error reading file
		return session.None, fmt.Errorf("failed to read state.yaml for feature %s: %w", featureName, err)
	}

	var state featureStateData
	if err := yaml.Unmarshal(data, &state); err != nil {
		return session.None, fmt.Errorf("failed to unmarshal state.yaml for feature %s: %w", featureName, err)
	}

	if state.LastActivePhase == "" {
		// If phase is empty string after unmarshal, treat as Define or an error depending on strictness
		// For now, let's default to Define if it's empty for some reason.
		return session.Define, nil
	}

	return state.LastActivePhase, nil
}

// SetFeaturePhase writes the given phase to a feature's state.yaml file.
func (s *Service) SetFeaturePhase(ctx context.Context, featureName string, phase session.Phase) error {
	featurePath := filepath.Join(s.featuresDir, featureName)
	stateFilePath := filepath.Join(featurePath, stateFileName)

	newState := featureStateData{LastActivePhase: phase}
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
