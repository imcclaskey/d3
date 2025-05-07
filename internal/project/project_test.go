package project

import (
	"context"
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

	proj := New(projectRoot, mockFS, mockSessionSvc, mockFeatureSvc, mockRulesSvc, mockPhaseSvc)
	return proj, mockFS, mockSessionSvc, mockFeatureSvc, mockRulesSvc, mockPhaseSvc
}

// TestProject_New tests the New function with gomocks
func TestProject_New_WithGoMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	// No need to call ctrl.Finish() if it's the only controller and the test ends.
	// It's good practice if you have multiple controllers or more complex test lifecycles.

	projectRoot := "/testroot_gomock" // We can use a specific root for this test if preferred, or t.TempDir()

	mockFS := portsmocks.NewMockFileSystem(ctrl)
	mockSessionSvc := NewMockStorageService(ctrl)
	mockFeatureSvc := NewMockFeatureServicer(ctrl)
	mockRulesSvc := NewMockRulesServicer(ctrl)
	mockPhaseSvc := NewMockPhaseServicer(ctrl)

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
	if proj.isInitialized {
		t.Error("Expected isInitialized to be false initially")
	}
}

func TestProject_IsInitialized(t *testing.T) {
	type statResult struct {
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name         string
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
			mockStat:     statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil}, // Use testutil.MockFileInfo
			expectIsInit: true,
		},
		{
			name:         "stat returns other error",
			mockStat:     statResult{info: nil, err: fmt.Errorf("some stat error")},
			expectIsInit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			proj, mockFS, _, _, _, _ := newTestProjectWithMocks(t, ctrl)

			d3DirForTest := proj.state.D3Dir

			mockFS.EXPECT().Stat(d3DirForTest).Return(tt.mockStat.info, tt.mockStat.err).Times(1)

			if got := proj.IsInitialized(); got != tt.expectIsInit {
				t.Errorf("IsInitialized() = %v, want %v", got, tt.expectIsInit)
			}
		})
	}
}

func TestProject_RequiresInitialized(t *testing.T) {
	type statResult struct {
		info os.FileInfo
		err  error
	}
	tests := []struct {
		name        string
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
			mockStat:    statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil}, // Use testutil.MockFileInfo
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, _, _, _, _ := newTestProjectWithMocks(t, ctrl)
			d3DirForTest := proj.state.D3Dir

			mockFS.EXPECT().Stat(d3DirForTest).Return(tt.mockStat.info, tt.mockStat.err).Times(1)

			gotError := proj.RequiresInitialized()
			if gotError != tt.expectError {
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
		name               string
		args               args
		setupMocks         func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string)
		wantErr            bool
		wantResultMsg      string
		verifyProjectState func(t *testing.T, proj *Project)
	}{
		{
			name: "already initialized, not clean",
			args: args{clean: false},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
			},
			wantErr:       false,
			wantResultMsg: "Project already initialized.",
		},
		{
			name: "not initialized, successful init",
			args: args{clean: false},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(featuresDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).DoAndReturn(func(state *session.SessionState) error {
					if state.Version != "1.0" {
						return fmt.Errorf("expected version 1.0, got %s", state.Version)
					}
					return nil
				}).Times(1)
				mockRules.EXPECT().RefreshRules("", "").Return(nil).Times(1)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully. Cursor rules have been updated.",
			verifyProjectState: func(t *testing.T, proj *Project) {
				if !proj.isInitialized {
					t.Errorf("Expected proj.isInitialized to be true after successful init")
				}
			},
		},
		{
			name: "initialized, clean init",
			args: args{clean: true},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				gomock.InOrder(
					mockFS.EXPECT().Stat(d3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(2), // Use testutil.MockFileInfo
					mockFS.EXPECT().RemoveAll(d3Dir).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(featuresDir, os.FileMode(0755)).Return(nil).Times(1),
					mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1),
					mockRules.EXPECT().RefreshRules("", "").Return(nil).Times(1),
				)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name: "error on RemoveAll during clean init",
			args: args{clean: true},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(2) // Use testutil.MockFileInfo
				mockFS.EXPECT().RemoveAll(d3Dir).Return(fmt.Errorf("failed to remove")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on MkdirAll for .d3",
			args: args{clean: false},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(fmt.Errorf("failed to mkdir .d3")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on session.Save",
			args: args{clean: false},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(featuresDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(fmt.Errorf("session save failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on rules.RefreshRules",
			args: args{clean: false},
			setupMocks: func(mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, d3Dir string, featuresDir string) {
				mockFS.EXPECT().Stat(d3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(d3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(featuresDir, os.FileMode(0755)).Return(nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", "").Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			proj, mockFS, mockSession, mockFeature, mockRules, _ := newTestProjectWithMocks(t, ctrl)
			_ = mockFeature

			d3Dir := proj.state.D3Dir
			featuresDir := proj.state.FeaturesDir

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, mockSession, mockRules, d3Dir, featuresDir)
			}

			result, err := proj.Init(tt.args.clean)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.FormatCLI() != tt.wantResultMsg {
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
		name                       string
		args                       args
		setupMocks                 func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr                    bool
		verifyProjectStateAndMocks func(t *testing.T, proj *Project, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "featureSvc.CreateFeature fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // Use testutil.MockFileInfo
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(nil, fmt.Errorf("create feature failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "sessionSvc.Load fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // Use testutil.MockFileInfo
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature"}, nil).Times(1)
				mockSession.EXPECT().Load().Return(nil, fmt.Errorf("load session failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "sessionSvc.Save fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect CreateFeature call to feature service
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: filepath.Join(proj.state.FeaturesDir, "test-feature")}, nil).Times(1)
				mockSession.EXPECT().Load().Return(&session.SessionState{}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).DoAndReturn(func(state *session.SessionState) error {
					// Verify the saved state
					if state.CurrentFeature != "test-feature" {
						return fmt.Errorf("expected saved feature test-feature, got %s", state.CurrentFeature)
					}
					// CurrentPhase is no longer in SessionState, so no check needed here
					return fmt.Errorf("save session failed")
				}).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rulesSvc.RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				featurePath := filepath.Join(proj.state.FeaturesDir, "test-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: featurePath}, nil).Times(1)
				mockSession.EXPECT().Load().Return(&session.SessionState{}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				// RefreshRules should be called with the feature name and the default phase (Define)
				mockRules.EXPECT().RefreshRules("test-feature", session.Define.String()).Return(fmt.Errorf("refresh rules failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful feature creation",
			args: args{ctx: context.Background(), featureName: "new-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				featurePath := filepath.Join(proj.state.FeaturesDir, "new-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "new-feature").Return(&feature.FeatureInfo{Name: "new-feature", Path: featurePath}, nil).Times(1)
				// Load might return old state, CreateFeature logic should overwrite CurrentFeature
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: "old-feature"}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).DoAndReturn(func(state *session.SessionState) error {
					if state.CurrentFeature != "new-feature" {
						return fmt.Errorf("expected saved feature new-feature, got %s", state.CurrentFeature)
					}
					// CurrentPhase is no longer in SessionState
					return nil
				}).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				// RefreshRules should be called with the new feature and Define phase
				mockRules.EXPECT().RefreshRules("new-feature", session.Define.String()).Return(nil).Times(1)
			},
			wantErr: false,
			verifyProjectStateAndMocks: func(t *testing.T, proj *Project, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				if proj.state.CurrentFeature != "new-feature" {
					t.Errorf("Project state CurrentFeature = %s, want new-feature", proj.state.CurrentFeature)
				}
				// Verify in-memory phase is set correctly
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

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockFeature, mockRules, mockPhase)
			}

			_, err := proj.CreateFeature(tt.args.ctx, tt.args.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFeature() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.verifyProjectStateAndMocks != nil {
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
		name                       string
		args                       args
		setupMocks                 func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr                    bool
		wantResultMsgContains      string // Check if result message contains this substring
		verifyProjectStateAndMocks func(t *testing.T, proj *Project, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "session.Load fails",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Set in-memory state to ensure the initial check passes
				proj.state.CurrentFeature = "feat-load-fail"
				proj.state.CurrentPhase = session.Define
				// Now expect SetFeaturePhase to be called (it happens before session load)
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl)
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), "feat-load-fail", session.Design).Return(nil).Times(1)
				proj.features = mockFeature
				// Expect Load to fail
				mockSession.EXPECT().Load().Return(nil, fmt.Errorf("session load failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "no active feature",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Ensure in-memory state has no active feature
				proj.state.CurrentFeature = ""
				proj.state.CurrentPhase = session.None
				// No other mocks should be called as it returns early
			},
			wantErr: true, // Expect ErrNoActiveFeature
		},
		{
			name: "already in target phase",
			args: args{ctx: context.Background(), targetPhase: session.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Set initial in-memory state for the test
				proj.state.CurrentFeature = "feat1"
				proj.state.CurrentPhase = session.Design
			},
			wantErr:               false,
			wantResultMsgContains: "Already in the design phase.",
		},
		{
			name: "featureSvc.SetFeaturePhase fails", // Renamed test case
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				featureName := "feat1"
				initialPhase := session.Design
				proj.state.CurrentFeature = featureName // Set initial state
				proj.state.CurrentPhase = initialPhase

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect call to SetFeaturePhase on the feature service
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl) // Need mockFeature instance
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(fmt.Errorf("set phase failed")).Times(1)
				proj.features = mockFeature // Assign mock back if recreated
			},
			wantErr: true,
		},
		{
			name: "session.Save fails after phase change",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				featureName := "feat1"
				initialPhase := session.Design
				proj.state.CurrentFeature = featureName // Set initial state
				proj.state.CurrentPhase = initialPhase

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect SetFeaturePhase to succeed
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl)
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(nil).Times(1)
				proj.features = mockFeature
				// Expect session Load and Save, but Save fails
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: featureName}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(fmt.Errorf("session save failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rules.RefreshRules fails after phase change",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				featureName := "feat1"
				initialPhase := session.Design
				proj.state.CurrentFeature = featureName // Set initial state
				proj.state.CurrentPhase = initialPhase

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect SetFeaturePhase and Session Load/Save to succeed
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl)
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(nil).Times(1)
				proj.features = mockFeature
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: featureName}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				// Expect Rules Refresh to fail
				mockRules.EXPECT().RefreshRules(featureName, session.Deliver.String()).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful phase change, no existing phase dir",
			args: args{ctx: context.Background(), targetPhase: session.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				featureName := "feat1"
				initialPhase := session.Design
				proj.state.CurrentFeature = featureName // Set initial state
				proj.state.CurrentPhase = initialPhase
				featureDir := filepath.Join(proj.state.FeaturesDir, featureName)
				phaseToCheckDir := filepath.Join(featureDir, session.Deliver.String())

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect SetFeaturePhase, Session Load/Save, Rules Refresh, EnsurePhaseFiles to succeed
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl)
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Deliver).Return(nil).Times(1)
				proj.features = mockFeature
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: featureName}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, session.Deliver.String()).Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featureDir).Return(nil).Times(1)
				// Stat for impact check
				mockFS.EXPECT().Stat(phaseToCheckDir).Return(nil, os.ErrNotExist).Times(1)
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
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				featureName := "featImpact"
				initialPhase := session.Design
				proj.state.CurrentFeature = featureName // Set initial state
				proj.state.CurrentPhase = initialPhase
				featureDir := filepath.Join(proj.state.FeaturesDir, featureName)
				phaseToCheckDir := filepath.Join(featureDir, session.Define.String())

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// Expect SetFeaturePhase, Session Load/Save, Rules Refresh, EnsurePhaseFiles to succeed
				mockFeature := NewMockFeatureServicer(proj.features.(*MockFeatureServicer).ctrl)
				mockFeature.EXPECT().SetFeaturePhase(gomock.Any(), featureName, session.Define).Return(nil).Times(1)
				proj.features = mockFeature
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: featureName}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, session.Define.String()).Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featureDir).Return(nil).Times(1)
				// Stat for impact check (returns existing dir)
				mockFS.EXPECT().Stat(phaseToCheckDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
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
			proj, mockFS, mockSession, mockFeature, mockRules, mockPhase := newTestProjectWithMocks(t, ctrl) // featureSvc not used by ChangePhase
			_ = mockFeature                                                                                  // Explicitly ignore mockFeature as it's not used

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockRules, mockPhase)
			}

			result, err := proj.ChangePhase(tt.args.ctx, tt.args.targetPhase)

			if (err != nil) != tt.wantErr {
				t.Errorf("ChangePhase() error = %v, wantErr %v", err, tt.wantErr)
				// If we expect an error and get one, or expect no error and get none, but it's the WRONG error/result, check below.
				// If we wanted an error and got nil, or wanted nil and got an error, this check is enough.
				// For specific error types, like ErrNoActiveFeature, an additional check is needed if wantErr is true.
				if tt.name == "no active feature" && err != ErrNoActiveFeature {
					t.Errorf("ChangePhase() error = %v, want specific error %v", err, ErrNoActiveFeature)
				}
				return // Important to return if error expectation mismatch, to avoid nil pointer on result.
			}

			if !tt.wantErr && result != nil && tt.wantResultMsgContains != "" {
				if !strings.Contains(result.FormatCLI(), tt.wantResultMsgContains) {
					t.Errorf("ChangePhase() result message = '%s', want to contain '%s'", result.FormatCLI(), tt.wantResultMsgContains)
				}
			}

			if tt.verifyProjectStateAndMocks != nil {
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
		name                       string
		args                       args
		setupMocks                 func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer)
		wantErr                    bool
		wantResultMsgContains      string
		verifyProjectStateAndMocks func(t *testing.T, proj *Project)
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "featureSvc.GetFeaturePhase fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.None, fmt.Errorf("phase read error")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "session.Load fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().Load().Return(nil, fmt.Errorf("session load failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "session.Save fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().Load().Return(&session.SessionState{}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).DoAndReturn(func(state *session.SessionState) error {
					if state.CurrentFeature != "test-feature" {
						return fmt.Errorf("expected saved feature test-feature, got %s", state.CurrentFeature)
					}
					return fmt.Errorf("save session failed")
				}).Times(1)
			},
			wantErr: true,
		},
		{
			name: "rules.RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "test-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().Load().Return(&session.SessionState{}, nil).Times(1)
				mockSession.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("test-feature", session.Design.String()).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful feature enter",
			args: args{ctx: context.Background(), featureName: "existing-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockSession *MockStorageService, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "existing-feature").Return(session.Design, nil).Times(1)
				mockSession.EXPECT().Load().Return(&session.SessionState{CurrentFeature: "other"}, nil).Times(1) // Load existing session
				mockSession.EXPECT().Save(gomock.Any()).DoAndReturn(func(state *session.SessionState) error {
					if state.CurrentFeature != "existing-feature" {
						return fmt.Errorf("expected saved feature existing-feature, got %s", state.CurrentFeature)
					}
					return nil
				}).Times(1)
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

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockSession, mockFeature, mockRules)
			}

			result, err := proj.EnterFeature(tt.args.ctx, tt.args.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnterFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil && tt.wantResultMsgContains != "" {
				if !strings.Contains(result.Message, tt.wantResultMsgContains) { // Check against Result.Message
					t.Errorf("EnterFeature() result message = '%s', want to contain '%s'", result.Message, tt.wantResultMsgContains)
				}
			}

			if tt.verifyProjectStateAndMocks != nil {
				tt.verifyProjectStateAndMocks(t, proj)
			}
		})
	}
}
