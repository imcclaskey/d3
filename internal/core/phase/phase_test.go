package phase

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	"github.com/imcclaskey/d3/internal/testutil"
)

func TestEnsurePhaseFiles(t *testing.T) {
	featureRoot := t.TempDir() // Use temp dir for isolated feature root

	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		wantErr    bool
	}{
		{
			name: "all directories and files need creation",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// In this test, we expect all phases to be processed fully
				// So we can use the exact map that the code itself uses
				for p, filename := range PhaseFileMap {
					phaseDir := filepath.Join(featureRoot, string(p))
					filePath := filepath.Join(phaseDir, filename)
					mockFS.EXPECT().MkdirAll(phaseDir, os.FileMode(0755)).Return(nil).Times(1)
					mockFS.EXPECT().Stat(filePath).Return(nil, os.ErrNotExist).Times(1)
					mockFS.EXPECT().Create(filePath).Return(testutil.NewClosableMockFile(t), nil).Times(1)
				}
			},
			wantErr: false,
		},
		{
			name: "directories and files already exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// In this test, all files exist already
				for p, filename := range PhaseFileMap {
					phaseDir := filepath.Join(featureRoot, string(p))
					filePath := filepath.Join(phaseDir, filename)
					mockFS.EXPECT().MkdirAll(phaseDir, os.FileMode(0755)).Return(nil).Times(1)   // MkdirAll is idempotent
					mockFS.EXPECT().Stat(filePath).Return(testutil.MockFileInfo{}, nil).Times(1) // File exists
					// Create should NOT be called
				}
			},
			wantErr: false,
		},
		{
			name: "error on MkdirAll for one phase",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// We need exactly ONE MkdirAll to fail, and that will cause the function to return immediately
				anyPhaseDir := filepath.Join(featureRoot, string(Define))
				mockFS.EXPECT().MkdirAll(anyPhaseDir, gomock.Any()).Return(fmt.Errorf("mkdir failed")).Times(1)

				// No other calls should happen after this error
			},
			wantErr: true,
		},
		{
			name: "error on Stat for one file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// First MkdirAll must succeed
				anyPhaseDir := filepath.Join(featureRoot, string(Define))
				mockFS.EXPECT().MkdirAll(anyPhaseDir, gomock.Any()).Return(nil).Times(1)

				// Then Stat will fail and cause early return
				filePath := filepath.Join(featureRoot, string(Define), PhaseFileMap[Define])
				mockFS.EXPECT().Stat(filePath).Return(nil, fmt.Errorf("stat failed")).Times(1)

				// No other calls should happen after this error
			},
			wantErr: true,
		},
		{
			name: "error on Create for one file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// First MkdirAll must succeed
				anyPhaseDir := filepath.Join(featureRoot, string(Define))
				mockFS.EXPECT().MkdirAll(anyPhaseDir, gomock.Any()).Return(nil).Times(1)

				// Then Stat needs to indicate file doesn't exist
				filePath := filepath.Join(featureRoot, string(Define), PhaseFileMap[Define])
				mockFS.EXPECT().Stat(filePath).Return(nil, os.ErrNotExist).Times(1)

				// Then Create will fail and cause early return
				mockFS.EXPECT().Create(filePath).Return(nil, fmt.Errorf("create failed")).Times(1)

				// No other calls should happen after this error
			},
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

			err := EnsurePhaseFiles(mockFS, featureRoot)

			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePhaseFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
