package feature

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
	"github.com/imcclaskey/d3/internal/core/session"
)

// createTool defines the d3_create_feature tool
var createTool = mcp.NewTool("d3_create_feature",
	mcp.WithDescription("Create a new feature and set it as the current context"),
	mcp.WithString("name",
		mcp.Required(),
		mcp.Description("Name of the feature to create"),
	),
)

// handleCreate returns a handler for the d3_create_feature tool
func handleCreate(services *core.Services) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract feature name
		featureName, ok := request.Params.Arguments["name"].(string)
		if !ok || featureName == "" {
			return mcp.NewToolResultError("Feature name is required"), nil
		}

		// Call the core feature service to create the feature
		result, err := services.Feature.Create(ctx, featureName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create feature: %s", err.Error())), nil
		}

		// Update the core rule file with the new context
		initialPhase := session.Define.String()
		if err := services.Files.GenerateCoreRuleFile(featureName, initialPhase); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update core rule file: %s", err.Error())), nil
		}

		// Set up context information
		contextInfo := map[string]string{
			"Feature":       result.FeatureName,
			"Feature Path":  result.FeaturePath,
			"Current Phase": "define",
		}

		// Build next steps
		nextSteps := []string{
			"Fill out define.md to describe your feature idea",
			"Navigate to the next phase when ready with d3_phase_navigate",
			"Add relevant technical requirements and implementation details",
			"Use d3_rule to get phase-specific guidance",
		}

		// Format the message with context and next steps
		message := result.Message

		// Add context section
		message += "\n\n--- CONTEXT ---"
		for key, value := range contextInfo {
			message += fmt.Sprintf("\n%s: %s", key, value)
		}

		// Add next steps section
		message += "\n\n--- NEXT STEPS ---"
		for i, step := range nextSteps {
			message += fmt.Sprintf("\n%d. %s", i+1, step)
		}

		// Return the formatted result
		return mcp.NewToolResultText(message), nil
	}
}
