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

// EnsurePhaseFiles creates the necessary directories and placeholder files for all standard phases
// within the given feature's directory. It ensures the directories exist and creates empty
// files if they are missing.
func EnsurePhaseFiles(fs ports.FileSystem, featureRoot string) error {
	// Process phases in a consistent order
	orderedPhases := []Phase{Define, Design, Deliver}

	for _, phase := range orderedPhases {
		filename, exists := PhaseFileMap[phase]
		if !exists {
			continue // Skip if somehow the phase is not in the map
		}

		phaseDir := filepath.Join(featureRoot, string(phase))
		filePath := filepath.Join(phaseDir, filename)

		// Create the phase directory if it doesn't exist.
		if err := fs.MkdirAll(phaseDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", phaseDir, err)
		}

		// Create the phase file only if it doesn't exist.
		_, err := fs.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				file, createErr := fs.Create(filePath)
				if createErr != nil {
					return fmt.Errorf("failed to create file %s: %w", filePath, createErr)
				}
				// Close the file immediately after creation to release the handle.
				if closeErr := file.Close(); closeErr != nil {
					// Log or handle the error if closing fails, though it's less critical than creation failure.
					fmt.Fprintf(os.Stderr, "warning: failed to close file %s: %v\n", filePath, closeErr)
				}
			} else {
				// Handle other potential errors from os.Stat (e.g., permission issues).
				return fmt.Errorf("failed to check file status %s: %w", filePath, err)
			}
		}
	}
	return nil
}
