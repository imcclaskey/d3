package session

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"

	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	// Import shared test utilities
)

func TestPhase_Valid(t *testing.T) {
	tests := []struct {
		phase Phase
		want  bool
	}{
		{None, true},
		{Define, true},
		{Design, true},
		{Deliver, true},
		{"invalid", false},
		{"Define", false}, // Case-sensitive
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := tt.phase.Valid(); got != tt.want {
				t.Errorf("Phase(%q).Valid() = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestPhase_Next(t *testing.T) {
	tests := []struct {
		phase Phase
		want  Phase
	}{
		{None, Define},
		{Define, Design},
		{Design, Deliver},
		{Deliver, Deliver},     // Stays at Deliver
		{"invalid", "invalid"}, // Invalid stays invalid
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := tt.phase.Next(); got != tt.want {
				t.Errorf("Phase(%q).Next() = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestPhase_String(t *testing.T) {
	tests := []struct {
		phase Phase
		want  string
	}{
		{None, ""},
		{Define, "define"},
		{Design, "design"},
		{Deliver, "deliver"},
		{"something", "something"},
	}
	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			if got := tt.phase.String(); got != tt.want {
				t.Errorf("Phase(%q).String() = %v, want %v", tt.phase, got, tt.want)
			}
		})
	}
}

func TestParsePhase(t *testing.T) {
	tests := []struct {
		input   string
		want    Phase
		wantErr bool
	}{
		{"", None, false},
		{"define", Define, false},
		{"design", Design, false},
		{"deliver", Deliver, false},
		{"Define", None, true}, // Case-sensitive
		{"invalid", None, true},
		{" deliver ", None, true}, // Leading/trailing space
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePhase(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePhase(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePhase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- Tests for Storage (Refactored) ---

func TestStorage_LoadActiveFeature(t *testing.T) {
	d3Dir := t.TempDir()
	sessionFilePath := filepath.Join(d3Dir, ".session") // Path updated

	tests := []struct {
		name            string
		setupMocks      func(mockFS *portsmocks.MockFileSystem)
		wantFeatureName string
		wantErr         bool
	}{
		{
			name: "session file does not exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantFeatureName: "", // Expect empty string if not found
			wantErr:         false,
		},
		{
			name: "error reading session file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			wantFeatureName: "",
			wantErr:         true,
		},
		{
			name: "successful load with content",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				fileContent := []byte("  my-active-feature  ") // Content with whitespace
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(fileContent, nil).Times(1)
			},
			wantFeatureName: "my-active-feature", // Expect trimmed content
			wantErr:         false,
		},
		{
			name: "successful load with empty content",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				fileContent := []byte(" \n ") // Whitespace only
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(fileContent, nil).Times(1)
			},
			wantFeatureName: "", // Expect empty string
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS)
			}

			storage := NewStorage(d3Dir, mockFS)
			gotFeatureName, err := storage.LoadActiveFeature()

			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.LoadActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFeatureName != tt.wantFeatureName {
				t.Errorf("Storage.LoadActiveFeature() gotFeatureName = %q, want %q", gotFeatureName, tt.wantFeatureName)
			}
		})
	}
}

func TestStorage_SaveActiveFeature(t *testing.T) {
	d3Dir := t.TempDir()
	sessionFilePath := filepath.Join(d3Dir, ".session") // Path updated
	dirOfSessionFile := filepath.Dir(sessionFilePath)

	tests := []struct {
		name              string
		featureNameToSave string
		setupMocks        func(mockFS *portsmocks.MockFileSystem, featureName string)
		wantErr           bool
	}{
		{
			name:              "error creating directory",
			featureNameToSave: "feat1",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featureName string) {
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:              "error writing file",
			featureNameToSave: "feat2",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featureName string) {
				data := []byte(featureName)
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(sessionFilePath, data, os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:              "successful save",
			featureNameToSave: "feat-ok",
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featureName string) {
				data := []byte(featureName)
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(sessionFilePath, data, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name:              "successful save of empty string",
			featureNameToSave: "", // e.g., when exiting a feature
			setupMocks: func(mockFS *portsmocks.MockFileSystem, featureName string) {
				data := []byte(featureName)
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(sessionFilePath, data, os.FileMode(0644)).Return(nil).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, tt.featureNameToSave)
			}

			storage := NewStorage(d3Dir, mockFS)
			err := storage.SaveActiveFeature(tt.featureNameToSave)

			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.SaveActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStorage_ClearActiveFeature(t *testing.T) {
	d3Dir := t.TempDir()
	sessionFilePath := filepath.Join(d3Dir, ".session") // Path updated

	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		wantErr    bool
	}{
		{
			name: "successful clear (file exists)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().Remove(sessionFilePath).Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "successful clear (file does not exist)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().Remove(sessionFilePath).Return(os.ErrNotExist).Times(1)
			},
			wantErr: false, // Not an error if already gone
		},
		{
			name: "error during removal",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().Remove(sessionFilePath).Return(fmt.Errorf("permission denied")).Times(1)
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

			storage := NewStorage(d3Dir, mockFS)
			err := storage.ClearActiveFeature()

			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.ClearActiveFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
