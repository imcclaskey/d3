package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RuleFileGenerator handles rule generation and file operations
type RuleFileGenerator struct {
	templatesDir string
	outputDir    string
}

// NewRuleFileGenerator creates a new rule generator
func NewRuleFileGenerator(templatesDir, outputDir string) *RuleFileGenerator {
	return &RuleFileGenerator{
		templatesDir: templatesDir,
		outputDir:    outputDir,
	}
}

// GenerateRuleContent generates rule content for a phase and feature without writing it
func (g *RuleFileGenerator) GenerateRuleContent(phase, feature string) (string, error) {
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

// generatePrefix creates a formatted prefix showing the current i3 context
func (g *RuleFileGenerator) generatePrefix(feature, phase string) string {
	if feature == "" && phase == "" {
		return "[i3] No active context"
	} else if feature != "" && phase == "" {
		return fmt.Sprintf("[i3] Feature: %s", feature)
	} else if feature == "" && phase != "" {
		return fmt.Sprintf("[i3] Phase: %s", phase)
	}
	return fmt.Sprintf("[i3] Feature: %s | Phase: %s", feature, phase)
}

// EnsureRuleFiles generates and writes core.gen.mdc and phase.gen.mdc files
func (g *RuleFileGenerator) EnsureRuleFiles(phase, feature string) error {
	// Generate content with placeholders replaced
	content, err := g.GenerateRuleContent(phase, feature)
	if err != nil {
		return err
	}
	
	// Ensure i3 directory exists
	i3Dir := filepath.Join(g.outputDir, "i3")
	if err := os.MkdirAll(i3Dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate the prefix for the current context
	prefix := g.generatePrefix(feature, phase)
	
	// Load core template
	coreTemplate, exists := Templates["core"]
	if !exists {
		return fmt.Errorf("core template not found")
	}
	
	// Replace prefix placeholder in core template
	coreContent := strings.ReplaceAll(coreTemplate, "{{prefix}}", prefix)
	
	// Create core.gen.mdc with the processed template
	corePath := filepath.Join(i3Dir, "core.gen.mdc")
	if err := os.WriteFile(corePath, []byte(coreContent), 0644); err != nil {
		return fmt.Errorf("failed to write core rule file: %w", err)
	}
	
	// Create phase.gen.mdc - just write the content directly
	// Template placeholders like {{feature}} and {{phase}} were already replaced by GenerateRuleContent
	phasePath := filepath.Join(i3Dir, "phase.gen.mdc")
	if err := os.WriteFile(phasePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write phase rule file: %w", err)
	}
	
	return nil
}

// ClearRuleFiles removes all rule files from the output directory
func (g *RuleFileGenerator) ClearRuleFiles() error {
	// Get the i3 directory path
	i3Dir := filepath.Join(g.outputDir, "i3")
	
	// Check if directory exists before proceeding
	if _, err := os.Stat(i3Dir); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to clear
		return nil
	}
	
	// Remove the phase.gen.mdc file
	phasePath := filepath.Join(i3Dir, "phase.gen.mdc")
	if err := removeFileIfExists(phasePath); err != nil {
		return fmt.Errorf("removing phase rule file: %w", err)
	}
	
	// Remove the core.gen.mdc file
	corePath := filepath.Join(i3Dir, "core.gen.mdc")
	if err := removeFileIfExists(corePath); err != nil {
		return fmt.Errorf("removing core rule file: %w", err)
	}
	
	return nil
}

// removeFileIfExists removes a file if it exists, does nothing if it doesn't exist
func removeFileIfExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// File doesn't exist, nothing to do
		return nil
	} else if err != nil {
		// Error checking file status
		return fmt.Errorf("checking file %s: %w", path, err)
	}
	
	// File exists, try to remove it
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing file %s: %w", path, err)
	}
	
	return nil
} 