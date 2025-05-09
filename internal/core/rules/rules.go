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
type RuleGenerator struct{}

// NewRuleGenerator creates a new rule generator
func NewRuleGenerator() *RuleGenerator {
	return &RuleGenerator{}
}

// GeneratePhaseContent generates rule content for a feature and phase
func (g *RuleGenerator) GeneratePhaseContent(feature, phase string) (string, error) {
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
	generator      Generator
	fs             ports.FileSystem
}

// NewService creates a new rules service
func NewService(projectRoot, cursorRulesDir string, generator Generator, fs ports.FileSystem) *Service {
	return &Service{
		projectRoot:    projectRoot,
		cursorRulesDir: cursorRulesDir,
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
