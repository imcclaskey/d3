package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Validator provides workspace validation with error/warning collection
type Validator struct {
	WorkspaceRoot string
	I3Dir         string
	FeaturesDir   string
	Warnings      []string
}

// NewValidator creates a new validator for the workspace
func NewValidator(workspaceRoot string) *Validator {
	i3Dir := filepath.Join(workspaceRoot, ".i3")
	
	return &Validator{
		WorkspaceRoot: workspaceRoot,
		I3Dir:         i3Dir,
		FeaturesDir:   filepath.Join(i3Dir, "features"),
		Warnings:      []string{},
	}
}

// ValidateInit checks if i3 is properly initialized
func (v *Validator) ValidateInit() error {
	return Init(v.I3Dir)
}

// ValidatePhase checks if a phase name is valid
func (v *Validator) ValidatePhase(name string) error {
	return Phase(name)
}

// ValidateFeature checks if a feature exists with all required files
func (v *Validator) ValidateFeature(name string) error {
	return Feature(v.FeaturesDir, name)
}

// CollectWarnings gathers validation warnings
func (v *Validator) CollectWarnings() {
	// Get content warnings
	warnings := ContentWarnings(v.I3Dir)
	v.Warnings = append(v.Warnings, warnings...)
}

// HasWarnings checks if there are any warnings
func (v *Validator) HasWarnings() bool {
	return len(v.Warnings) > 0
}

// FormatWarnings returns warnings as a formatted string
func (v *Validator) FormatWarnings() string {
	if !v.HasWarnings() {
		return ""
	}
	
	var sb strings.Builder
	sb.WriteString("Warnings:\n")
	
	for i, warning := range v.Warnings {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warning))
	}
	
	return sb.String()
}

// ClearWarnings resets the warnings list
func (v *Validator) ClearWarnings() {
	v.Warnings = []string{}
} 