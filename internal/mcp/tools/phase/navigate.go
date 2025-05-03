package phase

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/imcclaskey/d3/internal/core"
	"github.com/imcclaskey/d3/internal/core/session"
)

// navigateTool defines the d3_phase_navigate tool
var navigateTool = mcp.NewTool("d3_phase_navigate",
	mcp.WithDescription("Navigate to a different phase in the current feature"),
	mcp.WithString("to",
		mcp.Required(),
		mcp.Description("Target phase to navigate to (define, design, deliver)"),
	),
)

// handleNavigate returns a handler for the d3_phase_navigate tool
func handleNavigate(services *core.Services) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract target phase
		targetPhaseStr, ok := request.Params.Arguments["to"].(string)
		if !ok {
			return mcp.NewToolResultError("Target phase must be specified"), nil
		}

		// Parse target phase
		targetPhase, err := session.ParsePhase(targetPhaseStr)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get current context
		currentFeature, currentPhase, err := services.Session.GetContext()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get current context: %s", err.Error())), nil
		}

		// Ensure we have an active feature
		if currentFeature == "" {
			return mcp.NewToolResultError("No active feature. Use d3_create_feature or d3_enter_feature first."), nil
		}

		// Check for potential impact
		hasImpact := false
		if currentPhase != session.None {
			// Moving backward never has impact
			isBackward := isBackwardTransition(currentPhase, targetPhase)

			// Forward movement to phase with existing files might have impact
			if !isBackward && targetPhase.String() != currentPhase.String() {
				hasPhaseFiles := services.Files.HasPhaseFiles(currentFeature, targetPhase.String())
				hasImpact = hasPhaseFiles
			}
		}

		// Set the new phase
		if err := services.Session.SetPhase(targetPhase); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to navigate to %s: %s", targetPhase, err.Error())), nil
		}

		// Lazily create phase files
		newFiles, err := services.Files.EnsurePhaseFiles(currentFeature, targetPhase.String())
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create phase files: %s", err.Error())), nil
		}

		// Update the core rule file with the new context
		if err := services.Files.GenerateCoreRuleFile(currentFeature, targetPhase.String()); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update core rule file: %s", err.Error())), nil
		}

		// Build next steps based on context
		var nextSteps []string

		// Case 1: Impact detected (reconciliation needed)
		if hasImpact {
			nextSteps = append(nextSteps, "Review existing files for potential conflicts")
			nextSteps = append(nextSteps, fmt.Sprintf("Update %s phase files with information from previous phases", targetPhase))
		}

		// Case 2: New files created
		if len(newFiles) > 0 {
			nextSteps = append(nextSteps, fmt.Sprintf("Fill out newly created %s phase files:", targetPhase))
			for _, file := range newFiles {
				nextSteps = append(nextSteps, fmt.Sprintf("  - %s", file))
			}
		}

		// Add default next steps if empty
		if len(nextSteps) == 0 {
			nextSteps = append(nextSteps, fmt.Sprintf("Continue working in the %s phase", targetPhase))
		}

		// Remind to use d3_rule to get phase guidance
		nextSteps = append(nextSteps, "Use d3_rule to get phase-specific guidance")

		// Create base message
		message := fmt.Sprintf("Navigated to %s phase.", targetPhase)

		// Add notice about newly created files
		if len(newFiles) > 0 {
			message += fmt.Sprintf("\n\nCreated %d new files for this phase.", len(newFiles))
		}

		// Add impact warning
		if hasImpact {
			message += "\n\nNote: This phase transition may require reconciliation of existing files."
		}

		// Add next steps section
		if len(nextSteps) > 0 {
			message += "\n\n--- NEXT STEPS ---"
			for i, step := range nextSteps {
				message += fmt.Sprintf("\n%d. %s", i+1, step)
			}
		}

		return mcp.NewToolResultText(message), nil
	}
}

// isBackwardTransition checks if a phase transition is moving backward in the workflow
func isBackwardTransition(current, target session.Phase) bool {
	// Map phases to numeric values
	phaseValues := map[session.Phase]int{
		session.Define:  1,
		session.Design:  2,
		session.Deliver: 3,
	}

	currentValue, currentExists := phaseValues[current]
	targetValue, targetExists := phaseValues[target]

	if !currentExists || !targetExists {
		return false
	}

	return targetValue < currentValue
}
