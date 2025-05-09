package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/phase"
	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	"github.com/imcclaskey/d3/internal/testutil"
)

// Helper to create a default project with gomocks for testing
// This function now returns all the mocks it creates so they can be used for setting expectations.
func newTestProjectWithMocks(t *testing.T, ctrl *gomock.Controller) (*Project, *portsmocks.MockFileSystem, *MockFeatureServicer, *MockRulesServicer, *MockPhaseServicer, *MockFileOperator) {
	t.Helper()
	projectRoot := t.TempDir() // Using t.TempDir() for proper test isolation

	mockFS := portsmocks.NewMockFileSystem(ctrl)
	mockFeatureSvc := NewMockFeatureServicer(ctrl)
	mockRulesSvc := NewMockRulesServicer(ctrl)
	mockPhaseSvc := NewMockPhaseServicer(ctrl)
	mockFileOp := NewMockFileOperator(ctrl)

	proj := New(projectRoot, mockFS, mockFeatureSvc, mockRulesSvc, mockPhaseSvc, mockFileOp)
	return proj, mockFS, mockFeatureSvc, mockRulesSvc, mockPhaseSvc, mockFileOp
}

// TestProject_New tests the New function (which is now very simple)
func TestProject_New_WithGoMock(t *testing.T) {
	ctrl := gomock.NewController(t)

	projectRoot := "/testroot_new_gomock" // Use a distinct root

	mockFS := portsmocks.NewMockFileSystem(ctrl)
	mockFeatureSvc := NewMockFeatureServicer(ctrl)
	mockRulesSvc := NewMockRulesServicer(ctrl)
	mockPhaseSvc := NewMockPhaseServicer(ctrl)
	mockFileOp := NewMockFileOperator(ctrl)

	proj := New(projectRoot, mockFS, mockFeatureSvc, mockRulesSvc, mockPhaseSvc, mockFileOp)

	if proj == nil {
		t.Fatal("New() returned nil")
	}
	if proj.fs != mockFS {
		t.Errorf("Expected fs to be the mockFS")
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
	if proj.fileOp != mockFileOp {
		t.Errorf("Expected fileOp to be mockFileOp")
	}
	if proj.state.ProjectRoot != projectRoot {
		t.Errorf("Expected ProjectRoot to be '%s', got '%s'", projectRoot, proj.state.ProjectRoot)
	}
	expectedD3Dir := filepath.Join(projectRoot, ".d3")
	if proj.state.D3Dir != expectedD3Dir {
		t.Errorf("Expected D3Dir to be '%s', got '%s'", expectedD3Dir, proj.state.D3Dir)
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
			mockStat:     statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil},
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
			d3DirForCheck := proj.state.D3Dir
			mockFS.EXPECT().Stat(d3DirForCheck).Return(tt.mockStat.info, tt.mockStat.err).Times(1)
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
			mockStat:    statResult{info: testutil.MockFileInfo{FIsDir: true, FName: ".d3"}, err: nil},
			expectError: nil,
		},
		{
			name:        "stat returns other error",
			mockStat:    statResult{info: nil, err: fmt.Errorf("some stat error")},
			expectError: ErrNotInitialized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, _, _, _, _ := newTestProjectWithMocks(t, ctrl)
			d3DirForCheck := proj.state.D3Dir
			mockFS.EXPECT().Stat(d3DirForCheck).Return(tt.mockStat.info, tt.mockStat.err).Times(1)
			gotError := proj.RequiresInitialized()
			if !errors.Is(gotError, tt.expectError) {
				t.Errorf("RequiresInitialized() error = %v, wantErr %v", gotError, tt.expectError)
			}
		})
	}
}

func TestProject_Init(t *testing.T) {
	type args struct {
		clean   bool
		refresh bool
	}
	tests := []struct {
		name          string
		args          args
		setupMocks    func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer)
		wantErr       bool
		wantResultMsg string
		// verifyProjectState is removed as in-memory state for feature/phase is gone
	}{
		{
			name: "standard init on existing project (no flags)",
			args: args{clean: false, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
			},
			wantErr:       false,
			wantResultMsg: "Project already initialized. Use --refresh to update or --clean to reset.",
		},
		{
			name: "standard init on new project (no flags)",
			args: args{clean: false, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", string(phase.None)).Return(nil).Times(1) // Ensure phase.None is string
				mockFeature.EXPECT().ClearActiveFeature().Return(nil).Times(1)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully. Cursor rules have been updated.",
		},
		{
			name: "initialized, clean init",
			args: args{clean: true, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(proj.state.D3Dir).Return(nil).Times(1)
				// Standard init steps after cleanups
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", string(phase.None)).Return(nil).Times(1)
				mockFeature.EXPECT().ClearActiveFeature().Return(nil).Times(1) // This is the one that is actually called when (performedClean || !originalIsCurrentlyInitialized)
			},
			wantErr:       false,
			wantResultMsg: "Project cleaned and re-initialized successfully. Cursor rules have been updated.",
		},
		{
			name: "refresh on new project",
			args: args{clean: false, refresh: true},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1) // Determines originalIsCurrentlyInitialized = false
				gomock.InOrder(
					// Standard init steps first
					mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1),
					mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1),
					mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1),

					// Refresh specific calls
					mockFeature.EXPECT().GetActiveFeature().Return("", nil).Times(1),
					mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "").Return(phase.None, nil).Times(1),
					mockRules.EXPECT().RefreshRules("", string(phase.None)).Return(nil).Times(1),

					// Conditional call due to !originalIsCurrentlyInitialized
					mockFeature.EXPECT().ClearActiveFeature().Return(nil).Times(1),
				)
			},
			wantErr:       false,
			wantResultMsg: "Project initialized successfully (refresh on non-existent project). Cursor rules have been updated.",
		},
		{
			name: "refresh on existing project",
			args: args{clean: false, refresh: true},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // Determines originalIsCurrentlyInitialized = true
				gomock.InOrder(
					// Standard init steps first (these still run)
					mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1),
					mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1),
					mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1),
					mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1),

					// Refresh specific calls -  Order corrected here
					mockFeature.EXPECT().GetActiveFeature().Return("active-feature", nil).Times(1),
					mockFeature.EXPECT().GetFeaturePhase(gomock.Any(), "active-feature").Return(phase.Define, nil).Times(1),
					mockRules.EXPECT().RefreshRules("active-feature", string(phase.Define)).Return(nil).Times(1),
				)
				// No mockFeature.ClearActiveFeature() during refresh of an existing project (originalIsCurrentlyInitialized=true, performedClean=false)
			},
			wantErr:       false,
			wantResultMsg: "Project refreshed successfully. Cursor rules have been updated.",
		},
		// ... other error cases for Init remain largely the same, ensuring RefreshRules("", string(phase.None)) is used ...
		{
			name: "error on RemoveAll during clean init",
			args: args{clean: true, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(proj.state.D3Dir).Return(fmt.Errorf("failed to remove")).Times(1)
				// mockFeature.EXPECT().ClearActiveFeature().Return(nil).AnyTimes() // This call is inside the if originalIsInitialized for clean
			},
			wantErr: true,
		},
		{
			name: "error on MkdirAll for .d3",
			args: args{clean: false, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(fmt.Errorf("failed to mkdir .d3")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on featureSvc.ClearActiveFeature",
			args: args{clean: false, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", string(phase.None)).Return(nil).Times(1)
				mockFeature.EXPECT().ClearActiveFeature().Return(fmt.Errorf("clear active feature failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "error on rules.RefreshRules",
			args: args{clean: false, refresh: false},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer, mockFileOp *MockFileOperator, mockFeature *MockFeatureServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.D3Dir, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().MkdirAll(proj.state.FeaturesDir, os.FileMode(0755)).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureMCPJSON(mockFS, proj.state.ProjectRoot).Return(nil).Times(1)
				mockFileOp.EXPECT().EnsureD3GitignoreEntries(mockFS, proj.state.D3Dir, proj.state.CursorRulesDir, proj.state.ProjectRoot).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("", string(phase.None)).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, mockPhase, mockFileOp := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockRules, mockPhase, mockFileOp, mockFeature)
			}

			result, err := proj.Init(tt.args.clean, tt.args.refresh)

			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != nil && result.FormatCLI() != tt.wantResultMsg {
				t.Errorf("Init() result message = %s, want %s", result.FormatCLI(), tt.wantResultMsg)
			}
			// Removed verifyProjectState as in-memory state is gone
		})
	}
}

func TestProject_CreateFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer)
		wantErr    bool
		// verifyProjectStateAndMocks removed as in-memory state is gone
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "featureSvc.CreateFeature fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(nil, fmt.Errorf("create feature failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "featureSvc.SetActiveFeature fails, triggers cleanup",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				ctx := gomock.Any() // context.Background()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				featurePath := filepath.Join(proj.state.FeaturesDir, "test-feature")
				mockFeature.EXPECT().CreateFeature(ctx, "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: featurePath}, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("test-feature").Return(fmt.Errorf("set active failed")).Times(1)
				// Expect cleanup call
				mockFeature.EXPECT().DeleteFeature(ctx, "test-feature").Return(false, nil).Times(1) // activeContextCleared might be false, error nil for cleanup
			},
			wantErr: true,
		},
		{
			name: "rulesSvc.RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "test-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				featurePath := filepath.Join(proj.state.FeaturesDir, "test-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "test-feature").Return(&feature.FeatureInfo{Name: "test-feature", Path: featurePath}, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("test-feature").Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("test-feature", string(phase.Define)).Return(fmt.Errorf("refresh rules failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful feature creation",
			args: args{ctx: context.Background(), featureName: "new-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhase *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				featurePath := filepath.Join(proj.state.FeaturesDir, "new-feature")
				mockFeature.EXPECT().CreateFeature(gomock.Any(), "new-feature").Return(&feature.FeatureInfo{Name: "new-feature", Path: featurePath}, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("new-feature").Return(nil).Times(1)
				mockPhase.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("new-feature", string(phase.Define)).Return(nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, mockPhase, _ := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockFeature, mockRules, mockPhase)
			}

			_, err := proj.CreateFeature(tt.args.ctx, tt.args.featureName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateFeature() error = nil, wantErr %v", tt.wantErr)
				} else if tt.name == "project not initialized" && !errors.Is(err, ErrNotInitialized) {
					t.Errorf("CreateFeature() error = %v, want specific error %v", err, ErrNotInitialized)
				}
			} else if err != nil {
				t.Errorf("CreateFeature() unexpected error = %v", err)
			}
			// Removed verifyProjectStateAndMocks
		})
	}
}

func TestProject_ChangePhase(t *testing.T) {
	type args struct {
		ctx         context.Context
		targetPhase phase.Phase
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer)
		wantErr    bool
		wantMsg    string
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "no active feature",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For RequiresInitialized
				mockFeature.EXPECT().GetActiveFeature().Return("", nil).Times(1)
			},
			wantErr: true, // Expect ErrNoActiveFeature
		},
		{
			name: "GetActiveFeature fails",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("", fmt.Errorf("get active failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "GetFeaturePhase for current phase fails",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("active-feat", nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "active-feat").Return(phase.None, fmt.Errorf("get phase failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "already in target phase",
			args: args{ctx: context.Background(), targetPhase: phase.Define},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("active-feat", nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "active-feat").Return(phase.Define, nil).Times(1)
			},
			wantErr: false,
			wantMsg: "Already in the define phase.",
		},
		{
			name: "SetFeaturePhase fails",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("active-feat", nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "active-feat").Return(phase.Define, nil).Times(1)
				mockFeature.EXPECT().SetFeaturePhase(ctx, "active-feat", phase.Design).Return(fmt.Errorf("set phase failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "RefreshRules fails",
			args: args{ctx: context.Background(), targetPhase: phase.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("active-feat", nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "active-feat").Return(phase.Design, nil).Times(1)
				mockFeature.EXPECT().SetFeaturePhase(ctx, "active-feat", phase.Deliver).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("active-feat", string(phase.Deliver)).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful phase change, no impact",
			args: args{ctx: context.Background(), targetPhase: phase.Design},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				featureName := "active-feat"
				featurePath := filepath.Join(proj.state.FeaturesDir, featureName)
				targetPhaseDir := filepath.Join(featurePath, string(phase.Design))

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return(featureName, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, featureName).Return(phase.Define, nil).Times(1)
				mockFeature.EXPECT().SetFeaturePhase(ctx, featureName, phase.Design).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, string(phase.Design)).Return(nil).Times(1)
				mockPhaseSvc.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockFS.EXPECT().Stat(targetPhaseDir).Return(nil, os.ErrNotExist).Times(1) // No impact
			},
			wantErr: false,
			wantMsg: "Moved to design phase. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
		{
			name: "successful phase change, with impact",
			args: args{ctx: context.Background(), targetPhase: phase.Deliver},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer, mockPhaseSvc *MockPhaseServicer) {
				ctx := gomock.Any()
				featureName := "impact-feat"
				featurePath := filepath.Join(proj.state.FeaturesDir, featureName)
				targetPhaseDir := filepath.Join(featurePath, string(phase.Deliver))

				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return(featureName, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, featureName).Return(phase.Design, nil).Times(1)
				mockFeature.EXPECT().SetFeaturePhase(ctx, featureName, phase.Deliver).Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules(featureName, string(phase.Deliver)).Return(nil).Times(1)
				mockPhaseSvc.EXPECT().EnsurePhaseFiles(featurePath).Return(nil).Times(1)
				mockFS.EXPECT().Stat(targetPhaseDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // Has impact
			},
			wantErr: false,
			wantMsg: "Moved to deliver phase. Note: Existing files were detected for the target phase. Review required. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, mockPhaseSvc, _ := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockFeature, mockRules, mockPhaseSvc)
			}

			result, err := proj.ChangePhase(tt.args.ctx, tt.args.targetPhase)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ChangePhase() error = nil, wantErr %v", tt.wantErr)
				}
				// Specific error check for ErrNoActiveFeature
				if tt.name == "no active feature" && !errors.Is(err, ErrNoActiveFeature) {
					t.Errorf("ChangePhase() error = %v, want specific error %v", err, ErrNoActiveFeature)
				}
			} else {
				if err != nil {
					t.Errorf("ChangePhase() unexpected error = %v", err)
				}
				if result.FormatMCP() != tt.wantMsg {
					t.Errorf("ChangePhase() result msg = %q, want %q", result.FormatMCP(), tt.wantMsg)
				}
			}
		})
	}
}

func TestProject_EnterFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer)
		wantErr    bool
		wantMsg    string
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "any-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "GetFeaturePhase fails (feature might not exist or .phase file issue)",
			args: args{ctx: context.Background(), featureName: "nonexistent-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "nonexistent-feature").Return(phase.None, fmt.Errorf("phase file not found or feature missing")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "SetActiveFeature fails",
			args: args{ctx: context.Background(), featureName: "feat-set-active-fail"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "feat-set-active-fail").Return(phase.Design, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("feat-set-active-fail").Return(fmt.Errorf("set active failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "RefreshRules fails",
			args: args{ctx: context.Background(), featureName: "feat-rules-fail"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "feat-rules-fail").Return(phase.Define, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("feat-rules-fail").Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("feat-rules-fail", string(phase.Define)).Return(fmt.Errorf("rules refresh failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful enter feature",
			args: args{ctx: context.Background(), featureName: "my-feature"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetFeaturePhase(ctx, "my-feature").Return(phase.Design, nil).Times(1)
				mockFeature.EXPECT().SetActiveFeature("my-feature").Return(nil).Times(1)
				mockRules.EXPECT().RefreshRules("my-feature", string(phase.Design)).Return(nil).Times(1)
			},
			wantErr: false,
			wantMsg: "Entered feature 'my-feature' in phase 'design'. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, _, _ := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockFeature, mockRules)
			}

			result, err := proj.EnterFeature(tt.args.ctx, tt.args.featureName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EnterFeature() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("EnterFeature() unexpected error = %v", err)
				}
				if result.FormatMCP() != tt.wantMsg {
					t.Errorf("EnterFeature() result msg = %q, want %q", result.FormatMCP(), tt.wantMsg)
				}
			}
			// Removed checks for proj.state.CurrentFeature and proj.state.CurrentPhase
		})
	}
}

func TestProject_ExitFeature(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer)
		wantErr    bool
		wantMsg    string
	}{
		{
			name: "project not initialized",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "no active feature to exit (GetActiveFeature returns empty)",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("", nil).Times(1)
				mockRules.EXPECT().ClearGeneratedRules().Return(nil).Times(1) // Still attempt to clear rules
			},
			wantErr: false,
			wantMsg: "No active feature to exit. Cursor rules cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
		{
			name: "GetActiveFeature fails during exit",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("", fmt.Errorf("read active failed")).Times(1) // Error, but name is empty
				mockRules.EXPECT().ClearGeneratedRules().Return(nil).Times(1)                                 // Should still proceed to clear rules
			},
			wantErr: false, // Error is logged as warning, main operation proceeds as "no active feature"
			wantMsg: "No active feature to exit. Cursor rules cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
		{
			name: "ClearActiveFeature fails",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("some-active-feature", nil).Times(1)
				mockFeature.EXPECT().ClearActiveFeature().Return(fmt.Errorf("clear active failed")).Times(1)
				mockRules.EXPECT().ClearGeneratedRules().Return(nil).AnyTimes() // Added expectation as it's called even if ClearActiveFeature fails
			},
			wantErr: true,
		},
		{
			name: "ClearGeneratedRules fails (logged as warning)",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("exited-feature", nil).Times(1)
				mockFeature.EXPECT().ClearActiveFeature().Return(nil).Times(1)
				mockRules.EXPECT().ClearGeneratedRules().Return(fmt.Errorf("rules clear failed")).Times(1)
			},
			wantErr: false, // Error is warning
			wantMsg: "Exited feature 'exited-feature'. No active feature. Cursor rules cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
		{
			name: "successful exit",
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().GetActiveFeature().Return("feature-to-exit", nil).Times(1)
				mockFeature.EXPECT().ClearActiveFeature().Return(nil).Times(1)
				mockRules.EXPECT().ClearGeneratedRules().Return(nil).Times(1)
			},
			wantErr: false,
			wantMsg: "Exited feature 'feature-to-exit'. No active feature. Cursor rules cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, _, _ := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockFeature, mockRules)
			}

			result, err := proj.ExitFeature(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExitFeature() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("ExitFeature() unexpected error = %v", err)
				}
				if result.FormatMCP() != tt.wantMsg {
					t.Errorf("ExitFeature() result msg = %q, want %q", result.FormatMCP(), tt.wantMsg)
				}
			}
			// Removed checks for proj.state.CurrentFeature and proj.state.CurrentPhase
		})
	}
}

func TestProject_DeleteFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer)
		wantErr    bool
		wantMsg    string
	}{
		{
			name: "project not initialized",
			args: args{ctx: context.Background(), featureName: "any-feat"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name: "DeleteFeature service fails",
			args: args{ctx: context.Background(), featureName: "del-feat"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().DeleteFeature(ctx, "del-feat").Return(false, fmt.Errorf("svc delete failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "successful delete, not active, no rules impact",
			args: args{ctx: context.Background(), featureName: "del-feat1"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().DeleteFeature(ctx, "del-feat1").Return(false, nil).Times(1) // activeContextCleared = false
			},
			wantErr: false,
			wantMsg: "Feature 'del-feat1' deleted successfully.",
		},
		{
			name: "successful delete, was active, rules cleared",
			args: args{ctx: context.Background(), featureName: "active-del-feat"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().DeleteFeature(ctx, "active-del-feat").Return(true, nil).Times(1) // activeContextCleared = true
				mockRules.EXPECT().ClearGeneratedRules().Return(nil).Times(1)
			},
			wantErr: false,
			wantMsg: "Feature 'active-del-feat' deleted successfully. Active feature context has been cleared. Cursor rules have changed. Stop your current behavior and await further instruction.",
		},
		{
			name: "successful delete, was active, ClearGeneratedRules fails (warning)",
			args: args{ctx: context.Background(), featureName: "active-del-rules-fail"},
			setupMocks: func(proj *Project, mockFS *portsmocks.MockFileSystem, mockFeature *MockFeatureServicer, mockRules *MockRulesServicer) {
				ctx := gomock.Any()
				mockFS.EXPECT().Stat(proj.state.D3Dir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFeature.EXPECT().DeleteFeature(ctx, "active-del-rules-fail").Return(true, nil).Times(1) // activeContextCleared = true
				mockRules.EXPECT().ClearGeneratedRules().Return(fmt.Errorf("rules clear failed")).Times(1)
			},
			wantErr: false,
			wantMsg: "Feature 'active-del-rules-fail' deleted successfully. Warning: failed to clear rules after deleting active feature: rules clear failed Active feature context has been cleared.", // Note: No RulesChanged in MCP message if ClearGeneratedRules fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			proj, mockFS, mockFeature, mockRules, _, _ := newTestProjectWithMocks(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(proj, mockFS, mockFeature, mockRules)
			}

			result, err := proj.DeleteFeature(tt.args.ctx, tt.args.featureName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DeleteFeature() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("DeleteFeature() unexpected error = %v", err)
				}
				if result.FormatMCP() != tt.wantMsg {
					t.Errorf("DeleteFeature() result msg = %q, want %q", result.FormatMCP(), tt.wantMsg)
				}
			}
			// Removed checks for proj.state.CurrentFeature and proj.state.CurrentPhase
		})
	}
}
