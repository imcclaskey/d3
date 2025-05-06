package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/project"
)

// CreateTool defines the d3_create_feature tool
var CreateTool = mcp.NewTool("d3_create_feature",
	mcp.WithDescription("Create a new feature and set it as the current context"),
	mcp.WithString("name",
		mcp.Required(),
		mcp.Description("Name of the feature to create"),
	),
)

// HandleCreate returns a handler for the d3_create_feature tool
// It now accepts project.ProjectService interface for testability.
func HandleCreate(proj project.ProjectService) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract feature name
		featureName, ok := request.Params.Arguments["name"].(string)
		if !ok || featureName == "" {
			return mcp.NewToolResultError("Feature name 'name' is required"), nil
		}

		// Check if project is valid
		if proj == nil {
			return mcp.NewToolResultError("Internal error: Project context is nil"), nil
		}

		// Call the project method to create the feature
		result, err := proj.CreateFeature(ctx, featureName)
		if err != nil {
			if err == project.ErrNotInitialized {
				return mcp.NewToolResultError("Cannot create feature: project not initialized"), nil
			}
			// For other errors, return them as system errors to provide more detail to the MCP client if needed.
			return mcp.NewToolResultError(fmt.Sprintf("System error creating feature: %v", err)), nil
		}

		// Return the formatted result
		return mcp.NewToolResultText(result.FormatMCP()), nil
	}
}
