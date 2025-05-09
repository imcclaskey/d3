package feature

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/imcclaskey/d3/internal/core/phase"
	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	"github.com/imcclaskey/d3/internal/testutil"
	"gopkg.in/yaml.v3"
)

// Helper to create a new service with mock FS for testing
func newTestService(t *testing.T, ctrl *gomock.Controller) (*Service, *portsmocks.MockFileSystem) {
	t.Helper()
	projectRoot := t.TempDir()
	featuresDir := filepath.Join(projectRoot, "features")
	d3Dir := filepath.Join(projectRoot, ".d3") // Assuming d3Dir path is needed for context, though not used by feature service directly
	mockFS := portsmocks.NewMockFileSystem(ctrl)
	s := NewService(projectRoot, featuresDir, d3Dir, mockFS)
	return s, mockFS
}

func TestService_CreateFeature(t *testing.T) {
	type args struct {
		ctx         context.Context
		featureName string
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string)
		wantInfo   *FeatureInfo
		wantErr    bool
	}{
		{
			name: "feature already exists (Stat returns nil error)",
			args: args{ctx: context.Background(), featureName: "existing-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FName: featureName, FIsDir: true}, nil).Times(1)
			},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "Stat returns unexpected error",
			args: args{ctx: context.Background(), featureName: "error-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, fmt.Errorf("some stat error")).Times(1)
			},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "MkdirAll fails",
			args: args{ctx: context.Background(), featureName: "mkdir-fail-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantInfo: nil,
			wantErr:  true,
		},
		{
			name: "successful creation",
			args: args{ctx: context.Background(), featureName: "new-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				stateFilePath := filepath.Join(featurePath, stateFileName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), os.FileMode(0644)).Return(nil).Times(1)
			},
			wantInfo: &FeatureInfo{Name: "new-feature", Path: ""}, // Path will be dynamic based on temp dir
			wantErr:  false,
		},
		{
			name: "WriteFile for state.yaml fails",
			args: args{ctx: context.Background(), featureName: "state-fail-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				stateFilePath := filepath.Join(featurePath, stateFileName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), os.FileMode(0644)).Return(fmt.Errorf("write state failed")).Times(1)
			},
			wantInfo: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(s, mockFS, tt.args.featureName)
			}

			gotInfo, err := s.CreateFeature(tt.args.ctx, tt.args.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.CreateFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotInfo == nil || gotInfo.Name != tt.wantInfo.Name {
					t.Errorf("Service.CreateFeature() gotInfo.Name = %v, want %v", gotInfo.Name, tt.wantInfo.Name)
				}
				// Check if path has the correct base name
				expectedPathSuffix := filepath.Join("features", tt.args.featureName)
				if gotInfo == nil || !strings.HasSuffix(gotInfo.Path, expectedPathSuffix) {
					t.Errorf("Service.CreateFeature() gotInfo.Path = %q, want suffix %q", gotInfo.Path, expectedPathSuffix)
				}
			}
		})
	}
}

func TestService_FeatureExists(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		setupMocks  func(mockFS *portsmocks.MockFileSystem, path string)
		wantExists  bool
	}{
		{
			name:        "feature exists",
			featureName: "feat1",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, path string) {
				mockFS.EXPECT().Stat(path).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
			},
			wantExists: true,
		},
		{
			name:        "feature does not exist",
			featureName: "feat2",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, path string) {
				mockFS.EXPECT().Stat(path).Return(nil, os.ErrNotExist).Times(1)
			},
			wantExists: false,
		},
		{
			name:        "stat returns other error",
			featureName: "feat3",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, path string) {
				mockFS.EXPECT().Stat(path).Return(nil, fmt.Errorf("stat failed")).Times(1)
			},
			wantExists: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)
			featurePath := filepath.Join(s.featuresDir, tt.featureName)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, featurePath)
			}

			if gotExists := s.FeatureExists(tt.featureName); gotExists != tt.wantExists {
				t.Errorf("Service.FeatureExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func TestService_GetFeaturePath(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, _ := newTestService(t, ctrl)
	featureName := "my-special-feature"
	want := filepath.Join(s.featuresDir, featureName)
	got := s.GetFeaturePath(featureName)
	if got != want {
		t.Errorf("GetFeaturePath(%q) = %q, want %q", featureName, got, want)
	}
}

// MockDirEntry for ListFeatures test
type MockDirEntry struct {
	EntryName  string
	EntryIsDir bool
}

func (m MockDirEntry) Name() string { return m.EntryName }
func (m MockDirEntry) IsDir() bool  { return m.EntryIsDir }
func (m MockDirEntry) Type() fs.FileMode {
	if m.EntryIsDir {
		return fs.ModeDir
	}
	return 0
}
func (m MockDirEntry) Info() (fs.FileInfo, error) {
	return testutil.MockFileInfo{FName: m.EntryName, FIsDir: m.EntryIsDir}, nil
}

func TestService_ListFeatures(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem, featuresDir string)
		wantCount  int
		wantNames  []string // Expected names in the returned list
		wantErr    bool
	}{
		{
			name: "features directory does not exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featuresDir string) {
				mockFS.EXPECT().Stat(featuresDir).Return(nil, os.ErrNotExist).Times(1)
			},
			wantCount: 0,
			wantNames: []string{},
			wantErr:   false, // Returns empty list, not error
		},
		{
			name: "stat features directory fails",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featuresDir string) {
				mockFS.EXPECT().Stat(featuresDir).Return(nil, fmt.Errorf("stat failed")).Times(1)
			},
			wantCount: 0,
			wantNames: nil,
			wantErr:   true,
		},
		{
			name: "ReadDir fails",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featuresDir string) {
				mockFS.EXPECT().Stat(featuresDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadDir(featuresDir).Return(nil, fmt.Errorf("readdir failed")).Times(1)
			},
			wantCount: 0,
			wantNames: nil,
			wantErr:   true,
		},
		{
			name: "successful list with dirs and files",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featuresDir string) {
				dirEntries := []fs.DirEntry{
					MockDirEntry{EntryName: "feat1", EntryIsDir: true},
					MockDirEntry{EntryName: "a_file.txt", EntryIsDir: false},
					MockDirEntry{EntryName: "feat2", EntryIsDir: true},
				}
				mockFS.EXPECT().Stat(featuresDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadDir(featuresDir).Return(dirEntries, nil).Times(1)
			},
			wantCount: 2,
			wantNames: []string{"feat1", "feat2"},
			wantErr:   false,
		},
		{
			name: "successful list with no entries",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featuresDir string) {
				dirEntries := []fs.DirEntry{}
				mockFS.EXPECT().Stat(featuresDir).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadDir(featuresDir).Return(dirEntries, nil).Times(1)
			},
			wantCount: 0,
			wantNames: []string{},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, s.featuresDir)
			}

			gotFeatures, err := s.ListFeatures(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ListFeatures() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(gotFeatures) != tt.wantCount {
				t.Errorf("Service.ListFeatures() got %d features, want %d", len(gotFeatures), tt.wantCount)
			}
			if !tt.wantErr && tt.wantNames != nil { // Only check names if no error and expected names provided
				gotNames := make([]string, len(gotFeatures))
				for i, f := range gotFeatures {
					gotNames[i] = f.Name
				}
				// Simple comparison (doesn't require order match)
				if len(gotNames) != len(tt.wantNames) {
					t.Errorf("Service.ListFeatures() names length mismatch: got %v, want %v", gotNames, tt.wantNames)
				} else {
					// Check if all wanted names are present (order independent check)
					gotMap := make(map[string]bool)
					for _, name := range gotNames {
						gotMap[name] = true
					}
					for _, wantName := range tt.wantNames {
						if !gotMap[wantName] {
							t.Errorf("Service.ListFeatures() missing expected name: %s in got %v", wantName, gotNames)
						}
					}
				}
			}
		})
	}
}

// --- New Tests for Phase Management ---

func TestService_GetFeaturePhase(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		setupMocks  func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string)
		wantPhase   phase.Phase
		wantErr     bool
	}{
		{
			name:        "state file exists, valid phase",
			featureName: "feat1",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				validYAML := []byte("active_phase: design")
				mockFS.EXPECT().ReadFile(stateFilePath).Return(validYAML, nil).Times(1)
			},
			wantPhase: phase.Design,
			wantErr:   false,
		},
		{
			name:        "state file exists, empty phase defaults to define",
			featureName: "feat-empty-phase",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				emptyPhaseYAML := []byte("active_phase: \"\"") // Empty string phase
				mockFS.EXPECT().ReadFile(stateFilePath).Return(emptyPhaseYAML, nil).Times(1)
			},
			wantPhase: phase.Define,
			wantErr:   false,
		},
		{
			name:        "state file exists, missing phase key defaults to define",
			featureName: "feat-missing-key",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				missingKeyYAML := []byte("other_key: value")
				mockFS.EXPECT().ReadFile(stateFilePath).Return(missingKeyYAML, nil).Times(1)
			},
			wantPhase: phase.Define,
			wantErr:   false,
		},
		{
			name:        "state file does not exist, creates default (define)",
			featureName: "feat-new",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				mockFS.EXPECT().ReadFile(stateFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				// Expect write with default phase
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), gomock.Any()).DoAndReturn(func(path string, data []byte, perm os.FileMode) error {
					var state featureStateData
					yaml.Unmarshal(data, &state)
					if state.LastActivePhase != phase.Define {
						return fmt.Errorf("expected default phase to be define, got %s", state.LastActivePhase)
					}
					return nil
				}).Times(1)
			},
			wantPhase: phase.Define,
			wantErr:   false,
		},
		{
			name:        "state file does not exist, MkdirAll fails",
			featureName: "feat-mkdir-fail",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				mockFS.EXPECT().ReadFile(stateFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "state file does not exist, WriteFile fails",
			featureName: "feat-write-fail",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				mockFS.EXPECT().ReadFile(stateFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), gomock.Any()).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "ReadFile returns other error",
			featureName: "feat-read-err",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				mockFS.EXPECT().ReadFile(stateFilePath).Return(nil, fmt.Errorf("random read error")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "Unmarshal fails",
			featureName: "feat-unmarshal-err",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string) {
				invalidYAML := []byte("active_phase: [invalid]")
				mockFS.EXPECT().ReadFile(stateFilePath).Return(invalidYAML, nil).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)
			featurePath := filepath.Join(s.featuresDir, tt.featureName)
			stateFilePath := filepath.Join(featurePath, stateFileName)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, featurePath, stateFilePath)
			}

			gotPhase, err := s.GetFeaturePhase(context.Background(), tt.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetFeaturePhase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPhase != tt.wantPhase {
				t.Errorf("Service.GetFeaturePhase() gotPhase = %v, want %v", gotPhase, tt.wantPhase)
			}
		})
	}
}

func TestService_SetFeaturePhase(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		phaseToSet  phase.Phase
		setupMocks  func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string, p phase.Phase)
		wantErr     bool
	}{
		{
			name:        "successful set phase",
			featureName: "feat-set1",
			phaseToSet:  phase.Deliver,
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string, p phase.Phase) {
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				// Expect write with the correct phase
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), gomock.Any()).DoAndReturn(func(path string, data []byte, perm os.FileMode) error {
					var state featureStateData
					yaml.Unmarshal(data, &state)
					if state.LastActivePhase != p {
						return fmt.Errorf("expected phase %s, got %s", p, state.LastActivePhase)
					}
					return nil
				}).Times(1)
			},
			wantErr: false,
		},
		{
			name:        "MkdirAll fails",
			featureName: "feat-set-mkdir-fail",
			phaseToSet:  phase.Design,
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string, p phase.Phase) {
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "WriteFile fails",
			featureName: "feat-set-write-fail",
			phaseToSet:  phase.Design,
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featurePath string, stateFilePath string, p phase.Phase) {
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(stateFilePath, gomock.Any(), gomock.Any()).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)
			featurePath := filepath.Join(s.featuresDir, tt.featureName)
			stateFilePath := filepath.Join(featurePath, stateFileName)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, featurePath, stateFilePath, tt.phaseToSet)
			}

			err := s.SetFeaturePhase(context.Background(), tt.featureName, tt.phaseToSet)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SetFeaturePhase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteFeature(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name        string
		featureName string
		setupMocks  func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string)
		wantErr     bool
	}{
		{
			name:        "successful deletion",
			featureName: "feature-to-delete",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string) {
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte("some-other-active-feature"), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FName: "feature-to-delete", FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:        "feature not found",
			featureName: "non-existent-feature",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string) {
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte("some-other-active-feature"), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "stat fails with unexpected error",
			featureName: "stat-error-feature",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string) {
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte("some-other-active-feature"), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(nil, fmt.Errorf("some stat error")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "RemoveAll fails",
			featureName: "remove-fail-feature",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string) {
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte("some-other-active-feature"), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FName: "remove-fail-feature", FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(featurePath).Return(fmt.Errorf("remove all failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "attempt to delete active feature",
			featureName: "active-feature-to-delete",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featurePath string) {
				// 1. GetActiveFeature is called, returns "active-feature-to-delete"
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte("active-feature-to-delete"), nil).Times(1)
				// 2. ClearActiveFeature is called (which internally calls fs.Remove)
				mockFS.EXPECT().Remove(s.activeFeatureFilePath).Return(nil).Times(1)
				// 3. Stat is called on the feature path
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FName: "active-feature-to-delete", FIsDir: true}, nil).Times(1)
				// 4. RemoveAll is called on the feature path
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
			},
			wantErr: false, // Expect no error from DeleteFeature itself
			// We also need to check activeContextCleared, but that's done in the test body by checking the first return value.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)
			featurePath := filepath.Join(s.featuresDir, tt.featureName)

			if tt.setupMocks != nil {
				tt.setupMocks(s, mockFS, featurePath)
			}

			gotActiveContextCleared, err := s.DeleteFeature(ctx, tt.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteFeature() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Specific check for the "attempt to delete active feature" case
			if tt.name == "attempt to delete active feature" && !tt.wantErr {
				if !gotActiveContextCleared {
					t.Errorf("Service.DeleteFeature() gotActiveContextCleared = %v, want true for active feature deletion", gotActiveContextCleared)
				}
			} else if tt.name != "attempt to delete active feature" && gotActiveContextCleared {
				// For other cases, activeContextCleared should be false unless specifically designed otherwise
				t.Errorf("Service.DeleteFeature() gotActiveContextCleared = %v, want false for non-active feature deletion or error cases", gotActiveContextCleared)
			}
		})
	}
}

// --- Tests for Active Feature State Management ---

func TestService_GetActiveFeature(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, mockFS := newTestService(t, ctrl)
	activeFeatureFilePath := s.activeFeatureFilePath

	tests := []struct {
		name            string
		setupMocks      func(mockFS *portsmocks.MockFileSystem, filePath string)
		wantFeatureName string
		wantErr         bool
	}{
		{
			name: "active feature file does not exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, filePath string) {
				mockFS.EXPECT().ReadFile(filePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantFeatureName: "",
			wantErr:         false,
		},
		{
			name: "error reading active feature file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, filePath string) {
				mockFS.EXPECT().ReadFile(filePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			wantFeatureName: "",
			wantErr:         true,
		},
		{
			name: "successful load with content",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, filePath string) {
				fileContent := []byte("  my-active-feature  ")
				mockFS.EXPECT().ReadFile(filePath).Return(fileContent, nil).Times(1)
			},
			wantFeatureName: "my-active-feature",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, activeFeatureFilePath)
			}

			gotFeatureName, err := s.GetActiveFeature()

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFeatureName != tt.wantFeatureName {
				t.Errorf("Service.GetActiveFeature() gotFeatureName = %q, want %q", gotFeatureName, tt.wantFeatureName)
			}
		})
	}
}

func TestService_SetActiveFeature(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, mockFS := newTestService(t, ctrl)
	activeFeatureFilePath := s.activeFeatureFilePath
	dirOfActiveFeatureFile := filepath.Dir(activeFeatureFilePath)

	tests := []struct {
		name              string
		featureNameToSave string
		setupMocks        func(featureName string)
		wantErr           bool
	}{
		{
			name:              "error creating directory",
			featureNameToSave: "feat1",
			setupMocks: func(featureName string) {
				mockFS.EXPECT().MkdirAll(dirOfActiveFeatureFile, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:              "error writing file",
			featureNameToSave: "feat2",
			setupMocks: func(featureName string) {
				data := []byte(featureName)
				mockFS.EXPECT().MkdirAll(dirOfActiveFeatureFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(activeFeatureFilePath, data, os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:              "successful save",
			featureNameToSave: "feat-ok",
			setupMocks: func(featureName string) {
				data := []byte(featureName)
				mockFS.EXPECT().MkdirAll(dirOfActiveFeatureFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(activeFeatureFilePath, data, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(tt.featureNameToSave)
			err := s.SetActiveFeature(tt.featureNameToSave)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SetActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ClearActiveFeature(t *testing.T) {
	ctrl := gomock.NewController(t)
	s, mockFS := newTestService(t, ctrl)
	activeFeatureFilePath := s.activeFeatureFilePath

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "successful clear (file exists)",
			setupMocks: func() {
				mockFS.EXPECT().Remove(activeFeatureFilePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "successful clear (file does not exist)",
			setupMocks: func() {
				mockFS.EXPECT().Remove(activeFeatureFilePath).Return(os.ErrNotExist).Times(1)
			},
			wantErr: false,
		},
		{
			name: "error during removal",
			setupMocks: func() {
				mockFS.EXPECT().Remove(activeFeatureFilePath).Return(fmt.Errorf("permission denied")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := s.ClearActiveFeature()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ClearActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
