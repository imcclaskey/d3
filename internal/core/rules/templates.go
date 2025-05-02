// Package rules handles template loading and rule file generation
package rules

import (
	_ "embed" // Enable go:embed directive
)

// Templates is a map of phase name to template content
var Templates = map[string]string{
	"core":           coreTemplate,
	"ideation":       ideationTemplate,
	"instruction":    instructionTemplate,
	"implementation": implementationTemplate,
}

//go:embed templates/core.md
var coreTemplate string

//go:embed templates/ideation.md
var ideationTemplate string

//go:embed templates/instruction.md
var instructionTemplate string

//go:embed templates/implementation.md
var implementationTemplate string
