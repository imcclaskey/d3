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
		{"define phase", "my-feature", "define", "Defining my-feature"},
		{"design phase", "my-feature", "design", "Designing my-feature"},
		{"deliver phase", "my-feature", "deliver", "Delivering my-feature"},
		{"unknown phase", "my-feature", "unknown", "Unknowning my-feature"}, // Title case + ing
	}

	g := NewRuleGenerator()
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

	g := NewRuleGenerator()
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

	Templates["core"] = "Core Prefix: {{prefix}}"

	tests := []struct {
		name    string
		feature string
		phase   string
		want    string
		wantErr bool
	}{
		{"valid context", "my-feature", "design", "Core Prefix: Designing my-feature", false},
		{"no context", "", "", "Core Prefix: Ready", false},
	}

	g := NewRuleGenerator()
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
