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
	// "gopkg.in/yaml.v3" // No longer needed
)

// Helper to create a new service with mock FS for testing
func newTestService(t *testing.T, ctrl *gomock.Controller) (*Service, *portsmocks.MockFileSystem) {
	t.Helper()
	projectRoot := t.TempDir()
	// Construct paths as the NewService function would
	featuresDir := filepath.Join(projectRoot, "features")
	d3Dir := filepath.Join(projectRoot, ".d3")
	mockFS := portsmocks.NewMockFileSystem(ctrl)
	// NewService now uses activeFeatureFileName (".feature") internally
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
				phaseFilePath := filepath.Join(featurePath, phaseFileName) // Use new phaseFileName
				expectedPhaseContent := []byte(string(phase.Define))
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, expectedPhaseContent, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantInfo: &FeatureInfo{Name: "new-feature", Path: ""}, // Path will be dynamic based on temp dir
			wantErr:  false,
		},
		{
			name: "WriteFile for .phase fails", // Renamed test case
			args: args{ctx: context.Background(), featureName: "phase-fail-feature"},
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName) // Use new phaseFileName
				expectedPhaseContent := []byte(string(phase.Define))
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, expectedPhaseContent, os.FileMode(0644)).Return(fmt.Errorf("write phase failed")).Times(1)
				// Expect RemoveAll to be called for cleanup
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
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
				// Simple comparison (doesn\'t require order match)
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

func TestService_GetFeaturePhase(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		setupMocks  func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string)
		wantPhase   phase.Phase
		wantErr     bool
	}{
		{
			name:        "feature does not exist",
			featureName: "nonexistent-feat",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "phase file exists, valid phase",
			featureName: "feat1",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				validPhaseContent := []byte(string(phase.Design))
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For FeatureExists
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(validPhaseContent, nil).Times(1)
			},
			wantPhase: phase.Design,
			wantErr:   false,
		},
		{
			name:        "phase file exists, empty content",
			featureName: "feat-empty-phase",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				emptyContent := []byte("  ") // Whitespace only
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(emptyContent, nil).Times(1)
			},
			wantPhase: phase.None, // Expect error for empty phase file
			wantErr:   true,
		},
		{
			name:        "phase file exists, invalid phase string",
			featureName: "feat-invalid-phase-str",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				invalidContent := []byte("not_a_phase")
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(invalidContent, nil).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "phase file does not exist, creates default (define)",
			featureName: "feat-new",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				expectedWriteContent := []byte(string(phase.Define))
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For FeatureExists
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, expectedWriteContent, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantPhase: phase.Define,
			wantErr:   false,
		},
		{
			name:        "phase file does not exist, MkdirAll fails",
			featureName: "feat-mkdir-fail",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "phase file does not exist, WriteFile fails",
			featureName: "feat-write-fail",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, []byte(string(phase.Define)), os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
		{
			name:        "ReadFile returns other error",
			featureName: "feat-read-err",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().ReadFile(phaseFilePath).Return(nil, fmt.Errorf("random read error")).Times(1)
			},
			wantPhase: phase.None,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(s, mockFS, tt.featureName)
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
		setupMocks  func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, p phase.Phase)
		wantErr     bool
	}{
		{
			name:        "successful set phase",
			featureName: "feat-set1",
			phaseToSet:  phase.Deliver,
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, p phase.Phase) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				expectedWriteContent := []byte(string(p))
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1) // For FeatureExists
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, expectedWriteContent, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:        "feature does not exist",
			featureName: "nonexistent-feat-set",
			phaseToSet:  phase.Design,
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, p phase.Phase) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1) // For FeatureExists
			},
			wantErr: true,
		},
		{
			name:        "invalid phase to set (None)",
			featureName: "feat-set-invalid",
			phaseToSet:  phase.None,
			setupMocks:  nil, // No FS interaction if phase is invalid early
			wantErr:     true,
		},
		{
			name:        "invalid phase to set (custom string)",
			featureName: "feat-set-invalid-str",
			phaseToSet:  phase.Phase("bad-phase"),
			setupMocks:  nil, // No FS interaction
			wantErr:     true,
		},
		{
			name:        "MkdirAll fails",
			featureName: "feat-set-mkdir-fail",
			phaseToSet:  phase.Design,
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, p phase.Phase) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "WriteFile fails",
			featureName: "feat-set-write-fail",
			phaseToSet:  phase.Design,
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, p phase.Phase) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				phaseFilePath := filepath.Join(featurePath, phaseFileName)
				expectedWriteContent := []byte(string(p))
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().MkdirAll(featurePath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(phaseFilePath, expectedWriteContent, os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(s, mockFS, tt.featureName, tt.phaseToSet)
			}

			err := s.SetFeaturePhase(context.Background(), tt.featureName, tt.phaseToSet)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SetFeaturePhase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_DeleteFeature(t *testing.T) {
	tests := []struct {
		name                 string
		featureName          string
		activeFeatureContent string // Content of the .feature file (or empty if not exists)
		setupMocks           func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string)
		wantCleared          bool
		wantErr              bool
	}{
		{
			name:                 "delete existing feature, not active",
			featureName:          "feat-to-delete",
			activeFeatureContent: "other-active-feat",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				// GetActiveFeature reads .feature file
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				// Stat checks if feature directory exists
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				// RemoveAll deletes the feature directory
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
			},
			wantCleared: false,
			wantErr:     false,
		},
		{
			name:                 "delete existing feature, which is active",
			featureName:          "active-feat-to-delete",
			activeFeatureContent: "active-feat-to-delete", // This feature is active
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				// ClearActiveFeature will remove .feature file
				mockFS.EXPECT().Remove(s.activeFeatureFilePath).Return(nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
			},
			wantCleared: true,
			wantErr:     false,
		},
		{
			name:                 "delete active feature, ClearActiveFeature fails",
			featureName:          "active-feat-clear-fail",
			activeFeatureContent: "active-feat-clear-fail",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				mockFS.EXPECT().Remove(s.activeFeatureFilePath).Return(fmt.Errorf("failed to remove .feature")).Times(1)
				// No further calls if ClearActiveFeature fails
			},
			wantCleared: false,
			wantErr:     true,
		},
		{
			name:                 "feature to delete does not exist (Stat returns ErrNotExist)",
			featureName:          "nonexistent-delete",
			activeFeatureContent: "some-active-feat",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(nil, os.ErrNotExist).Times(1)
				// No RemoveAll if Stat shows it doesn\'t exist
			},
			wantCleared: false,
			wantErr:     true, // Error because feature not found
		},
		{
			name:                 "Stat for feature to delete fails with other error",
			featureName:          "stat-fail-delete",
			activeFeatureContent: "some-active-feat",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(nil, fmt.Errorf("stat failed badly")).Times(1)
			},
			wantCleared: false,
			wantErr:     true,
		},
		{
			name:                 "RemoveAll fails",
			featureName:          "removeall-fail-delete",
			activeFeatureContent: "other-feat",
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return([]byte(activeFeatureContent), nil).Times(1)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(featurePath).Return(fmt.Errorf("remove all failed")).Times(1)
			},
			wantCleared: false,
			wantErr:     true,
		},
		{
			name:                 "GetActiveFeature fails (ReadFile for .feature fails)",
			featureName:          "getactive-fail-delete",
			activeFeatureContent: "", // Content doesn\'t matter as ReadFile will fail
			setupMocks: func(s *Service, mockFS *portsmocks.MockFileSystem, featureName string, activeFeatureContent string) {
				// ReadFile for .feature fails
				mockFS.EXPECT().ReadFile(s.activeFeatureFilePath).Return(nil, fmt.Errorf("failed to read .feature")).Times(1)
				// Stat for the feature to delete (should still be called as error from GetActive is warning)
				featurePath := filepath.Join(s.featuresDir, featureName)
				mockFS.EXPECT().Stat(featurePath).Return(testutil.MockFileInfo{FIsDir: true}, nil).Times(1)
				mockFS.EXPECT().RemoveAll(featurePath).Return(nil).Times(1)
			},
			wantCleared: false, // Not cleared because currentActiveFeature would be empty due to error
			wantErr:     false, // Delete itself succeeds, GetActiveFeature error is a warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(s, mockFS, tt.featureName, tt.activeFeatureContent)
			}

			gotCleared, err := s.DeleteFeature(context.Background(), tt.featureName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.DeleteFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCleared != tt.wantCleared {
				t.Errorf("Service.DeleteFeature() gotCleared = %v, want %v", gotCleared, tt.wantCleared)
			}
		})
	}
}

func TestService_GetActiveFeature(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string)
		wantName   string
		wantErr    bool
	}{
		{
			name: "active feature file exists and has content",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().ReadFile(activeFeaturePath).Return([]byte("  my-active-feature  "), nil).Times(1)
			},
			wantName: "my-active-feature",
			wantErr:  false,
		},
		{
			name: "active feature file exists and is empty",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().ReadFile(activeFeaturePath).Return([]byte("  "), nil).Times(1) // Whitespace only
			},
			wantName: "",
			wantErr:  false,
		},
		{
			name: "active feature file does not exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().ReadFile(activeFeaturePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantName: "",
			wantErr:  false, // Not an error, means no active feature
		},
		{
			name: "ReadFile returns other error",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().ReadFile(activeFeaturePath).Return(nil, fmt.Errorf("read failed")).Times(1)
			},
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, s.activeFeatureFilePath) // s.activeFeatureFilePath uses .feature now
			}

			gotName, err := s.GetActiveFeature()

			if (err != nil) != tt.wantErr {
				t.Errorf("Service.GetActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotName != tt.wantName {
				t.Errorf("Service.GetActiveFeature() gotName = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestService_SetActiveFeature(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		setupMocks  func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string, featureName string)
		wantErr     bool
	}{
		{
			name:        "successful set",
			featureName: "new-active-feature",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string, featureName string) {
				dirPath := filepath.Dir(activeFeaturePath)
				mockFS.EXPECT().MkdirAll(dirPath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(activeFeaturePath, []byte(featureName), os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:        "MkdirAll fails for active feature file directory",
			featureName: "mkdirall-fail-active",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string, featureName string) {
				dirPath := filepath.Dir(activeFeaturePath)
				mockFS.EXPECT().MkdirAll(dirPath, os.FileMode(0755)).Return(fmt.Errorf("mkdirall failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "WriteFile fails for active feature file",
			featureName: "write-fail-active",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string, featureName string) {
				dirPath := filepath.Dir(activeFeaturePath)
				mockFS.EXPECT().MkdirAll(dirPath, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(activeFeaturePath, []byte(featureName), os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, s.activeFeatureFilePath, tt.featureName) // s.activeFeatureFilePath uses .feature now
			}

			err := s.SetActiveFeature(tt.featureName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.SetActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_ClearActiveFeature(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string)
		wantErr    bool
	}{
		{
			name: "successful clear (file exists)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().Remove(activeFeaturePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "successful clear (file does not exist, Remove returns ErrNotExist)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().Remove(activeFeaturePath).Return(os.ErrNotExist).Times(1)
			},
			wantErr: false, // ErrNotExist is ignored
		},
		{
			name: "Remove returns other error",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, activeFeaturePath string) {
				mockFS.EXPECT().Remove(activeFeaturePath).Return(fmt.Errorf("remove failed")).Times(1)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			s, mockFS := newTestService(t, ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, s.activeFeatureFilePath) // s.activeFeatureFilePath uses .feature now
			}

			err := s.ClearActiveFeature()
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ClearActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
