package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imcclaskey/d3/internal/core/ports"
)

// Generator defines the interface for rule content generation
//
//go:generate mockgen -destination=mocks/mock_generator.go -package=mocks github.com/imcclaskey/d3/internal/core/rules Generator
type Generator interface {
	GeneratePhaseContent(feature, phase string) (string, error)
	GenerateCoreContent(feature, phase string) (string, error)
	GeneratePrefix(feature, phase string) string
}

// RuleGenerator generates rule content
type RuleGenerator struct {
	projectRoot string
	fs          ports.FileSystem
}

// NewRuleGenerator creates a new rule generator
func NewRuleGenerator(projectRoot string, fs ports.FileSystem) *RuleGenerator {
	return &RuleGenerator{
		projectRoot: projectRoot,
		fs:          fs,
	}
}

// getCustomTemplateDir returns the path to the custom templates directory
func (g *RuleGenerator) getCustomTemplateDir() string {
	if g.projectRoot == "" {
		return ""
	}
	return filepath.Join(g.projectRoot, ".d3", "rules")
}

// tryReadCustomTemplate attempts to read a custom template file
func (g *RuleGenerator) tryReadCustomTemplate(templateName string) (string, bool, error) {
	if g.fs == nil || g.projectRoot == "" {
		return "", false, nil
	}

	customDir := g.getCustomTemplateDir()
	templatePath := filepath.Join(customDir, templateName+".md")

	// Check if the template file exists
	_, err := g.fs.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("error checking custom template %s: %w", templatePath, err)
	}

	// Read the template file
	content, err := g.fs.ReadFile(templatePath)
	if err != nil {
		return "", false, fmt.Errorf("error reading custom template %s: %w", templatePath, err)
	}

	return string(content), true, nil
}

// GeneratePhaseContent generates rule content for a feature and phase
func (g *RuleGenerator) GeneratePhaseContent(feature, phase string) (string, error) {
	var template string
	var exists bool
	var err error

	// Try to use custom template first
	template, exists, err = g.tryReadCustomTemplate(phase)
	if err != nil {
		return "", err
	}

	// Fall back to embedded template if custom not found
	if !exists {
		template, exists = Templates[phase]
		if !exists {
			return "", fmt.Errorf("template for phase '%s' not found", phase)
		}
	}

	// Render template with replacements
	rendered := template
	rendered = strings.ReplaceAll(rendered, "{{feature}}", feature)
	rendered = strings.ReplaceAll(rendered, "{{phase}}", phase)

	return rendered, nil
}

// GenerateCoreContent generates the core rule content with the current context
func (g *RuleGenerator) GenerateCoreContent(feature, phase string) (string, error) {
	var coreTemplate string
	var exists bool
	var err error

	// Try to use custom core template first
	coreTemplate, exists, err = g.tryReadCustomTemplate("core")
	if err != nil {
		return "", err
	}

	// Fall back to embedded template if custom not found
	if !exists {
		coreTemplate, exists = Templates["core"]
		if !exists {
			return "", fmt.Errorf("core template not found")
		}
	}

	// Generate the prefix for the current context
	prefix := g.GeneratePrefix(feature, phase)

	// Replace prefix placeholder in core template
	coreContent := strings.ReplaceAll(coreTemplate, "{{prefix}}", prefix)

	return coreContent, nil
}

// GeneratePrefix creates a formatted prefix showing the current d3 context
func (g *RuleGenerator) GeneratePrefix(feature, phase string) string {
	// Return "Ready" if either feature or phase is missing
	if feature == "" || phase == "" {
		return "Ready"
	}

	return fmt.Sprintf("%s - %s", feature, phase)
}

// Service provides rule management operations
type Service struct {
	projectRoot    string
	cursorRulesDir string
	customRulesDir string
	generator      Generator
	fs             ports.FileSystem
}

// NewService creates a new rules service
func NewService(projectRoot, cursorRulesDir string, generator Generator, fs ports.FileSystem) *Service {
	customRulesDir := filepath.Join(projectRoot, ".d3", "rules")
	return &Service{
		projectRoot:    projectRoot,
		cursorRulesDir: cursorRulesDir,
		customRulesDir: customRulesDir,
		generator:      generator,
		fs:             fs,
	}
}

// RefreshRules generates core rule files based on current state
func (s *Service) RefreshRules(feature string, phase string) error {

	// Ensure d3 directory exists
	d3Dir := filepath.Join(s.cursorRulesDir, "d3")
	if err := s.fs.MkdirAll(d3Dir, 0755); err != nil {
		return fmt.Errorf("failed to create rule directory: %w", err)
	}

	if feature != "" {
		// Generate core rule content
		coreContent, err := s.generator.GenerateCoreContent(feature, phase)
		if err != nil {
			return fmt.Errorf("failed to generate core rule: %w", err)
		}
		// Write core rule file
		corePath := filepath.Join(d3Dir, "core.gen.mdc")
		if err := s.fs.WriteFile(corePath, []byte(coreContent), 0644); err != nil {
			return fmt.Errorf("failed to write core rule file: %w", err)
		}
	} else {
		// Delete core rule file if it exists
		corePath := filepath.Join(d3Dir, "core.gen.mdc")
		if err := s.fs.Remove(corePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete core rule file: %w", err)
		}
	}

	if phase == "define" || phase == "design" || phase == "deliver" {
		// Generate phase rule content
		phaseContent, err := s.generator.GeneratePhaseContent(feature, phase)
		if err != nil {
			return fmt.Errorf("failed to generate phase rule: %w", err)
		}

		// Write phase rule file
		phasePath := filepath.Join(d3Dir, "phase.gen.mdc")
		if err := s.fs.WriteFile(phasePath, []byte(phaseContent), 0644); err != nil {
			return fmt.Errorf("failed to write phase rule file: %w", err)
		}
	} else {
		// Delete phase rule file if it exists
		phasePath := filepath.Join(d3Dir, "phase.gen.mdc")
		if err := s.fs.Remove(phasePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete phase rule file: %w", err)
		}
	}

	return nil
}

// ClearGeneratedRules removes all files matching *.gen.mdc in the rule directory.
func (s *Service) ClearGeneratedRules() error {
	d3RuleDir := filepath.Join(s.cursorRulesDir, "d3")
	pattern := filepath.Join(d3RuleDir, "*.gen.mdc")

	// Use the injected filesystem to find matching files
	matches, err := s.fs.Glob(pattern)
	if err != nil {
		// Glob errors might include permission issues, but often indicate pattern syntax error (unlikely here)
		return fmt.Errorf("error finding generated rule files with pattern %s: %w", pattern, err)
	}

	var firstErr error
	for _, match := range matches {
		err := s.fs.Remove(match)
		if err != nil {
			// Log the error but continue trying to remove others
			fmt.Fprintf(os.Stderr, "warning: failed to remove rule file %s: %v\n", match, err)
			// Store the first error encountered
			if firstErr == nil {
				firstErr = fmt.Errorf("failed to remove rule file %s: %w", match, err)
			}
		}
	}

	return firstErr // Return the first error encountered, or nil if all succeeded
}

// InitCustomRulesDir initializes the custom rules directory with copies of the default templates
func (s *Service) InitCustomRulesDir() error {
	// Ensure custom rules directory exists
	if err := s.fs.MkdirAll(s.customRulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom rules directory: %w", err)
	}

	// Process templates in a deterministic order to make testing more reliable
	templateOrder := []string{"core", "define", "design", "deliver"}
	for _, templateName := range templateOrder {
		templateContent, exists := Templates[templateName]
		if !exists {
			continue // Skip if template doesn't exist in the map
		}

		templatePath := filepath.Join(s.customRulesDir, templateName+".md")

		// Skip if file already exists to avoid overwriting user modifications
		if _, err := s.fs.Stat(templatePath); err == nil {
			fmt.Fprintf(os.Stderr, "Custom template %s already exists, skipping\n", templatePath)
			continue
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("error checking template file %s: %w", templatePath, err)
		}

		// Write the template file
		if err := s.fs.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
			return fmt.Errorf("failed to write template file %s: %w", templatePath, err)
		}
	}

	return nil
}
