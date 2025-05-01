package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Context holds the current workspace context (feature and phase).
type Context struct {
	Feature string `json:"feature,omitempty"`
	Phase   string `json:"phase,omitempty"`
}

// contextFilePath returns the expected path for the context file.
func contextFilePath(i3Dir string) string {
	return filepath.Join(i3Dir, "context.json")
}

// LoadContext loads the workspace context from the context file.
// If the file doesn't exist, it returns an empty context (not an error).
func LoadContext(i3Dir string) (Context, error) {
	path := contextFilePath(i3Dir)
	data, err := os.ReadFile(path)
	if err != nil {
		// File not existing is okay, just means no context is set.
		if errors.Is(err, os.ErrNotExist) {
			return Context{}, nil
		}
		// Other read errors are problems.
		return Context{}, fmt.Errorf("reading context file %s: %w", path, err)
	}

	// Handle empty file case (equivalent to no context)
	if len(data) == 0 {
		return Context{}, nil
	}

	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		// Corrupted file is an error.
		return Context{}, fmt.Errorf("unmarshaling context file %s: %w", path, err)
	}

	return ctx, nil
}

// SaveContext saves the workspace context to the context file.
func SaveContext(i3Dir string, ctx Context) error {
	path := contextFilePath(i3Dir)

	// Ensure the directory exists before writing
	if err := os.MkdirAll(i3Dir, 0755); err != nil {
		return fmt.Errorf("ensuring context directory %s exists: %w", i3Dir, err)
	}

	// If context is empty, write an empty file or delete it? Writing empty is simpler.
	isEmpty := ctx.Feature == "" && ctx.Phase == ""
	var data []byte
	var err error

	if isEmpty {
		data = []byte("{}\n") // Write empty JSON object for clarity
	} else {
		data, err = json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling context: %w", err)
		}
		data = append(data, '\n') // Add trailing newline
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing context file %s: %w", path, err)
	}

	return nil
} 