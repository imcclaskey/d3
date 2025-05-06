package testutil

import (
	"os"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// MockFileInfo is a minimal implementation of os.FileInfo for shared test usage.
type MockFileInfo struct {
	FName    string
	FIsDir   bool
	FSize    int64
	FMode    os.FileMode
	FModTime time.Time
}

func (mfi MockFileInfo) Name() string       { return mfi.FName }
func (mfi MockFileInfo) Size() int64        { return mfi.FSize }
func (mfi MockFileInfo) Mode() os.FileMode  { return mfi.FMode }
func (mfi MockFileInfo) ModTime() time.Time { return mfi.FModTime }
func (mfi MockFileInfo) IsDir() bool        { return mfi.FIsDir }
func (mfi MockFileInfo) Sys() interface{}   { return nil }

// NewClosableMockFile creates a dummy *os.File that can be successfully closed.
// It uses os.Pipe and returns the writer end, closing the reader end immediately.
// This is useful for mocking os.Create when the test only needs to verify Close() succeeds.
func NewClosableMockFile(t *testing.T) *os.File {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("testutil.NewClosableMockFile: os.Pipe() failed: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("testutil.NewClosableMockFile: closing read pipe failed (non-critical): %v", err)
	}
	return w
}

// MCPCallToolRequestParams represents the anonymous struct for CallToolRequest.Params.
// We define it here as a named type for convenience in test helpers.
// Note: This exactly mirrors the anonymous struct defined in mcp.CallToolRequest.
type MCPCallToolRequestParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Meta      *struct {
		ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
	} `json:"_meta,omitempty"`
}

// NewTestCallToolRequest creates an mcp.CallToolRequest for testing.
// Tool name can be provided if the handler under test uses it.
func NewTestCallToolRequest(toolName string, args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: MCPCallToolRequestParams{
			Name:      toolName,
			Arguments: args,
		},
	}
}
