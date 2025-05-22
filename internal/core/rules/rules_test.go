package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "embed"

	"github.com/golang/mock/gomock"

	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	rulesmocks "github.com/imcclaskey/d3/internal/core/rules/mocks" // Import generated mock for Generator
)

func TestRuleGenerator_GeneratePrefix(t *testing.T) {
	tests := []struct {
		name    string
		feature string
		phase   string
		want    string
	}{
		{"no context", "", "", "Ready"},
		{"feature only", "feat1", "", "Ready"},
		{"phase only", "", "define", "Ready"},
		{"define phase", "my-feature", "define", "my-feature - define"},
		{"design phase", "my-feature", "design", "my-feature - design"},
		{"deliver phase", "my-feature", "deliver", "my-feature - deliver"},
		{"unknown phase with feature", "my-feature", "unknown", "my-feature - unknown"},
		// Add a case for a phase with different casing to ensure it's used as-is
		{"mixedCase phase", "another-feat", "DefineNew", "another-feat - DefineNew"},
	}

	// Create a mock filesystem since the constructor now requires it
	ctrl := gomock.NewController(t)
	mockFS := portsmocks.NewMockFileSystem(ctrl)
	g := NewRuleGenerator("", mockFS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := g.GeneratePrefix(tt.feature, tt.phase); got != tt.want {
				t.Errorf("GeneratePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleGenerator_GeneratePhaseContent(t *testing.T) {
	// Temporarily store and restore original templates if modification is needed for tests
	// For now, assume templates are loaded correctly via //go:embed
	originalTemplates := make(map[string]string)
	for k, v := range Templates {
		originalTemplates[k] = v
	}
	// Restore templates after test
	defer func() {
		Templates = originalTemplates
	}()

	// Setup some dummy templates for testing interpolation
	Templates["define"] = "Define: {{feature}} / {{phase}}"
	Templates["custom"] = "Custom: {{feature}} only"

	tests := []struct {
		name    string
		feature string
		phase   string
		want    string
		wantErr bool
	}{
		{"valid define phase", "feat1", "define", "Define: feat1 / define", false},
		{"valid custom phase", "feat2", "custom", "Custom: feat2 only", false},
		{"missing phase template", "feat3", "nonexistent", "", true},
	}

	// Create a mock filesystem since the constructor now requires it
	ctrl := gomock.NewController(t)
	mockFS := portsmocks.NewMockFileSystem(ctrl)
	g := NewRuleGenerator("", mockFS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GeneratePhaseContent(tt.feature, tt.phase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePhaseContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GeneratePhaseContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleGenerator_GenerateCoreContent(t *testing.T) {
	// Similar template setup as GeneratePhaseContent if needed
	originalTemplates := make(map[string]string)
	for k, v := range Templates {
		originalTemplates[k] = v
	}
	defer func() {
		Templates = originalTemplates
	}()

	Templates["core"] = "### d3 - {{prefix}}"

	tests := []struct {
		name    string
		feature string
		phase   string
		want    string
		wantErr bool
	}{
		{"valid context", "my-feature", "design", "### d3 - my-feature - design", false},
		{"no context", "", "", "### d3 - Ready", false},
	}

	// Create a mock filesystem since the constructor now requires it
	ctrl := gomock.NewController(t)
	mockFS := portsmocks.NewMockFileSystem(ctrl)
	g := NewRuleGenerator("", mockFS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := g.GenerateCoreContent(tt.feature, tt.phase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCoreContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateCoreContent() got = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case for missing core template
	t.Run("missing core template", func(t *testing.T) {
		delete(Templates, "core") // Remove core template
		_, err := g.GenerateCoreContent("any", "any")
		if err == nil {
			t.Errorf("GenerateCoreContent() did not return error when core template was missing")
		}
	})
}

func TestRuleGenerator_tryReadCustomTemplate(t *testing.T) {
	projectRoot := "/test/project"
	customTemplateDir := filepath.Join(projectRoot, ".d3", "rules")

	tests := []struct {
		name         string
		templateName string
		setupMocks   func(mockFS *portsmocks.MockFileSystem)
		wantContent  string
		wantExists   bool
		wantErr      bool
	}{
		{
			name:         "custom template exists",
			templateName: "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "define.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return([]byte("custom define template"), nil).Times(1)
			},
			wantContent: "custom define template",
			wantExists:  true,
			wantErr:     false,
		},
		{
			name:         "custom template does not exist",
			templateName: "nonexistent",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "nonexistent.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantContent: "",
			wantExists:  false,
			wantErr:     false,
		},
		{
			name:         "stat error",
			templateName: "error",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "error.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, fmt.Errorf("stat error")).Times(1)
			},
			wantContent: "",
			wantExists:  false,
			wantErr:     true,
		},
		{
			name:         "read file error",
			templateName: "readerror",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "readerror.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			wantContent: "",
			wantExists:  false,
			wantErr:     true,
		},
		{
			name:         "nil filesystem",
			templateName: "any",
			setupMocks:   func(mockFS *portsmocks.MockFileSystem) {},
			wantContent:  "",
			wantExists:   false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS)
			}

			g := &RuleGenerator{
				projectRoot: projectRoot,
				fs:          mockFS,
			}

			// Special case for "nil filesystem" test
			if tt.name == "nil filesystem" {
				g.fs = nil
			}

			gotContent, gotExists, err := g.tryReadCustomTemplate(tt.templateName)
			if (err != nil) != tt.wantErr {
				t.Errorf("tryReadCustomTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotContent != tt.wantContent {
				t.Errorf("tryReadCustomTemplate() gotContent = %v, want %v", gotContent, tt.wantContent)
			}
			if gotExists != tt.wantExists {
				t.Errorf("tryReadCustomTemplate() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestRuleGenerator_GeneratePhaseContentWithCustomTemplate(t *testing.T) {
	projectRoot := "/test/project"
	customTemplateDir := filepath.Join(projectRoot, ".d3", "rules")

	// Temporarily store and restore original templates if modification is needed for tests
	originalTemplates := make(map[string]string)
	for k, v := range Templates {
		originalTemplates[k] = v
	}
	// Restore templates after test
	defer func() {
		Templates = originalTemplates
	}()

	// Setup some dummy templates for testing interpolation
	Templates["define"] = "Default Define: {{feature}} / {{phase}}"
	Templates["design"] = "Default Design: {{feature}} / {{phase}}"

	tests := []struct {
		name       string
		feature    string
		phase      string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		want       string
		wantErr    bool
	}{
		{
			name:    "use custom template",
			feature: "feat1",
			phase:   "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "define.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return([]byte("Custom Define: {{feature}} / {{phase}}"), nil).Times(1)
			},
			want:    "Custom Define: feat1 / define",
			wantErr: false,
		},
		{
			name:    "fallback to default template",
			feature: "feat2",
			phase:   "design",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "design.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
			},
			want:    "Default Design: feat2 / design",
			wantErr: false,
		},
		{
			name:    "custom template read error",
			feature: "feat3",
			phase:   "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "define.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			want:    "",
			wantErr: true,
		},
		{
			name:    "non-existent template",
			feature: "feat4",
			phase:   "nonexistent",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "nonexistent.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS)
			}

			g := &RuleGenerator{
				projectRoot: projectRoot,
				fs:          mockFS,
			}

			got, err := g.GeneratePhaseContent(tt.feature, tt.phase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GeneratePhaseContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GeneratePhaseContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleGenerator_GenerateCoreContentWithCustomTemplate(t *testing.T) {
	projectRoot := "/test/project"
	customTemplateDir := filepath.Join(projectRoot, ".d3", "rules")

	// Temporarily store and restore original templates if modification is needed for tests
	originalTemplates := make(map[string]string)
	for k, v := range Templates {
		originalTemplates[k] = v
	}
	// Restore templates after test
	defer func() {
		Templates = originalTemplates
	}()

	// Setup some dummy templates for testing interpolation
	Templates["core"] = "Default Core: ### d3 - {{prefix}}"

	tests := []struct {
		name       string
		feature    string
		phase      string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		want       string
		wantErr    bool
	}{
		{
			name:    "use custom core template",
			feature: "my-feature",
			phase:   "design",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "core.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return([]byte("Custom Core: ### d3 - {{prefix}}"), nil).Times(1)
			},
			want:    "Custom Core: ### d3 - my-feature - design",
			wantErr: false,
		},
		{
			name:    "fallback to default core template",
			feature: "other-feature",
			phase:   "deliver",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "core.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
			},
			want:    "Default Core: ### d3 - other-feature - deliver",
			wantErr: false,
		},
		{
			name:    "custom core template read error",
			feature: "error-feature",
			phase:   "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				templatePath := filepath.Join(customTemplateDir, "core.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
				mockFS.EXPECT().ReadFile(templatePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS)
			}

			g := &RuleGenerator{
				projectRoot: projectRoot,
				fs:          mockFS,
			}

			got, err := g.GenerateCoreContent(tt.feature, tt.phase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCoreContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateCoreContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_InitCustomRulesDir(t *testing.T) {
	projectRoot := "/test/project"
	cursorRulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	customRulesDir := filepath.Join(projectRoot, ".d3", "rules")

	// Create a fixed template set for testing to avoid issues with map iteration order
	originalTemplates := Templates
	testTemplates := map[string]string{
		"core":    "core template content",
		"define":  "define template content",
		"design":  "design template content",
		"deliver": "deliver template content",
	}

	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		wantErr    bool
	}{
		{
			name: "successful initialization",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Create directory
				mockFS.EXPECT().MkdirAll(customRulesDir, os.FileMode(0755)).Return(nil).Times(1)

				// For each template in our fixed test templates, check if it exists and write if not
				for name, content := range testTemplates {
					templatePath := filepath.Join(customRulesDir, name+".md")
					mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
					mockFS.EXPECT().WriteFile(templatePath, []byte(content), os.FileMode(0644)).Return(nil).Times(1)
				}
			},
			wantErr: false,
		},
		{
			name: "directory creation fails",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().MkdirAll(customRulesDir, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "template already exists",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Create directory
				mockFS.EXPECT().MkdirAll(customRulesDir, os.FileMode(0755)).Return(nil).Times(1)

				// First template exists, others don't
				firstFound := false
				for name, content := range testTemplates {
					templatePath := filepath.Join(customRulesDir, name+".md")

					if !firstFound {
						// First template exists
						mockFS.EXPECT().Stat(templatePath).Return(nil, nil).Times(1)
						firstFound = true
					} else {
						// Other templates don't exist
						mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
						mockFS.EXPECT().WriteFile(templatePath, []byte(content), os.FileMode(0644)).Return(nil).Times(1)
					}
				}
			},
			wantErr: false,
		},
		{
			name: "stat error",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Create directory
				mockFS.EXPECT().MkdirAll(customRulesDir, os.FileMode(0755)).Return(nil).Times(1)

				// Test with the first template in our fixed order
				templatePath := filepath.Join(customRulesDir, "core.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, fmt.Errorf("stat error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "write file error",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Create directory
				mockFS.EXPECT().MkdirAll(customRulesDir, os.FileMode(0755)).Return(nil).Times(1)

				// Test with the first template in our fixed order
				templatePath := filepath.Join(customRulesDir, "core.md")
				mockFS.EXPECT().Stat(templatePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().WriteFile(templatePath, []byte("core template content"), os.FileMode(0644)).Return(fmt.Errorf("write error")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)
			mockGen := rulesmocks.NewMockGenerator(ctrl)

			if tt.setupMocks != nil {
				// Replace Templates with our fixed test templates for this test
				Templates = testTemplates
				defer func() {
					// Restore original templates after the test
					Templates = originalTemplates
				}()

				tt.setupMocks(mockFS)
			}

			s := NewService(projectRoot, cursorRulesDir, mockGen, mockFS)
			err := s.InitCustomRulesDir()

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.InitCustomRulesDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Tests for Service ---

func TestService_RefreshRules(t *testing.T) {
	projectRoot := "/test/project"
	cursorRulesDir := filepath.Join(projectRoot, ".cursor", "rules")
	d3Dir := filepath.Join(cursorRulesDir, "d3")
	corePath := filepath.Join(d3Dir, "core.gen.mdc")
	phasePath := filepath.Join(d3Dir, "phase.gen.mdc")

	tests := []struct {
		name       string
		feature    string
		phase      string
		setupMocks func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator)
		wantErr    bool
	}{
		{
			name:    "successful core only (no phase)",
			feature: "feat1",
			phase:   "",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat1", "").Return("core content", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core content"), os.FileMode(0644)).Return(nil).Times(1)
				mockFS.EXPECT().Remove(phasePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "successful core and phase (define)",
			feature: "feat2",
			phase:   "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat2", "define").Return("core content 2", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core content 2"), os.FileMode(0644)).Return(nil).Times(1)
				mockGen.EXPECT().GeneratePhaseContent("feat2", "define").Return("define content", nil).Times(1)
				mockFS.EXPECT().WriteFile(phasePath, []byte("define content"), os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "MkdirAll fails",
			feature: "feat3",
			phase:   "design",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:    "GenerateCoreContent fails",
			feature: "feat4",
			phase:   "deliver",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat4", "deliver").Return("", fmt.Errorf("core gen failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:    "WriteFile core fails",
			feature: "feat5",
			phase:   "define",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat5", "define").Return("core ok", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core ok"), os.FileMode(0644)).Return(fmt.Errorf("core write failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:    "GeneratePhaseContent fails",
			feature: "feat6",
			phase:   "design",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat6", "design").Return("core ok 6", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core ok 6"), os.FileMode(0644)).Return(nil).Times(1)
				mockGen.EXPECT().GeneratePhaseContent("feat6", "design").Return("", fmt.Errorf("phase gen failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:    "WriteFile phase fails",
			feature: "feat7",
			phase:   "deliver",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("feat7", "deliver").Return("core ok 7", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core ok 7"), os.FileMode(0644)).Return(nil).Times(1)
				mockGen.EXPECT().GeneratePhaseContent("feat7", "deliver").Return("deliver content", nil).Times(1)
				mockFS.EXPECT().WriteFile(phasePath, []byte("deliver content"), os.FileMode(0644)).Return(fmt.Errorf("phase write failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:    "no feature, no phase - core and phase removed",
			feature: "",
			phase:   "",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				// No GenerateCoreContent or WriteFile for core
				mockFS.EXPECT().Remove(corePath).Return(nil).Times(1)
				// No GeneratePhaseContent or WriteFile for phase
				mockFS.EXPECT().Remove(phasePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "no feature, no phase - core and phase removed (idempotent)",
			feature: "",
			phase:   "",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().Remove(corePath).Return(os.ErrNotExist).Times(1)  // Return ErrNotExist
				mockFS.EXPECT().Remove(phasePath).Return(os.ErrNotExist).Times(1) // Return ErrNotExist
			},
			wantErr: false,
		},
		{
			name:    "feature present, invalid phase - core written, phase removed",
			feature: "featValid",
			phase:   "invalidPhase",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("featValid", "invalidPhase").Return("core content valid", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core content valid"), os.FileMode(0644)).Return(nil).Times(1)
				// No GeneratePhaseContent or WriteFile for phase
				mockFS.EXPECT().Remove(phasePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "no feature, valid phase - core removed, phase written",
			feature: "",
			phase:   "define", // A valid phase
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				// Core file removed
				mockFS.EXPECT().Remove(corePath).Return(nil).Times(1)
				// Phase file generated and written
				mockGen.EXPECT().GeneratePhaseContent("", "define").Return("define content no feature", nil).Times(1)
				mockFS.EXPECT().WriteFile(phasePath, []byte("define content no feature"), os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:    "Remove core fails (not IsNotExist)",
			feature: "", // Triggers core removal
			phase:   "", // Triggers phase removal
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().Remove(corePath).Return(fmt.Errorf("some disk error")).Times(1)
				// Remove for phasePath might or might not be called depending on early exit.
				// For simplicity, we don't set an expectation for it here,
				// as the error from corePath removal should cause a failure.
			},
			wantErr: true,
		},
		{
			name:    "Remove phase fails (not IsNotExist)",
			feature: "featWithPhase",
			phase:   "invalidPhase", // Triggers phase removal
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockGen *rulesmocks.MockGenerator) {
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockGen.EXPECT().GenerateCoreContent("featWithPhase", "invalidPhase").Return("core content", nil).Times(1)
				mockFS.EXPECT().WriteFile(corePath, []byte("core content"), os.FileMode(0644)).Return(nil).Times(1)
				mockFS.EXPECT().Remove(phasePath).Return(fmt.Errorf("some disk error")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)
			mockGen := rulesmocks.NewMockGenerator(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockGen)
			}

			s := NewService(projectRoot, cursorRulesDir, mockGen, mockFS)
			err := s.RefreshRules(tt.feature, tt.phase)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.RefreshRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
