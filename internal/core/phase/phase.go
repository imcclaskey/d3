// Package phase implements core phase operations
package phase

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/ports"
)

// Phase represents a d3 phase (e.g., define, design, deliver).
type Phase string

const (
	Define  Phase = "define"
	Design  Phase = "design"
	Deliver Phase = "deliver"
)

// PhaseFileMap defines the standard file associated with each phase.
var PhaseFileMap = map[Phase]string{
	Define:  "problem.md",
	Design:  "plan.md",
	Deliver: "progress.yaml",
}

// Service provides phase management operations
type Service struct {
	fs ports.FileSystem
}

// NewService creates a new phase service
func NewService(fs ports.FileSystem) *Service {
	return &Service{
		fs: fs,
	}
}

// EnsurePhaseFiles creates the necessary directories and placeholder files for all standard phases
// within the given feature's directory. It ensures the directories exist and creates empty
// files if they are missing.
func (s *Service) EnsurePhaseFiles(featureRoot string) error {
	// Process phases in a consistent order
	orderedPhases := []Phase{Define, Design, Deliver}

	for _, p := range orderedPhases {
		phaseDir := filepath.Join(featureRoot, string(p))
		filename := PhaseFileMap[p]
		filePath := filepath.Join(phaseDir, filename)

		// Ensure the phase directory exists
		if err := s.fs.MkdirAll(phaseDir, 0755); err != nil {
			return fmt.Errorf("failed to create phase directory %s: %w", phaseDir, err)
		}

		// Check if the file exists
		_, err := s.fs.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// Create the file if it doesn't exist
				file, err := s.fs.Create(filePath)
				if err != nil {
					return fmt.Errorf("failed to create phase file %s: %w", filePath, err)
				}
				file.Close()
			} else {
				return fmt.Errorf("failed to check phase file %s: %w", filePath, err)
			}
		}
		// File exists, nothing to do
	}

	return nil
}

// EnsurePhaseFiles creates the necessary directories and placeholder files for all standard phases
// within the given feature's directory. It ensures the directories exist and creates empty
// files if they are missing.
func EnsurePhaseFiles(fs ports.FileSystem, featureRoot string) error {
	// Process phases in a consistent order
	orderedPhases := []Phase{Define, Design, Deliver}

	for _, p := range orderedPhases {
		phaseDir := filepath.Join(featureRoot, string(p))
		filename := PhaseFileMap[p]
		filePath := filepath.Join(phaseDir, filename)

		// Ensure the phase directory exists
		if err := fs.MkdirAll(phaseDir, 0755); err != nil {
			return fmt.Errorf("failed to create phase directory %s: %w", phaseDir, err)
		}

		// Check if the file exists
		_, err := fs.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// Create the file if it doesn't exist
				file, err := fs.Create(filePath)
				if err != nil {
					return fmt.Errorf("failed to create phase file %s: %w", filePath, err)
				}
				file.Close()
			} else {
				return fmt.Errorf("failed to check phase file %s: %w", filePath, err)
			}
		}
		// File exists, nothing to do
	}

	return nil
}
