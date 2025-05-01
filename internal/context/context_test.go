package context

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadContext(t *testing.T) {
	i3Dir := t.TempDir()
	contextFile := contextFilePath(i3Dir)

	tests := []struct {
		name        string
		setup       func(t *testing.T)
		wantContext Context
		wantErr     bool
	}{
		{
			name: "file not found",
			setup: func(t *testing.T) {
				// No setup needed, file doesn't exist
			},
			wantContext: Context{}, // Expect empty context
			wantErr:     false,
		},
		{
			name: "empty file",
			setup: func(t *testing.T) {
				require.NoError(t, os.WriteFile(contextFile, []byte(""), 0644))
			},
			wantContext: Context{}, // Expect empty context
			wantErr:     false,
		},
		{
			name: "valid context file",
			setup: func(t *testing.T) {
				content := `{"feature": "feat-a", "phase": "ideation"}`
				require.NoError(t, os.WriteFile(contextFile, []byte(content), 0644))
			},
			wantContext: Context{Feature: "feat-a", Phase: "ideation"},
			wantErr:     false,
		},
		{
			name: "partially valid context file",
			setup: func(t *testing.T) {
				content := `{"feature": "feat-b"}` // Phase omitted
				require.NoError(t, os.WriteFile(contextFile, []byte(content), 0644))
			},
			wantContext: Context{Feature: "feat-b", Phase: ""},
			wantErr:     false,
		},
		{
			name: "invalid json file",
			setup: func(t *testing.T) {
				content := `{"feature": "feat-c",` // Malformed JSON
				require.NoError(t, os.WriteFile(contextFile, []byte(content), 0644))
			},
			wantContext: Context{}, // Expect zero value
			wantErr:     true,
		},
		{
			name: "directory instead of file",
			setup: func(t *testing.T) {
				require.NoError(t, os.Mkdir(contextFile, 0755)) // Create dir at the path
			},
			wantContext: Context{}, // Expect zero value
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up before setup (in case previous test left artifacts)
			_ = os.RemoveAll(contextFile)

			tc.setup(t)

			ctx, err := LoadContext(i3Dir)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantContext, ctx)
			}
		})
	}
}

func TestSaveContext(t *testing.T) {
	i3Dir := t.TempDir()
	contextFile := contextFilePath(i3Dir)

	tests := []struct {
		name    string
		saveCtx Context
		wantStr string // Expected file content
		wantErr bool
	}{
		{
			name:    "save full context",
			saveCtx: Context{Feature: "f1", Phase: "p1"},
			wantStr: `{` + "\n" + `  "feature": "f1",` + "\n" + `  "phase": "p1"` + "\n" + `}` + "\n",
			wantErr: false,
		},
		{
			name:    "save partial context (feature only)",
			saveCtx: Context{Feature: "f2"},
			wantStr: `{` + "\n" + `  "feature": "f2"` + "\n" + `}` + "\n", // Phase omitted due to omitempty
			wantErr: false,
		},
		{
			name:    "save empty context",
			saveCtx: Context{},
			wantStr: "{}\n", // Writes empty object
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clean slate
			_ = os.RemoveAll(contextFile)

			err := SaveContext(i3Dir, tc.saveCtx)

			if tc.wantErr {
				assert.Error(t, err)
				_, statErr := os.Stat(contextFile)
				assert.True(t, os.IsNotExist(statErr) || statErr == nil) // File might or might not exist on error
			} else {
				assert.NoError(t, err)
				content, readErr := os.ReadFile(contextFile)
				require.NoError(t, readErr)
				assert.Equal(t, tc.wantStr, string(content))
			}
		})
	}

	// Test case: cannot create directory
	t.Run("cannot create directory", func(t *testing.T) {
		// Create a file where the i3Dir should be
		_ = os.RemoveAll(i3Dir)
		f, err := os.Create(i3Dir)
		require.NoError(t, err)
		f.Close()

		err = SaveContext(i3Dir, Context{Feature: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ensuring context directory")
	})

	// Test case: cannot write file (e.g., permissions)
	t.Run("cannot write file", func(t *testing.T) {
		// Make dir read-only
		_ = os.RemoveAll(i3Dir)
		require.NoError(t, os.MkdirAll(i3Dir, 0755))
		require.NoError(t, os.Chmod(i3Dir, 0555)) // Read-only
		defer func() { _ = os.Chmod(i3Dir, 0755) }() // Cleanup

		err := SaveContext(i3Dir, Context{Feature: "test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing context file")
	})
} 