package session

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"gopkg.in/yaml.v3"

	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
	"github.com/imcclaskey/d3/internal/testutil" // Import shared test utilities
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

// --- Tests for Storage ---

func TestStorage_Load(t *testing.T) {
	d3Dir := t.TempDir() // Use a temporary directory for test isolation
	sessionFilePath := filepath.Join(d3Dir, "session.yaml")

	tests := []struct {
		name       string
		setupMocks func(mockFS *portsmocks.MockFileSystem)
		wantState  *SessionState // Note: This SessionState struct no longer has CurrentPhase
		wantErr    bool
	}{
		{
			name: "session file does not exist",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().Stat(sessionFilePath).Return(nil, os.ErrNotExist).Times(1)
			},
			wantState: nil,
			wantErr:   true, // Expect specific error about file not existing
		},
		{
			name: "error reading session file",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				mockFS.EXPECT().Stat(sessionFilePath).Return(testutil.MockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(nil, fmt.Errorf("read error")).Times(1)
			},
			wantState: nil,
			wantErr:   true,
		},
		{
			name: "error parsing session file (invalid YAML)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				invalidYAML := []byte("current_feature: test\ncurrent_phase: define: oops") // Phase field still exists in data, but ignored by struct
				mockFS.EXPECT().Stat(sessionFilePath).Return(testutil.MockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(invalidYAML, nil).Times(1)
			},
			wantState: nil,
			wantErr:   true,
		},
		{
			name: "successful load (ignoring phase field in yaml)",
			setupMocks: func(mockFS *portsmocks.MockFileSystem) {
				// YAML data might still contain current_phase from old files, but it should be ignored on load
				validYAML := []byte("current_feature: my-feat\ncurrent_phase: design\nversion: 1.1")
				mockFS.EXPECT().Stat(sessionFilePath).Return(testutil.MockFileInfo{}, nil).Times(1)
				mockFS.EXPECT().ReadFile(sessionFilePath).Return(validYAML, nil).Times(1)
			},
			// Expected state only includes fields present in the struct
			wantState: &SessionState{CurrentFeature: "my-feat", Version: "1.1"},
			wantErr:   false,
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
			gotState, err := storage.Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare states (ignoring time.Time and removed CurrentPhase)
			if !tt.wantErr {
				if gotState == nil || tt.wantState == nil {
					if gotState != tt.wantState { // Handles case where one is nil and the other isn't
						t.Errorf("Storage.Load() gotState = %v, want %v", gotState, tt.wantState)
					}
				} else if gotState.CurrentFeature != tt.wantState.CurrentFeature ||
					// gotState.CurrentPhase != tt.wantState.CurrentPhase || // Comparison removed
					gotState.Version != tt.wantState.Version {
					t.Errorf("Storage.Load() gotState = %+v, want %+v", gotState, tt.wantState)
				}
			}
		})
	}
}

func TestStorage_Save(t *testing.T) {
	d3Dir := t.TempDir() // Use a temporary directory for test isolation
	sessionFilePath := filepath.Join(d3Dir, "session.yaml")
	dirOfSessionFile := filepath.Dir(sessionFilePath)

	tests := []struct {
		name        string
		stateToSave *SessionState // Note: This SessionState struct no longer has CurrentPhase
		setupMocks  func(mockFS *portsmocks.MockFileSystem, state *SessionState)
		wantErr     bool
	}{
		{
			name:        "error creating directory",
			stateToSave: &SessionState{CurrentFeature: "feat"}, // Removed CurrentPhase
			setupMocks: func(mockFS *portsmocks.MockFileSystem, state *SessionState) {
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(fmt.Errorf("mkdir failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "error writing file",
			stateToSave: &SessionState{CurrentFeature: "feat"}, // Removed CurrentPhase
			setupMocks: func(mockFS *portsmocks.MockFileSystem, state *SessionState) {
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(sessionFilePath, gomock.Any(), os.FileMode(0644)).Return(fmt.Errorf("write failed")).Times(1)
			},
			wantErr: true,
		},
		{
			name:        "successful save",
			stateToSave: &SessionState{CurrentFeature: "feat-ok", Version: "1.2"}, // Removed CurrentPhase
			setupMocks: func(mockFS *portsmocks.MockFileSystem, state *SessionState) {
				mockFS.EXPECT().MkdirAll(dirOfSessionFile, os.FileMode(0755)).Return(nil).Times(1)
				mockFS.EXPECT().WriteFile(sessionFilePath, gomock.Any(), os.FileMode(0644)).DoAndReturn(func(path string, data []byte, perm os.FileMode) error {
					var writtenState SessionState
					if err := yaml.Unmarshal(data, &writtenState); err != nil {
						return fmt.Errorf("failed to unmarshal written data for verification: %w", err)
					}
					// Verify only fields present in the struct
					if writtenState.CurrentFeature != state.CurrentFeature || writtenState.Version != state.Version {
						return fmt.Errorf("written data mismatch: got %+v, expected fields from %+v", writtenState, state)
					}
					if writtenState.LastModified.IsZero() { // Ensure LastModified was set
						return fmt.Errorf("LastModified was not set in written data")
					}
					return nil // Success
				}).Times(1)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockFS, tt.stateToSave)
			}

			storage := NewStorage(d3Dir, mockFS)
			err := storage.Save(tt.stateToSave)

			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
