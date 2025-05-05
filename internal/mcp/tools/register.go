// Package tools provides tool registration and management for MCP
package tools

import (
	"github.com/imcclaskey/d3/internal/project"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all available d3 tools with the MCP server
func RegisterTools(mcpServer *server.MCPServer, proj *project.Project) {
	mcpServer.AddTool(MoveTool, HandleMove(proj))
	mcpServer.AddTool(CreateTool, HandleCreate(proj))
	mcpServer.AddTool(InitTool, HandleInit(proj))
}
