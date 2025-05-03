// Package phase implements d3 phase-related tools
package phase

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
)

// RegisterTools registers all phase-related tools with the MCP server
func RegisterTools(mcpServer *server.MCPServer, services *core.Services) {
	mcpServer.AddTool(navigateTool, handleNavigate(services))
	// TODO: Add current phase tool
	// TODO: Add phase guidance tool
}
