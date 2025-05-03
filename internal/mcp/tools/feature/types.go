// Package feature implements d3 feature-related tools
package feature

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
)

// RegisterTools registers all feature-related tools with the MCP server
func RegisterTools(mcpServer *server.MCPServer, services *core.Services) {
	// Register the create feature tool
	mcpServer.AddTool(createTool, handleCreate(services))

	// TODO: Add switch feature tool
	// TODO: Add list features tool
	// TODO: Add current feature tool
}
