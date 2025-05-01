package rulegen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generator handles rule generation and file operations
type Generator struct {
	templatesDir string
	outputDir    string
}

// NewGenerator creates a new rule generator
func NewGenerator(templatesDir, outputDir string) *Generator {
	return &Generator{
		templatesDir: templatesDir,
		outputDir:    outputDir,
	}
}

// GenerateContent generates rule content for a phase and feature without writing it
func (g *Generator) GenerateContent(phase, feature string) (string, error) {
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

// CreateRuleFiles generates and writes core.gen.mdc and phase.gen.mdc files
func (g *Generator) CreateRuleFiles(phase, feature string) error {
	// Generate content
	content, err := g.GenerateContent(phase, feature)
	if err != nil {
		return err
	}
	
	// Ensure i3 directory exists
	i3Dir := filepath.Join(g.outputDir, "i3")
	if err := os.MkdirAll(i3Dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Create core.gen.mdc - this would contain shared rules
	coreContent := "---\n"
	coreContent += "description: Core rules for i3 framework\n"
	coreContent += "globs: \n"
	coreContent += "alwaysApply: true\n"
	coreContent += "---\n\n"
	coreContent += "# i3 Core Rules\n\n"
	coreContent += "These rules apply to all phases of the i3 framework.\n\n"
	corePath := filepath.Join(i3Dir, "core.gen.mdc")
	if err := os.WriteFile(corePath, []byte(coreContent), 0644); err != nil {
		return fmt.Errorf("failed to write core rule file: %w", err)
	}
	
	// Create phase.gen.mdc - contains the phase-specific content
	phaseContent := "---\n"
	phaseContent += "description: Phase-specific rules for i3 framework\n"
	phaseContent += "globs: \n"
	phaseContent += "alwaysApply: true\n"
	phaseContent += "---\n\n"
	phaseContent += fmt.Sprintf("# i3 Phase Rules: %s\n\n", phase)
	if feature != "" {
		phaseContent += fmt.Sprintf("Feature: %s\n\n", feature)
	}
	phaseContent += content
	phasePath := filepath.Join(i3Dir, "phase.gen.mdc")
	if err := os.WriteFile(phasePath, []byte(phaseContent), 0644); err != nil {
		return fmt.Errorf("failed to write phase rule file: %w", err)
	}
	
	return nil
} 