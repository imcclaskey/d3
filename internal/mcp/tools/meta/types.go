// Package meta implements d3 meta and utility tools
package meta

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
)

// RegisterTools registers all meta and utility tools with the MCP server
func RegisterTools(mcpServer *server.MCPServer, services *core.Services) {
	mcpServer.AddTool(initTool, handleInit(services))
	mcpServer.AddTool(ruleTool, handleRule())
	// TODO: Add refresh tool
	// TODO: Add summarize tool
}
