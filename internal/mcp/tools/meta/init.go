package meta

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
)

// initTool defines the d3_init tool
var initTool = mcp.NewTool("d3_init",
	mcp.WithDescription("TEST Initialize d3 in the current workspace and create base project files"),
	mcp.WithBoolean("clean",
		mcp.Description("Perform a clean initialization (remove existing files)"),
	),
)

// handleInit returns a handler for the d3_init tool
func handleInit(services *core.Services) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract clean flag
		cleanFlag := false
		if cleanVal, ok := request.Params.Arguments["clean"].(bool); ok {
			cleanFlag = cleanVal
		}

		// Use the Files service to initialize the workspace
		message, newlyCreated, err := services.Files.InitWorkspace(cleanFlag)
		if err != nil {
			// Return clean error without diagnostics that might break JSON
			return mcp.NewToolResultError(fmt.Sprintf("Failed to initialize workspace: %v", err)), err
		}

		// Format a clean, concise result message
		resultMsg := message
		if len(newlyCreated) > 0 {
			resultMsg = fmt.Sprintf("%s\nCreated files: %v", message, newlyCreated)
		}

		// Return a clean result with just the essential information
		return mcp.NewToolResultText(resultMsg), nil
	}
}
