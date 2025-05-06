// Package feature implements core feature operations
package feature

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/ports"
)

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

// CreateFeature creates a new feature directory
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

	return &FeatureInfo{
		Name: featureName,
		Path: featurePath,
	}, nil
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
