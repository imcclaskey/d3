package project

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/core/feature"
	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	"github.com/imcclaskey/d3/internal/core/session"
	"github.com/imcclaskey/d3/internal/testutil" // Import shared test utilities
)

// Helper to create a default project with gomocks for testing
// This function now returns all the mocks it creates so they can be used for setting expectations.
func newTestProjectWithMocks(t *testing.T, ctrl *gomock.Controller) (*Project, *portsmocks.MockFileSystem, *MockStorageService, *MockFeatureServicer, *MockRulesServicer, *MockPhaseServicer) {
	t.Helper()
	projectRoot := t.TempDir() // Using t.TempDir() for proper test isolation

	mockFS := portsmocks.NewMockFileSystem(ctrl)
	mockSessionSvc := NewMockStorageService(ctrl)
	mockFeatureSvc := NewMockFeatureServicer(ctrl)
	mockRulesSvc := NewMockRulesServicer(ctrl)
	mockPhaseSvc := NewMockPhaseServicer(ctrl)

	// New() is now "dumb" and does not perform I/O for initialization status.
	// The _isInitialized field will be nil.
	// The first call to proj.IsInitialized() in a test will trigger fs.Stat().
	proj := New(projectRoot, mockFS, mockSessionSvc, mockFeatureSvc, mockRulesSvc, mockPhaseSvc)
	return proj, mockFS, mockSessionSvc, mockFeatureSvc, mockRulesSvc, mockPhaseSvc
}

// TestProject_New tests the New function (which is now very simple)
func TestProject_New_WithGoMock(t *testing.T) {
	ctrl := gomock.NewController(t)

	projectRoot := "/testroot_new_gomock" // Use a distinct root

	mockFS := portsmocks.NewMockFileSystem(ctrl)
	mockSessionSvc := NewMockStorageService(ctrl)
	mockFeatureSvc := NewMockFeatureServicer(ctrl)
	mockRulesSvc := NewMockRulesServicer(ctrl)
	mockPhaseSvc := NewMockPhaseServicer(ctrl)

	// New() no longer performs I/O, so no fs.Stat mock needed here.
	proj := New(projectRoot, mockFS, mockSessionSvc, mockFeatureSvc, mockRulesSvc, mockPhaseSvc)

	if proj == nil {
		t.Fatal("New() returned nil")
	}
	if proj.fs != mockFS {
		t.Errorf("Expected fs to be the mockFS")
	}
	if proj.session != mockSessionSvc {
		t.Errorf("Expected session to be mockSessionSvc")
	}
	if proj.features != mockFeatureSvc {
		t.Errorf("Expected features to be mockFeatureSvc")
	}
	if proj.rules != mockRulesSvc {
		t.Errorf("Expected rules to be mockRulesSvc")
	}
	if proj.phases != mockPhaseSvc {
		t.Errorf("Expected phases to be mockPhaseSvc")
	}
	if proj.state.ProjectRoot != projectRoot {
		t.Errorf("Expected ProjectRoot to be '%s', got '%s'", projectRoot, proj.state.ProjectRoot)
	}
	expectedD3Dir := filepath.Join(projectRoot, ".d3")
	if proj.state.D3Dir != expectedD3Dir {
		t.Errorf("Expected D3Dir to be '%s', got '%s'", expectedD3Dir, proj.state.D3Dir)
	}
	// proj.isInitialized is no longer a field.
	// The IsInitialized() method now performs the check directly.
	// We can verify its initial state by calling it, but need to mock fs.Stat
	// For this test, we just verify New() doesn't crash and assigns services correctly.
}

func TestProject_IsInitialized(t *testing.T) {
	type statResult struct {
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name string
		// The Stat call that IsInitialized() will perform.
		mockStat     statResult
		expectIsInit bool
	}{
		{
			name:         "d3 directory does not exist",
			mockStat:     statResult{info: nil, err: os.ErrNotExist},
			expectIsInit: false,
		},
		{
			name:         "d3 directory exists",
			mockStat:     statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil},
			expectIsInit: true,
		},
		{
			name:     "stat returns other error",
			mockStat: statResult{info: nil, err: fmt.Errorf("some stat error")},
			// checkInitialized returns false if Stat returns an error.
			expectIsInit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, _, _, _, _ := newTestProjectWithMocks(t, ctrl)

			d3DirForCheck := proj.state.D3Dir

			// Expect the fs.Stat call made directly by proj.IsInitialized() -> checkInitialized()
			mockFS.EXPECT().Stat(d3DirForCheck).Return(tt.mockStat.info, tt.mockStat.err).Times(1)

			// Call IsInitialized() - this triggers the Stat call
			if got := proj.IsInitialized(); got != tt.expectIsInit {
				t.Errorf("IsInitialized() = %v, want %v", got, tt.expectIsInit)
			}

			// NOTE: Removed check for caching, as IsInitialized now always calls Stat.
		})
	}
}

func TestProject_RequiresInitialized(t *testing.T) {
	type statResult struct {
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name string
		// The Stat call that RequiresInitialized() -> IsInitialized() will perform.
		mockStat    statResult
		expectError error
	}{
		{
			name:        "not initialized",
			mockStat:    statResult{info: nil, err: os.ErrNotExist},
			expectError: ErrNotInitialized,
		},
		{
			name:        "initialized",
			mockStat:    statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil},
			expectError: nil,
		},
		{
			name:     "stat returns other error",
			mockStat: statResult{info: nil, err: fmt.Errorf("some stat error")},
			// If Stat fails, checkInitialized is false, IsInitialized is false, RequiresInitialized returns ErrNotInitialized
			expectError: ErrNotInitialized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, _, _, _, _ := newTestProjectWithMocks(t, ctrl)

			d3DirForCheck := proj.state.D3Dir

			// Expect the fs.Stat call made by proj.RequiresInitialized() -> IsInitialized() -> checkInitialized()
			mockFS.EXPECT().Stat(d3DirForCheck).Return(tt.mockStat.info, tt.mockStat.err).Times(1)

			gotError := proj.RequiresInitialized()

			// Check if the error matches the expected error type/value
			if !errors.Is(gotError, tt.expectError) {
				// Use errors.Is for checking wrapped errors or specific error values.
				// If tt.expectError is nil, errors.Is(gotError, nil) is equivalent to gotError == nil.
				t.Errorf("RequiresInitialized() error = %v, wantErr %v", gotError, tt.expectError)
			}
		})
	}
}

func TestProject_Init(t *testing.T) {
	type args struct {
		clean bool
	}
	tests := []struct {
		name string
		args args
		// Single function to set up all mocks for the test case
		setupMocks         func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr            bool
		wantResultMsg      string
		verifyProjectState func(t *testing.T, proj *Project)
	}{
		{
			name: "already initialized, not clean",
			args: args{clean: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				// Expect the Stat call from Init -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// No other mocks needed as Init returns early
			},
			wantErr:       false,
			wantResultMsg: "Project already initialized.",
		},
		{
			name: "not initialized, successful init",
			args: args{clean: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				// Expect the Stat call from Init -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				// Expect subsequent calls within Init
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().ClearActiveFeature().Return(nil).Times(1)

				// Mocks for ensureGitignoreEntries
				d3GitignorePath := filepath.Join(proj.state.D3Dir, ".gitignore")
				cursorRulesD3Dir := filepath.Join(proj.state.CursorRulesDir, "d3")
				cursorGitignorePath := filepath.Join(cursorRulesD3Dir, ".gitignore")
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1) // Called by ensureGitignoreEntries for .d3/.gitignore
				mockFS.EXPECT().WriteFile(d3GitignorePath, []byte(".session\n"), os.FileMode(0644)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(cursorRulesD3Dir, os.FileMode(0755)).Return(nil).Times(1) // Called by ensureGitignoreEntries for .cursor/rules/d3/.gitignore
				mockFS.EXPECT().WriteFile(cursorGitignorePath, []byte("*.gen.rules\n"), os.FileMode(0644)).Return(nil).Times(1)

				mockRules.EXPECT().RefreshRules("", "").Return(nil).Times(1)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully. Cursor rules have been updated.",
			verifyProjectState: func(t *testing.T, proj *Project) {
				mockFS, ok := proj.fs.(*portsmocks.MockFileSystem)
				if !ok {
					t.Fatal("proj.fs is not a mock in verifyProjectState")
				}
				// Expect the Stat call *before* calling IsInitialized again.
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				if !proj.IsInitialized() {
					t.Errorf("Expected proj.IsInitialized() to be true after successful init")
				}
			},
		},
		{
			name: "initialized, clean init",
			args: args{clean: true},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				// Expect the Stat call from Init -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)

				// Define variables needed within InOrder block first
				d3GitignorePath := filepath.Join(proj.state.D3Dir, ".gitignore")
				cursorRulesD3Dir := filepath.Join(proj.state.CursorRulesDir, "d3")
				cursorGitignorePath := filepath.Join(cursorRulesD3Dir, ".gitignore")

				// Expect subsequent calls within Init in order
				gomock.InOrder(
					mockFS.EXPECT().RemoveAll(proj.state.D3Dir).Return(nil).Times(1),
					mockSession.EXPECT().ClearActiveFeature().Return(nil).Times(1), // Called during clean
					mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1),
					mockSession.EXPECT().ClearActiveFeature().Return(nil).Times(1), // Called again by Init logic
					// Mocks for ensureGitignoreEntries within InOrder
					mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().WriteFile(d3GitignorePath, []byte(".session\n"), os.FileMode(0644)).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(cursorRulesD3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().WriteFile(cursorGitignorePath, []byte("*.gen.rules\n"), os.FileMode(0644)).Return(nil).Times(1),
					mockRules.EXPECT().RefreshRules("", "").Return(nil).Times(1),
				)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name: "error on RemoveAll during clean init",
			args: args{clean: true},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(proj.state.D3Dir).Return(fmt.Errorf("failed to remove")).Times(1)
				mockSession.EXPECT().ClearActiveFeature().Return(nil).AnyTimes() // Allow clean's ClearActiveFeature if it happens before error
			},
			wantErr: true,
		},
		{
			name: "error on MkdirAll for .d3",
			args: args{clean: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(fmt.Errorf("failed to mkdir .d3")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on session.ClearActiveFeature",
			args: args{clean: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().ClearActiveFeature().Return(fmt.Errorf("session clear failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on rules.RefreshRules",
			args: args{clean: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().ClearActiveFeature().Return(nil).Times(1)
				// Mocks for ensureGitignoreEntries (success case needed before rules fail)
				d3GitignorePath := filepath.Join(proj.state.D3Dir, ".gitignore")
				cursorRulesD3Dir := filepath.Join(proj.state.CursorRulesDir, "d3")
				cursorGitignorePath := filepath.Join(cursorRulesD3Dir, ".gitignore")
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(d3GitignorePath, []byte(".session\n"), os.FileMode(0644)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(cursorRulesD3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(cursorGitignorePath, []byte("*.gen.rules\n"), os.FileMode(0644)).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", "").Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockSession, _, mockRules, mockPhase := newTestProjectWithMocks(t, ctrl)

			// Set up all mocks for this test case
			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockRules, mockPhase)
			}
			_ = mockPhase // Avoid unused variable error if setupMocks doesn't use it

			result, err := proj.Init(tt.args.clean)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != nil && result.FormatCLI() != tt.wantResultMsg {
				t.Errorf("Init() result message = %s, want %s", result.FormatCLI(), tt.wantResultMsg)
			}
			if tt.verifyProjectState != nil {
				tt.verifyProjectState(t, proj)
			}
		})
	}
}

func TestProject_CreateFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name string
		args args
		// Single function to set up all mocks for the test case
		setupMocks                 func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr                    bool
		verifyProjectStateAndMocks func(t *testing.T, proj *Project, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				// Expect Stat from CreateFeature -> RequiresInitialized -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				// No other mocks needed as it returns early
			},
			wantErr: true, // Expect ErrNotInitialized
		},
		{
			name: "featureSvc.CreateFeature fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(nil, fmt.Errorf("create feature failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "sessionSvc.SaveActiveFeature fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featurePath := filepath.Join(proj.state.FeaturesDir, "test-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: featurePath}, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("test-feature").Return(fmt.Errorf("save session failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rulesSvc.RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featurePath := filepath.Join(proj.state.FeaturesDir, "test-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: featurePath}, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("test-feature").Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("test-feature", session.Define.String()).Return(fmt.Errorf("refresh rules failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful feature creation",
			args: args{ctx: context.Background(), featureName: "new-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featurePath := filepath.Join(proj.state.FeaturesDir, "new-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "new-feature").Return(&feature.FeatureInfo{Name: "new-feature", Path: featurePath}, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("new-feature").Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("new-feature", session.Define.String()).Return(nil).Times(1)
			},
			wantErr: false,
			verifyProjectStateAndMocks: func(t *testing.T, proj *Project, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				if proj.state.CurrentFeature != "new-feature" {
					t.Errorf("Project state CurrentFeature = %s, want new-feature", proj.state.CurrentFeature)
				}
				if proj.state.CurrentPhase != session.Define {
					t.Errorf("Project state CurrentPhase = %s, want Define", proj.state.CurrentPhase)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockSession, mockFeature, mockRules, mockPhase := newTestProjectWithMocks(t, ctrl)

			// Set up all mocks for this test case
			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockFeature, mockRules, mockPhase)
			}

			_, err := proj.CreateFeature(tt.args.ctx, tt.args.featureName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateFeature() error = nil, wantErr %v", tt.wantErr)
				} else if tt.name == "project not initialized" && !errors.Is(err, ErrNotInitialized) {
					// Special check for ErrNotInitialized
					t.Errorf("CreateFeature() error = %v, want specific error %v", err, ErrNotInitialized)
				}
				// Add other specific error checks if needed
			} else if err != nil {
				t.Errorf("CreateFeature() unexpected error = %v", err)
			}

			if !tt.wantErr && tt.verifyProjectStateAndMocks != nil {
				tt.verifyProjectStateAndMocks(t, proj, mockSession, mockFeature, mockRules, mockPhase)
			}
		})
	}
}

func TestProject_ChangePhase(t *testing.T) {
	type args struct {
		ctx         context.Context
		targetPhase session.Phase
	}
	tests := []struct {
		name string
		args args
		// Single function to set up all mocks and initial state for the test case
		setupMocksAndState         func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr                    bool
		wantResultMsgContains      string
		verifyProjectStateAndMocks func(t *testing.T, proj *Project, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				// Expect Stat from ChangePhase -> RequiresInitialized -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true, // Expect ErrNotInitialized
		},
		{
			name: "session.LoadActiveFeature fails (during internal check)",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				// Set initial state to force LoadActiveFeature check
				proj.state.CurrentFeature = ""
				mockSession.EXPECT().LoadActiveFeature().Return("", fmt.Errorf("session load failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "no active feature",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				// Set initial state
				proj.state.CurrentFeature = ""
				proj.state.CurrentPhase = session.None
				// Expect the LoadActiveFeature check
				mockSession.EXPECT().LoadActiveFeature().Return("", nil).Times(1) // No feature found
			},
			wantErr: true, // Expect ErrNoActiveFeature
		},
		{
			name: "already in target phase",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				// Set initial state
				proj.state.CurrentFeature = "feat1"
				proj.state.CurrentPhase = session.Design
				// ChangePhase checks state directly, shouldn't need LoadActiveFeature here
			},
			wantErr:               false,
			wantResultMsgContains: "Already in the design phase.",
		},
		{
			name: "featureSvc.SetFeaturePhase fails",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				// Set initial state
				featureName := "feat1"
				proj.state.CurrentFeature = featureName
				proj.state.CurrentPhase = session.Design
				// Expect SetFeaturePhase to fail
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(fmt.Errorf("set phase failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rules.RefreshRules fails after phase change",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featureName := "feat1"
				proj.state.CurrentFeature = featureName
				proj.state.CurrentPhase = session.Design
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, session.Deliver.String()).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful phase change, no existing phase dir",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featureName := "feat1"
				proj.state.CurrentFeature = featureName
				proj.state.CurrentPhase = session.Design
				featureDir := filepath.Join(proj.state.FeaturesDir, featureName)
				phaseToCheckDir := filepath.Join(featureDir, session.Deliver.String())

				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, session.Deliver.String()).Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featureDir).Return(nil).Times(1)
				mockFS.EXPECT().Stat(phaseToCheckDir).Return(nil, os.ErrNotExist).Times(1) // Stat for impact check
			},
			wantErr:               false,
			wantResultMsgContains: "Moved to deliver phase.",
			verifyProjectStateAndMocks: func(t *testing.T, proj *Project, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				if proj.state.CurrentPhase != session.Deliver {
					t.Errorf("Project state CurrentPhase = %s, want deliver", proj.state.CurrentPhase)
				}
			},
		},
		{
			name: "successful phase change, with existing phase dir (hasImpact)",
			args: args{ctx: context.Background(), targetPhase: session.Define},
			setupMocksAndState: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				featureName := "featImpact"
				proj.state.CurrentFeature = featureName
				proj.state.CurrentPhase = session.Design
				featureDir := filepath.Join(proj.state.FeaturesDir, featureName)
				phaseToCheckDir := filepath.Join(featureDir, session.Define.String())

				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Define).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, session.Define.String()).Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featureDir).Return(nil).Times(1)
				mockFS.EXPECT().Stat(phaseToCheckDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // Stat for impact check
			},
			wantErr:               false,
			wantResultMsgContains: "Moved to define phase. Note: Existing files were detected",
			verifyProjectStateAndMocks: func(t *testing.T, proj *Project, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				if proj.state.CurrentPhase != session.Define {
					t.Errorf("Project state CurrentPhase = %s, want define", proj.state.CurrentPhase)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockSession, mockFeature, mockRules, mockPhase := newTestProjectWithMocks(t, ctrl)

			// Set up all mocks and potentially initial proj.state for this test case
			if tt.setupMocksAndState != nil {
				tt.setupMocksAndState(proj, mockFS, mockSession, mockFeature, mockRules, mockPhase)
			}

			result, err := proj.ChangePhase(tt.args.ctx, tt.args.targetPhase)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ChangePhase() error = nil, wantErr %v", tt.wantErr)
				} else if tt.name == "project not initialized" && !errors.Is(err, ErrNotInitialized) {
					t.Errorf("ChangePhase() error = %v, want specific error %v", err, ErrNotInitialized)
				} else if tt.name == "no active feature" && !errors.Is(err, ErrNoActiveFeature) {
					t.Errorf("ChangePhase() error = %v, want specific error %v", err, ErrNoActiveFeature)
				}
				// Add checks for other specific errors if needed
			} else if err != nil {
				t.Errorf("ChangePhase() unexpected error = %v", err)
			}

			if !tt.wantErr && result != nil && tt.wantResultMsgContains != "" {
				if !strings.Contains(result.FormatCLI(), tt.wantResultMsgContains) {
					t.Errorf("ChangePhase() result message = '%s', want to contain '%s'", result.FormatCLI(), tt.wantResultMsgContains)
				}
			}

			if !tt.wantErr && tt.verifyProjectStateAndMocks != nil {
				tt.verifyProjectStateAndMocks(t, proj, mockSession, mockRules, mockPhase)
			}
		})
	}
}

// --- New Test for EnterFeature ---

func TestProject_EnterFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name string
		args args
		// Single function to set up all mocks for the test case
		setupMocks                 func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer)
		wantErr                    bool
		wantResultMsgContains      string
		verifyProjectStateAndMocks func(t *testing.T, proj *Project)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				// Expect Stat from EnterFeature -> RequiresInitialized -> IsInitialized
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true, // Expect ErrNotInitialized
		},
		{
			name: "featureSvc.GetFeaturePhase fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.None, fmt.Errorf("phase read error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "session.SaveActiveFeature fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("test-feature").Return(fmt.Errorf("save session failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rules.RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("test-feature").Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("test-feature", session.Design.String()).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful feature enter",
			args: args{ctx: context.Background(), featureName: "existing-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "existing-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().SaveActiveFeature("existing-feature").Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("existing-feature", session.Design.String()).Return(nil).Times(1)
			},
			wantErr:               false,
			wantResultMsgContains: "Entered feature 'existing-feature' in phase 'design'.",
			verifyProjectStateAndMocks: func(t *testing.T, proj *Project) {
				if proj.state.CurrentFeature != "existing-feature" {
					t.Errorf("Project state CurrentFeature = %s, want existing-feature", proj.state.CurrentFeature)
				}
				if proj.state.CurrentPhase != session.Design {
					t.Errorf("Project state CurrentPhase = %s, want design", proj.state.CurrentPhase)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockSession, mockFeature, mockRules, _ := newTestProjectWithMocks(t, ctrl)

			// Set up all mocks for this test case
			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockFeature, mockRules)
			}

			result, err := proj.EnterFeature(tt.args.ctx, tt.args.featureName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EnterFeature() error = nil, wantErr %v", tt.wantErr)
				} else if tt.name == "project not initialized" && !errors.Is(err, ErrNotInitialized) {
					t.Errorf("EnterFeature() error = %v, want specific error %v", err, ErrNotInitialized)
				}
				// Add checks for other specific errors if needed
			} else if err != nil {
				t.Errorf("EnterFeature() unexpected error = %v", err)
			}

			if !tt.wantErr && result != nil && tt.wantResultMsgContains != "" {
				if !strings.Contains(result.Message, tt.wantResultMsgContains) { // Check against Result.Message
					t.Errorf("EnterFeature() result message = '%s', want to contain '%s'", result.Message, tt.wantResultMsgContains)
				}
			}

			if !tt.wantErr && tt.verifyProjectStateAndMocks != nil {
				tt.verifyProjectStateAndMocks(t, proj)
			}
		})
	}
}
