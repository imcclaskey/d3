package tools

import (
	"context"
	"fmt"

	"github.com/imcclaskey/d3/internal/project"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// InitTool defines the d3_init tool
var InitTool = mcp.NewTool("d3_init",
	mcp.WithDescription("Initialize d3 in the current workspace"),
	mcp.WithBoolean("clean", mcp.Description("Perform a clean initialization")),
)

// HandleInit returns a handler for the d3_init tool
func HandleInit(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract clean flag
		cleanFlag := false
		if cleanVal, ok := request.Params.Arguments["clean"].(bool); ok {
			cleanFlag = cleanVal
		}

		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Call project Init
		result, err := proj.Init(cleanFlag)
		if err != nil {
			// For other errors, return them as system errors
			return mcp.NewToolResultError(fmt.Sprintf("System error initializing project: %v", err)), nil
		}

		// Return formatted result
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}
