package projectfiles

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// Constants for gitignore management
	D3GitignoreSectionMarker = "# d3"
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
	mcpPath := filepath.Join(projectRoot, ".cursor", "mcp.json")

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

// EnsureRootGitignoreEntries manages D3-specific entries in the root .gitignore file.
// It reads the existing file if present, preserves user entries, and either updates
// or creates a D3-specific section marked with "# d3".
func (op *DefaultFileOperator) EnsureRootGitignoreEntries(fs ports.FileSystem, projectRootAbs string) error {
	gitignorePath := filepath.Join(projectRootAbs, ".gitignore")

	// These are the patterns we want to ensure are in the gitignore file
	d3Patterns := []string{
		".cursor/rules/d3/",          // d3 rules directory
		".cursor/rules/d3/*.gen.mdc", // generated rule files
		".d3/.feature",               // active feature marker
		".d3/features/*/.phase",      // phase markers
	}

	// Check if the gitignore file exists
	fileExists := false
	data, err := fs.ReadFile(gitignorePath)
	if err == nil {
		fileExists = true
	} else if !os.IsNotExist(err) {
		// Some error other than "file doesn't exist"
		return fmt.Errorf("failed to read .gitignore: %w", err)
	}

	var newContent []byte

	if !fileExists {
		// File doesn't exist, create a new one with our patterns
		var buffer bytes.Buffer
		buffer.WriteString(D3GitignoreSectionMarker + "\n")
		for _, pattern := range d3Patterns {
			buffer.WriteString(pattern + "\n")
		}
		newContent = buffer.Bytes()
	} else {
		// File exists, update or add the D3 section
		newContent = op.updateGitignoreContent(data, d3Patterns)
	}

	// Write back the file
	if err := fs.WriteFile(gitignorePath, newContent, 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

// updateGitignoreContent handles updating an existing .gitignore file
// by either updating the D3 section or appending it
func (op *DefaultFileOperator) updateGitignoreContent(existingContent []byte, d3Patterns []string) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(existingContent))
	var outputBuffer bytes.Buffer

	// State tracking
	inD3Section := false
	foundD3Section := false
	lineCount := 0

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// Check if this is the start of a D3 section
		if strings.TrimSpace(line) == D3GitignoreSectionMarker {
			inD3Section = true
			foundD3Section = true

			// Write the marker and our patterns
			outputBuffer.WriteString(D3GitignoreSectionMarker + "\n")
			for _, pattern := range d3Patterns {
				outputBuffer.WriteString(pattern + "\n")
			}
		} else if inD3Section {
			// Skip lines in the D3 section (we've already written our updated patterns)
			// Check if we're leaving the D3 section (blank line or new section marker)
			if strings.TrimSpace(line) == "" || (strings.HasPrefix(line, "#") && !strings.Contains(line, "D3")) {
				inD3Section = false
				outputBuffer.WriteString(line + "\n") // Include the line that ended the section
			}
		} else {
			// Not in D3 section, copy the line as is
			outputBuffer.WriteString(line + "\n")
		}
	}

	// If we didn't find a D3 section, append one at the end
	if !foundD3Section {
		// Add a blank line if the file doesn't end with one
		if lineCount > 0 && !strings.HasSuffix(string(existingContent), "\n\n") {
			outputBuffer.WriteString("\n")
		}

		// Add our D3 section
		outputBuffer.WriteString(D3GitignoreSectionMarker + "\n")
		for _, pattern := range d3Patterns {
			outputBuffer.WriteString(pattern + "\n")
		}
	}

	return outputBuffer.Bytes()
}

// EnsureProjectFiles creates the necessary project files in the .d3 directory
func (op *DefaultFileOperator) EnsureProjectFiles(fs ports.FileSystem, d3DirAbs string) error {
	// Create project.md
	projectMdPath := filepath.Join(d3DirAbs, "project.md")
	projectMdContent := ""

	// Create tech.md
	techMdPath := filepath.Join(d3DirAbs, "tech.md")
	techMdContent := ""

	// Write project.md
	if err := fs.WriteFile(projectMdPath, []byte(projectMdContent), 0644); err != nil {
		return fmt.Errorf("failed to write project.md: %w", err)
	}

	// Write tech.md
	if err := fs.WriteFile(techMdPath, []byte(techMdContent), 0644); err != nil {
		return fmt.Errorf("failed to write tech.md: %w", err)
	}

	return nil
}
