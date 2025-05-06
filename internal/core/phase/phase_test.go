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
				// Let Define succeed
				defineDir := filepath.Join(featureRoot, string(Define))
				defineFile := filepath.Join(defineDir, PhaseFileMap[Define])
				mockFS.EXPECT().MkdirAll(defineDir, gomock.Any()).Return(nil).Times(1)
				mockFS.EXPECT().Stat(defineFile).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().Create(defineFile).Return(testutil.NewClosableMockFile(t), nil).Times(1)

				// Fail Design
				designDir := filepath.Join(featureRoot, string(Design))
				mockFS.EXPECT().MkdirAll(designDir, gomock.Any()).Return(fmt.Errorf("mkdir failed")).Times(1)
				// No further calls expected for Design or Deliver after error
			},
			wantErr: true,
		},
		{
			name: "error on Stat for one file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Let Define succeed
				defineDir := filepath.Join(featureRoot, string(Define))
				defineFile := filepath.Join(defineDir, PhaseFileMap[Define])
				mockFS.EXPECT().MkdirAll(defineDir, gomock.Any()).Return(nil).Times(1)
				mockFS.EXPECT().Stat(defineFile).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().Create(defineFile).Return(testutil.NewClosableMockFile(t), nil).Times(1)

				// Let Design MkdirAll succeed
				designDir := filepath.Join(featureRoot, string(Design))
				designFile := filepath.Join(designDir, PhaseFileMap[Design])
				mockFS.EXPECT().MkdirAll(designDir, gomock.Any()).Return(nil).Times(1)
				// Fail Stat
				mockFS.EXPECT().Stat(designFile).Return(nil, fmt.Errorf("stat failed")).Times(1)
				// No further calls expected
			},
			wantErr: true,
		},
		{
			name: "error on Create for one file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// Let Define succeed
				defineDir := filepath.Join(featureRoot, string(Define))
				defineFile := filepath.Join(defineDir, PhaseFileMap[Define])
				mockFS.EXPECT().MkdirAll(defineDir, gomock.Any()).Return(nil).Times(1)
				mockFS.EXPECT().Stat(defineFile).Return(nil, os.ErrNotExist).Times(1)
				mockFS.EXPECT().Create(defineFile).Return(testutil.NewClosableMockFile(t), nil).Times(1)

				// Let Design Stat succeed (needs create)
				designDir := filepath.Join(featureRoot, string(Design))
				designFile := filepath.Join(designDir, PhaseFileMap[Design])
				mockFS.EXPECT().MkdirAll(designDir, gomock.Any()).Return(nil).Times(1)
				mockFS.EXPECT().Stat(designFile).Return(nil, os.ErrNotExist).Times(1)
				// Fail Create
				mockFS.EXPECT().Create(designFile).Return(nil, fmt.Errorf("create failed")).Times(1)
				// No further calls expected
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
