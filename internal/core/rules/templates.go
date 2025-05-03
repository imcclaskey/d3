// Package rules handles template loading and rule file generation
package rules

import (
	_ "embed" // Enable go:embed directive
)

// Templates is a map of phase name to template content
var Templates = map[string]string{
	"core":    coreTemplate,
	"define":  defineTemplate,
	"design":  designTemplate,
	"deliver": deliverTemplate,
}

//go:embed templates/core.md
var coreTemplate string

//go:embed templates/define.md
var defineTemplate string

//go:embed templates/design.md
var designTemplate string

//go:embed templates/deliver.md
var deliverTemplate string
