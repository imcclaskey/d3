package projectfiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	portsmocks "github.com/imcclaskey/d3/internal/core/ports/mocks"
)

func TestEnsureMCPJSON(t *testing.T) {
	tests := []struct {
		name              string
		projectRoot       string
		initialMCPContent []byte // nil if file does not exist, or actual content
		readFileErr       error  // Error for ReadFile mock
		writeFileErr      error  // Error for WriteFile mock
		expectErr         bool
		expectedMCPConfig *MCPRootConfig // Expected structure written to file or returned
	}{
		// Scenario 1: New file creation (file doesn't exist)
		{
			name:        "create new mcp.json (file doesn't exist)",
			projectRoot: "/testroot",
			expectedMCPConfig: &MCPRootConfig{
				MCPServers: MCPServersMap{
					D3ServerName: MCPServerDetail{
						Command: D3Command,
						Args:    MCPServerArgs{fmt.Sprintf("%s%s", D3ServeArgPrefix, "/testroot")},
					},
				},
			},
		},
		{
			name:              "create new mcp.json (explicitly not exist)",
			projectRoot:       "/testroot2",
			initialMCPContent: nil,
			readFileErr:       os.ErrNotExist,
			expectedMCPConfig: &MCPRootConfig{
				MCPServers: MCPServersMap{
					D3ServerName: MCPServerDetail{
						Command: D3Command,
						Args:    MCPServerArgs{fmt.Sprintf("%s%s", D3ServeArgPrefix, "/testroot2")},
					},
				},
			},
		},
		// Scenario 2: tryPreserve=true, existing valid mcp.json
		{
			name:        "update existing mcp.json, preserve other entries",
			projectRoot: "/testroot3",
			initialMCPContent: []byte(`{
				"mcpServers": {
					"otherServer": {
						"command": "other",
						"args": ["arg1"]
					},
					"d3": {
						"command": "d3",
						"args": ["serve --workdir /old/path"]
					}
				}
			}`),
			expectedMCPConfig: &MCPRootConfig{
				MCPServers: MCPServersMap{
					"otherServer": {Command: "other", Args: MCPServerArgs{"arg1"}},
					D3ServerName:  {Command: D3Command, Args: MCPServerArgs{fmt.Sprintf("%s%s", D3ServeArgPrefix, "/testroot3")}},
				},
			},
		},
		// Scenario 3: tryPreserve=true, existing corrupted mcp.json
		{
			name:              "overwrite corrupted mcp.json",
			projectRoot:       "/testroot4",
			initialMCPContent: []byte("{\"invalid_json\": ..."),
			expectedMCPConfig: &MCPRootConfig{
				MCPServers: MCPServersMap{
					D3ServerName: MCPServerDetail{
						Command: D3Command,
						Args:    MCPServerArgs{fmt.Sprintf("%s%s", D3ServeArgPrefix, "/testroot4")},
					},
				},
			},
		},
		// Scenario 4: WriteFile error
		{
			name:         "WriteFile error",
			projectRoot:  "/testroot5",
			writeFileErr: errors.New("disk full"),
			expectErr:    true,
		},
		// Scenario 5: ReadFile error (not os.ErrNotExist)
		{
			name:        "ReadFile error (not ErrNotExist)",
			projectRoot: "/testroot6",
			readFileErr: errors.New("permission denied"),
			expectErr:   true,
		},
		// Scenario 6: .cursor directory does not exist - EnsureMCPJSON should attempt write and fail
		{
			name:        "error if .cursor directory does not exist",
			projectRoot: "/testroot_no_cursor_dir",
			readFileErr: os.ErrNotExist, // ReadFile will indicate mcp.json doesn't exist
			// EnsureMCPJSON will proceed to try and write. This write should fail because .cursor doesn't exist.
			writeFileErr: errors.New("simulated WriteFile error: dir does not exist"),
			expectErr:    true, // The overall function should return this error
			// No expectedMCPConfig as the write fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			mcpPath := filepath.Join(tt.projectRoot, ".cursor", "mcp.json")
			// mcpDir := filepath.Dir(mcpPath) // Not strictly needed here if MkdirAll is removed from main code for this path
			var capturedConfigForTest *MCPRootConfig

			readFileContent := tt.initialMCPContent
			readFileError := tt.readFileErr
			if readFileContent == nil && readFileError == nil {
				readFileError = os.ErrNotExist
			}
			mockFS.EXPECT().ReadFile(mcpPath).
				Return(readFileContent, readFileError).MaxTimes(1)

			// WriteFile should only be expected if we don't expect an error from EnsureMCPJSON,
			// or if the expected error is specifically from the WriteFile operation itself.
			if !tt.expectErr || tt.writeFileErr != nil {
				// If MkdirAll was part of the non-error path, it would be expected here.
				// e.g., mockFS.EXPECT().MkdirAll(mcpDir, os.FileMode(0755)).Return(nil).MaxTimes(1)
				// However, with the latest change, EnsureMCPJSON does not call MkdirAll for mcpPath.

				mockFS.EXPECT().WriteFile(mcpPath, gomock.Any(), os.FileMode(0644)).
					DoAndReturn(func(_ string, data []byte, _ os.FileMode) error {
						if tt.writeFileErr == nil { // If WriteFile itself is not mocked to error
							var parsedConfig MCPRootConfig
							if err := json.Unmarshal(data, &parsedConfig); err != nil {
								t.Fatalf("Failed to unmarshal data written to mcp.json in mock: %v", err)
							}
							capturedConfigForTest = &parsedConfig // Capture the config

							if tt.expectedMCPConfig != nil { // Perform checks if there's an expected config
								if len(parsedConfig.MCPServers) != len(tt.expectedMCPConfig.MCPServers) {
									t.Errorf("Expected %d server entries in written data, got %d", len(tt.expectedMCPConfig.MCPServers), len(parsedConfig.MCPServers))
								}
								d3Expected, okExpected := tt.expectedMCPConfig.MCPServers[D3ServerName]
								d3ActualInWrite, okActualInWrite := parsedConfig.MCPServers[D3ServerName]
								if okExpected != okActualInWrite || (okExpected && (d3ActualInWrite.Command != d3Expected.Command || d3ActualInWrite.Args[0] != d3Expected.Args[0])) {
									t.Errorf("Mismatched d3 entry in written data. Expected: %+v, Got: %+v", d3Expected, d3ActualInWrite)
								}
							}
						}
						return tt.writeFileErr
					}).MaxTimes(1)
			}

			op := NewDefaultFileOperator()
			err := op.EnsureMCPJSON(mockFS, tt.projectRoot)

			if (err != nil) != tt.expectErr {
				t.Errorf("EnsureMCPJSON() error = %v, wantErr %v", err, tt.expectErr)
				return
			}

			if !tt.expectErr && tt.expectedMCPConfig != nil {
				// Use capturedConfigForTest for these assertions
				if capturedConfigForTest == nil {
					t.Errorf("EnsureMCPJSON() was expected to succeed and write a file, but captured config is nil.")
					return
				}
				if len(capturedConfigForTest.MCPServers) != len(tt.expectedMCPConfig.MCPServers) {
					t.Errorf("Captured config (from written file) has %d server entries, want %d", len(capturedConfigForTest.MCPServers), len(tt.expectedMCPConfig.MCPServers))
				}
				d3Expected, okExpected := tt.expectedMCPConfig.MCPServers[D3ServerName]
				d3ActualReturned, okActualReturned := capturedConfigForTest.MCPServers[D3ServerName]
				if okExpected != okActualReturned || (okExpected && (d3ActualReturned.Command != d3Expected.Command || d3ActualReturned.Args[0] != d3Expected.Args[0])) {
					t.Errorf("Captured d3 entry (from written file) mismatch. Expected: %+v, Got: %+v", d3Expected, d3ActualReturned)
				}
			}
		})
	}
}

// TestEnsureRootGitignoreEntries tests the EnsureRootGitignoreEntries function
func TestEnsureRootGitignoreEntries(t *testing.T) {
	tests := []struct {
		name             string
		projectRoot      string
		initialContent   []byte // nil if file does not exist, otherwise the initial content
		readFileErr      error  // Error for ReadFile mock
		writeFileErr     error  // Error for WriteFile mock
		expectErr        bool
		expectedPatterns []string // Patterns that should be present in the final file
	}{
		{
			name:        "create new gitignore (file doesn't exist)",
			projectRoot: "/testroot",
			readFileErr: os.ErrNotExist,
			expectedPatterns: []string{
				"# d3",
				".cursor/rules/d3/",
				".cursor/rules/d3/*.gen.mdc",
				".d3/.feature",
				".d3/features/*/.phase",
			},
		},
		{
			name:        "update existing gitignore without D3 section",
			projectRoot: "/testroot2",
			initialContent: []byte(`# Go Binaries
bin/
*.exe
*.dll
*.so
*.dylib

# Test coverage
*.out
*.test

# IDE / Editor directories
.idea/
.vscode/
`),
			expectedPatterns: []string{
				"# Go Binaries",
				"bin/",
				"# d3",
				".cursor/rules/d3/",
				".cursor/rules/d3/*.gen.mdc",
				".d3/.feature",
				".d3/features/*/.phase",
			},
		},
		{
			name:        "update existing gitignore with D3 section",
			projectRoot: "/testroot3",
			initialContent: []byte(`# Go Binaries
bin/
*.exe
*.dll

# d3
.cursor/rules/d3/
.d3/.feature
# This is an outdated pattern
.d3/some_old_pattern

# Other entries
*.log
`),
			expectedPatterns: []string{
				"# Go Binaries",
				"bin/",
				"# d3",
				".cursor/rules/d3/",
				".cursor/rules/d3/*.gen.mdc", // New pattern should be added
				".d3/.feature",
				".d3/features/*/.phase", // New pattern should be added
				"# Other entries",
				"*.log",
			},
			// These patterns should NOT be in the output
			// .d3/some_old_pattern (should be replaced)
		},
		{
			name:         "error on WriteFile",
			projectRoot:  "/testroot4",
			writeFileErr: errors.New("write error"),
			expectErr:    true,
		},
		{
			name:        "error on ReadFile (not os.ErrNotExist)",
			projectRoot: "/testroot5",
			readFileErr: errors.New("read error"),
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockFS := portsmocks.NewMockFileSystem(ctrl)

			gitignorePath := filepath.Join(tt.projectRoot, ".gitignore")

			// Set up ReadFile expectation
			readFileContent := tt.initialContent
			readFileError := tt.readFileErr
			mockFS.EXPECT().ReadFile(gitignorePath).Return(readFileContent, readFileError).Times(1)

			// Set up WriteFile expectation
			if !tt.expectErr || tt.writeFileErr != nil {
				mockFS.EXPECT().WriteFile(gitignorePath, gomock.Any(), os.FileMode(0644)).
					DoAndReturn(func(_ string, data []byte, _ os.FileMode) error {
						if tt.writeFileErr == nil { // If WriteFile itself is not mocked to error
							content := string(data)

							// Check that all expected patterns are in the file
							for _, pattern := range tt.expectedPatterns {
								if !strings.Contains(content, pattern) {
									t.Errorf("Expected pattern %q not found in gitignore content:\n%s", pattern, content)
								}
							}

							// If this is an update case, make sure we preserved the existing content outside the D3 section
							if tt.initialContent != nil && len(tt.initialContent) > 0 {
								// Check that non-D3 sections from the initial content are preserved
								// Here we check for a few key lines that should be preserved
								for _, line := range strings.Split(string(tt.initialContent), "\n") {
									trimmed := strings.TrimSpace(line)
									// Skip empty lines, D3 section marker, or lines in the D3 section
									if trimmed == "" || trimmed == "# d3" || strings.HasPrefix(trimmed, ".d3/") || strings.HasPrefix(trimmed, ".cursor/rules/d3") {
										continue
									}
									// Any other line should be preserved
									if !strings.Contains(content, line) {
										t.Errorf("Expected line %q to be preserved in gitignore content", line)
									}
								}
							}
						}
						return tt.writeFileErr
					}).Times(1)
			}

			op := NewDefaultFileOperator()
			err := op.EnsureRootGitignoreEntries(mockFS, tt.projectRoot)

			if (err != nil) != tt.expectErr {
				t.Errorf("EnsureRootGitignoreEntries() error = %v, wantErr %v", err, tt.expectErr)
			}
		})
	}
}
