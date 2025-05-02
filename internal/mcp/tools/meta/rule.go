package meta

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/i3/internal/common"
	"github.com/imcclaskey/i3/internal/core/rules"
	"github.com/imcclaskey/i3/internal/core/session"
)

// ruleTool defines the i3_rule tool
var ruleTool = mcp.NewTool("i3_rule",
	mcp.WithDescription("Get phase-specific rules and guidance for i3 workflow"),
)

// handleRule returns a handler for the i3_rule tool
func handleRule() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get workspace root
		workspaceRoot, err := common.GetWorkspaceRoot()
		if err != nil {
			// Cannot proceed without workspace root
			return mcp.NewToolResultError(fmt.Sprintf("Failed to determine workspace root: %s", err.Error())), nil
		}

		// Create session manager
		sessionManager := session.NewManager(workspaceRoot)

		// Get current feature and phase using the session manager
		feature, phase, err := sessionManager.GetContext()

		// If there's an error or no active feature/phase, provide guidance
		if err != nil || feature == "" || phase == "" {
			message := "No active feature or phase detected."
			message += "\n\n--- NEXT STEPS ---"
			message += "\n1. Run i3_init to initialize i3 in your workspace"
			message += "\n2. Run i3_create_feature to create a new feature"

			return mcp.NewToolResultText(message), nil
		}

		// Create rule generator
		ruleGen := rules.NewRuleGenerator()

		// Get the rule content directly from the rules package
		ruleContent, err := ruleGen.GenerateRuleContent(
			phase.String(),
			feature,
		)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get rule content: %s", err.Error())), nil
		}

		// Add context to the rule content
		message := ruleContent
		message += "\n\n--- CONTEXT ---"
		message += fmt.Sprintf("\nfeature: %s", feature)
		message += fmt.Sprintf("\nphase: %s", phase)
		message += fmt.Sprintf("\nworkspace: %s", workspaceRoot)

		// Return the rule content with context
		return mcp.NewToolResultText(message), nil
	}
}
