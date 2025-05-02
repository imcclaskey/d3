// Package mcp implements Model Context Protocol server functionality for i3
package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/i3/internal/core"
	"github.com/imcclaskey/i3/internal/mcp/tools"
	"github.com/imcclaskey/i3/internal/version"
)

// Server represents an MCP server for i3
type Server struct {
	mcpServer *server.MCPServer
	services  *core.Services
	toolMgr   *tools.ToolManager
}

// NewServer creates a new MCP server for i3
func NewServer(workspaceRoot string) *Server {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"i3 - Ideate, Instruct, Implement!",
		version.Version,
		server.WithInstructions("i3 is a structured workflow engine for AI-driven development within Cursor"),
		server.WithToolCapabilities(true),
	)

	// Create core services
	services := core.NewServices(workspaceRoot)

	// Create tool manager
	toolMgr := tools.NewToolManager(services)

	// Create i3 server
	s := &Server{
		mcpServer: mcpServer,
		services:  services,
		toolMgr:   toolMgr,
	}

	// Register tools
	s.registerTools()

	return s
}

// registerTools registers all tools with the MCP server
func (s *Server) registerTools() {
	// Register all tools
	tools.RegisterAllTools(s.mcpServer, s.services)
}

// Serve starts the MCP server
func (s *Server) Serve() error {
	return server.ServeStdio(s.mcpServer)
}
