package projectfiles

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings" // Uncommented and to be used

	// "strings" // Will be needed by EnsureD3GitignoreEntries

	"github.com/imcclaskey/d3/internal/core/ports"
)

// MCPServerArgs represents the arguments for an MCP server command.
// For d3, this will typically contain one string like "serve --workdir <path>".
type MCPServerArgs []string

// MCPServerDetail holds the command and arguments for a specific MCP server.
type MCPServerDetail struct {
	Command string        `json:"command"`
	Args    MCPServerArgs `json:"args"`
}

// MCPServersMap maps server names (e.g., "d3") to their details.
type MCPServersMap map[string]MCPServerDetail

// MCPRootConfig is the root structure of the mcp.json file.
type MCPRootConfig struct {
	MCPServers MCPServersMap `json:"mcpServers"`
}

// Constants for d3 server entry in mcp.json
const (
	D3ServerName     = "d3"
	D3Command        = "d3"
	D3ServeArgPrefix = "serve --workdir "
)

// DefaultFileOperator implements file operations for project initialization.
type DefaultFileOperator struct{}

// NewDefaultFileOperator creates a new DefaultFileOperator.
func NewDefaultFileOperator() *DefaultFileOperator {
	return &DefaultFileOperator{}
}

// EnsureMCPJSON creates or updates mcp.json in the project root.
// It ensures the 'd3' server entry has the correct command and workdir.
// It always attempts to preserve other entries if mcp.json exists and is valid.
func (op *DefaultFileOperator) EnsureMCPJSON(fs ports.FileSystem, projectRoot string) error {
	mcpPath := filepath.Join(projectRoot, "mcp.json")
	rootConfig := MCPRootConfig{
		MCPServers: make(MCPServersMap),
	}

	d3ServerArgs := MCPServerArgs{fmt.Sprintf("%s%s", D3ServeArgPrefix, projectRoot)}
	d3ServerDetail := MCPServerDetail{
		Command: D3Command,
		Args:    d3ServerArgs,
	}

	// Always try to read and preserve existing mcp.json content
	data, errReadFile := fs.ReadFile(mcpPath)
	if errReadFile == nil { // File exists and is readable
		var existingRootConfig MCPRootConfig
		if jsonErr := json.Unmarshal(data, &existingRootConfig); jsonErr == nil {
			rootConfig = existingRootConfig   // Start with existing config
			if rootConfig.MCPServers == nil { // Ensure map is initialized if it was nil
				rootConfig.MCPServers = make(MCPServersMap)
			}
		} else {
			// If existing file is corrupt, warn and proceed to create a new one with the d3 entry.
			fmt.Fprintf(os.Stderr, "warning: mcp.json is corrupted or unparsable, creating new with d3 entry: %v\n", jsonErr)
			rootConfig.MCPServers = make(MCPServersMap) // Reset to ensure a clean state for the d3 entry
		}
	} else if !os.IsNotExist(errReadFile) { // Some other error reading the file (not just "doesn't exist")
		return fmt.Errorf("failed to read mcp.json for update: %w", errReadFile)
	}
	// If os.IsNotExist(errReadFile), we just proceed with the new rootConfig, which is correct.

	rootConfig.MCPServers[D3ServerName] = d3ServerDetail

	jsonData, err := json.MarshalIndent(rootConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mcp.json: %w", err)
	}

	if err := fs.WriteFile(mcpPath, jsonData, 0644); err != nil {
		return err
	}
	return nil
}

// EnsureD3GitignoreEntries creates .gitignore files in specific d3 directories.
// It takes d3Dir and cursorRulesDir as absolute paths.
func (op *DefaultFileOperator) EnsureD3GitignoreEntries(fs ports.FileSystem, d3DirAbs, cursorRulesD3DirAbs, projectRootAbs string) error {
	type gitignoreTarget struct {
		path    string // Relative to project root, for constructing full path
		content string
	}

	targets := []gitignoreTarget{
		{
			// Path for .d3/.gitignore, ensure it's relative to project root for joining
			path:    filepath.Join(strings.TrimPrefix(d3DirAbs, projectRootAbs+string(filepath.Separator)), ".gitignore"),
			content: ".session\n",
		},
		{
			// Path for .cursor/rules/d3/.gitignore, ensure it's relative for joining
			path:    filepath.Join(strings.TrimPrefix(cursorRulesD3DirAbs, projectRootAbs+string(filepath.Separator)), "d3", ".gitignore"),
			content: "*.gen.mdc\n",
		},
	}

	for _, target := range targets {
		// Construct full path from projectRoot and the relative target path
		fullPath := filepath.Join(projectRootAbs, target.path)
		dir := filepath.Dir(fullPath)

		if err := fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s for .gitignore: %w", dir, err)
		}

		if err := fs.WriteFile(fullPath, []byte(target.content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fullPath, err)
		}
	}
	return nil
}
