package tools

import (
	"context"

	"github.com/imcclaskey/d3/internal/project"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// InitTool defines the d3_init tool
var InitTool = mcp.NewTool("d3_init",
	mcp.WithDescription("Initialize d3 in the current workspace. This tool now guides to use the CLI."),
	mcp.WithBoolean("clean", mcp.Description("Perform a clean initialization (CLI only)")),
)

// HandleInit returns a handler for the d3_init tool
func HandleInit(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Guide user to CLI
		guidanceMessage := "To initialize d3 in your project, please run the `d3 init` command in your terminal. You can use flags like `--clean` or `--refresh` as needed. For example: `d3 init --refresh`"
		return mcp.NewToolResultText(guidanceMessage), nil
	}
}
