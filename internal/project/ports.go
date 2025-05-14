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
}

// PhaseServicer defines the interface for phase management operations.
type PhaseServicer interface {
	EnsurePhaseFiles(featureDir string) error
}

// FileOperator defines operations for project file manipulations needed by ProjectService.
type FileOperator interface {
	EnsureMCPJSON(fs ports.FileSystem, projectRoot string) error
	EnsureD3GitignoreEntries(fs ports.FileSystem, d3DirAbs, cursorRulesD3DirAbs, projectRootAbs string) error
	EnsureProjectFiles(fs ports.FileSystem, d3DirAbs string) error
}
