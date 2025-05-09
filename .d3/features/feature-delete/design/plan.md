# Technical Design: Feature Deletion

## 1. Technical Approach Overview

The feature deletion capability will be implemented by:

*   Introducing a new CLI command (e.g., `d3 feature delete <feature-name>`).
*   Adding a corresponding MCP tool action to trigger feature deletion.
*   Augmenting the existing `internal/core/feature/feature.go:Service` with a method for the actual deletion logic. This service will:
    *   Identify the feature's directory based on its name (e.g., `.d3/features/<feature-name>`).
    *   Perform a sanity check to ensure the target is a valid feature directory.
    *   Remove the entire feature directory and its contents.
    *   Provide clear success or failure feedback.
*   Ensuring user confirmation is obtained before any destructive action is taken, both in CLI and MCP contexts.

Key architectural decisions:
*   Centralize deletion logic by adding it to the existing `feature.Service` to ensure consistency and maintainability.
*   Deletion will be a direct file system operation (recursive removal of the feature directory). No "soft delete" or archiving will be implemented, as per `problem.md`.

System components affected:
*   CLI application: A new command will be added.
*   MCP toolset: A new tool or an extension to an existing tool will be required.
*   Internal core logic: A new method within `feature.Service` for identifying and removing feature directories.
*   No direct database interaction is anticipated as feature definitions are primarily file-system based.

## 2. Delivery Steps

1.  **[code]** Add a `DeleteFeature` method to the existing `feature.Service` in `internal/core/feature/feature.go`. This method will take a feature name, locate its directory (e.g., `.d3/features/<feature-name>`), and remove it. It should include comprehensive error handling (e.g., feature not found, permission issues, attempting to delete the active feature if applicable). Write unit tests for this new method.
2.  **[test]** Run unit tests for the feature deletion service.
3.  **[code]** Implement the CLI command `d3 feature delete <feature-name>`.
    *   This command will call the core deletion service.
    *   It must include a confirmation prompt (e.g., "Are you sure you want to delete feature 'X'? This action cannot be undone. [y/N]").
    *   It should provide clear output on success or failure.
    *   Write integration tests for the CLI command.
4.  **[test]** Run integration tests for the CLI delete command.
5.  **[code]** Implement the MCP tool integration for feature deletion.
    *   This will likely involve creating a new `mcp_d3_d3_feature_delete` tool.
    *   The MCP tool will invoke the same core deletion service.
    *   It must include a confirmation mechanism within the MCP interface.
    *   It should provide clear visual feedback in the MCP interface.
    *   Write integration tests for the MCP tool.
6.  **[test]** Run integration tests for the MCP delete tool.
7.  **[verify]** Manually test CLI deletion:
    *   Create a dummy feature.
    *   Attempt to delete it, confirming the prompt. Verify directory removal.
    *   Attempt to delete a non-existent feature. Verify error message.
    *   Attempt to cancel deletion at the prompt. Verify no action is taken.
8.  **[verify]** Manually test MCP tool deletion:
    *   Create a dummy feature.
    *   Attempt to delete it via MCP, confirming the prompt. Verify directory removal.
    *   Attempt to delete a non-existent feature via MCP (if applicable). Verify error feedback.
    *   Attempt to cancel deletion at the MCP prompt. Verify no action is taken.
9.  **[commit]** Commit all changes related to the feature deletion implementation (core service, CLI, MCP).
10. **[code]** Review and update any relevant documentation (e.g., CLI help text, user guides if they exist) to reflect the new feature deletion capability.
11. **[commit]** Commit documentation updates.

## 3. Technical Constraints & Requirements

*   **Performance:** Deletion should be reasonably fast for typical feature sizes. Since it's primarily a file system operation, performance is not expected to be a major constraint.
*   **Compatibility:** Ensure no impact on other CLI commands or MCP tools.
*   **Security:** The operation removes files. Ensure it only targets valid feature directories within the `.d3/features/` path to prevent accidental deletion of other system files. Path validation and sanitization are crucial.
*   **Error Handling:**
    *   Feature not found: Clear error message.
    *   Permissions issues: Clear error message if the system cannot delete the directory/files.
    *   Attempting to delete an "active" or "current" feature (if such a concept exists and is relevant): This needs clarification. For now, assume any feature can be deleted if not active. If a feature can be "active" (e.g. current context for `d3 feature enter`), attempting to delete it should probably be disallowed or require exiting the feature first.
*   **Idempotency:** While not strictly required to be idempotent (deleting a non-existent feature is an error), the system should handle such cases gracefully.

## 4. Considerations & Alternatives

*   **Alternative - Soft Delete/Archive:** `problem.md` explicitly excludes this. If requirements change, this would involve moving the feature directory to an archive location instead of permanent deletion, potentially with metadata about the deletion.
*   **Alternative - Feature "State":** Instead of direct deletion, a feature could be marked as "deleted" or "archived" in some metadata file. This was excluded by `problem.md`.
*   **Future Extensions:**
    *   Batch deletion of features.
    *   Role-based access control for deletion (currently out of scope).
*   **Technical Debt:** None anticipated if implemented cleanly. Ad-hoc deletion scripts without a proper service layer would incur technical debt.
*   **Impact on "Current Feature" Context:** If a user deletes the feature they are currently "in" (via `d3 feature enter`), the system needs to handle this gracefully. The CLI/MCP might need to clear the current feature context or prevent deletion of the active feature. This interaction needs to be defined. For instance, the `mcp_d3_d3_feature_exit`