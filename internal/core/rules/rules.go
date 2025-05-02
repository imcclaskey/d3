package rules

import (
	"fmt"
	"strings"
)

// RuleGenerator generates rule content
type RuleGenerator struct{}

// NewRuleGenerator creates a new rule generator
func NewRuleGenerator() *RuleGenerator {
	return &RuleGenerator{}
}

// GenerateRuleContent generates rule content for a phase and feature
func (g *RuleGenerator) GenerateRuleContent(phase, feature string) (string, error) {
	// Only use embedded templates
	template, exists := Templates[phase]
	if !exists {
		return "", fmt.Errorf("template for phase '%s' not found", phase)
	}

	// Render template with replacements
	rendered := template
	rendered = strings.ReplaceAll(rendered, "{{feature}}", feature)
	rendered = strings.ReplaceAll(rendered, "{{phase}}", phase)

	return rendered, nil
}

// GenerateCoreContent generates the core rule content with the current context
func (g *RuleGenerator) GenerateCoreContent(feature, phase string) (string, error) {
	// Load core template
	coreTemplate, exists := Templates["core"]
	if !exists {
		return "", fmt.Errorf("core template not found")
	}

	// Generate the prefix for the current context
	prefix := g.GeneratePrefix(feature, phase)

	// Replace prefix placeholder in core template
	coreContent := strings.ReplaceAll(coreTemplate, "{{prefix}}", prefix)

	return coreContent, nil
}

// GeneratePrefix creates a formatted prefix showing the current i3 context
func (g *RuleGenerator) GeneratePrefix(feature, phase string) string {
	// Return "Ready" if either feature or phase is missing
	if feature == "" || phase == "" {
		return "Ready"
	}

	// Simple map for common phases
	phaseVerbs := map[string]string{
		"ideation":       "Ideating",
		"instruction":    "Instructing",
		"implementation": "Implementing",
	}

	// Get verb form or create it
	verb := phaseVerbs[strings.ToLower(phase)]
	if verb == "" {
		verb = strings.Title(phase) + "ing"
	}

	return fmt.Sprintf("%s %s", verb, feature)
}
