package project

import (
	"context"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
	// "github.com/imcclaskey/d3/internal/core/session"
)

//go:generate mockgen -package=project -destination=interfaces_mock.go . FeatureServicer,RulesServicer,PhaseServicer,FileOperator

// FeatureServicer defines the interface for feature management operations.
type FeatureServicer interface {
	CreateFeature(ctx context.Context, featureName string) (*feature.FeatureInfo, error)
	GetFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error)
	SetFeaturePhase(ctx context.Context, featureName string, p phase.Phase) error
	FeatureExists(featureName string) bool
	GetFeaturePath(featureName string) string
	ListFeatures(ctx context.Context) ([]feature.FeatureInfo, error)
	DeleteFeature(ctx context.Context, featureName string) (activeContextCleared bool, err error)
	GetActiveFeature() (string, error)
	SetActiveFeature(featureName string) error
	ClearActiveFeature() error
}

// RulesServicer defines the interface for rule management operations.
type RulesServicer interface {
	RefreshRules(feature string, phaseStr string) error
	ClearGeneratedRules() error
	InitCustomRulesDir() error
}

// PhaseServicer defines the interface for phase management operations.
type PhaseServicer interface {
	EnsurePhaseFiles(featureDir string) error
}

// FileOperator defines operations for project file manipulations needed by ProjectService.
type FileOperator interface {
	EnsureMCPJSON(fs ports.FileSystem, projectRoot string) error

	// EnsureRootGitignoreEntries manages D3-specific entries in the root .gitignore file.
	// It preserves user entries and maintains the D3 section of the file.
	EnsureRootGitignoreEntries(fs ports.FileSystem, projectRootAbs string) error

	// EnsureRootCursorignoreEntries manages D3-specific entries in the root .cursorignore file.
	// It preserves user entries and maintains the D3 section of the file.
	EnsureRootCursorignoreEntries(fs ports.FileSystem, projectRootAbs string) error

	// EnsureIgnoreFileEntries is a generic function for managing entries in ignore files.
	// It preserves user entries and maintains a specific section of the file.
	EnsureIgnoreFileEntries(fs ports.FileSystem, ignoreFilePath string, patterns []string, sectionMarker string) error

	EnsureProjectFiles(fs ports.FileSystem, d3DirAbs string) error
}
