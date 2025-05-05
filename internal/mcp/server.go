// Package mcp implements Model Context Protocol server functionality for d3
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/mcp/tools"
	"github.com/imcclaskey/d3/internal/project"
	"github.com/imcclaskey/d3/internal/version"
)

// NewServer creates a new MCP server for d3
func NewServer(workspaceRoot string) *server.MCPServer {
	// Initialize project
	proj := project.New(workspaceRoot)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"d3 - Define, Design, Deliver!",
		version.Version,
		server.WithInstructions("d3 is a structured workflow engine for AI-driven development within Cursor"),
		server.WithToolCapabilities(true),
		// TODO: Integrate project with MCP service/tools here (e.g. pass to tool handlers)
	)

	// Register tools
	tools.RegisterTools(mcpServer, proj)

	return mcpServer
}

// ServeStdio starts the MCP server over stdio
func ServeStdio(s *server.MCPServer) error {
	return server.ServeStdio(s)
}
