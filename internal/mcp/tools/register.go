package tools

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
	"github.com/imcclaskey/d3/internal/mcp/tools/artifact"
	"github.com/imcclaskey/d3/internal/mcp/tools/feature"
	"github.com/imcclaskey/d3/internal/mcp/tools/meta"
	"github.com/imcclaskey/d3/internal/mcp/tools/phase"
)

// RegisterAllTools registers all d3 tools with the MCP server
func RegisterAllTools(mcpServer *server.MCPServer, services *core.Services) {
	// Register feature tools
	feature.RegisterTools(mcpServer, services)

	// Register phase tools
	phase.RegisterTools(mcpServer, services)

	// Register artifact tools
	artifact.RegisterTools(mcpServer, services)

	// Register meta tools
	meta.RegisterTools(mcpServer, services)
}
