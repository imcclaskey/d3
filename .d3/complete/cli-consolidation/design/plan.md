# Technical Design: CLI Consolidation & Initialization Enhancement

## 1. Technical Approach Overview

This feature will enhance the existing `d3 init` CLI command (found in `internal/cli/command/init.go`) to ensure robust project initialization and introduce a non-destructive "refresh" capability. The goal is to make `d3 init` the sole, reliable method for setting up and maintaining the d3 project structure, addressing the requirements outlined in `problem.md`.

The enhanced `d3 init` command will:
*   Confirm and utilize the project root (current working directory where the command is run).
*   Modify the existing `project.ProjectService.Init()` method (and potentially related structs/interfaces) to:
    *   Create/update `mcp.json` in the project root, correctly populating it with d3 server information and the determined working directory. Special care must be taken to preserve user-modified values in `mcp.json` (e.g., server port) during a refresh.
    *   Ensure the standard d3 project structure is present: `.d3/`, `.d3/features/`, `.d3/rules/`. This creation must be idempotent.
    *   Manage a default `.gitignore` file: create if non-existent, or append d3-specific entries if it exists and they are missing.
*   Introduce a `--refresh` flag to the `d3 init` command. When this flag is used:
    *   The command will check for the existence of standard files/directories and create any that are missing.
    *   It will *not* delete or overwrite existing user-generated content.
    *   It will *not* overwrite `mcp.json` entirely but will ensure essential d3 configuration is present, preserving user settings.
*   The existing `--clean` flag's behavior might need review to ensure it doesn't conflict with the refresh logic or user expectations. It currently implies removing existing files. We need to clarify its role or potentially deprecate it if `--refresh` covers the intended "ensure environment is correct" use case more safely.
*   The `mcp_tool_init` agent tool will be deprecated and its functionality replaced by guiding users to use the `d3 init` command.

## 2. Delivery Steps

1.  **[code]** Analyze the existing `internal/cli/command/init.go` and `project.ProjectService.Init()` implementation to fully understand current behavior, especially regarding the `--clean` flag and `mcp.json` handling.
2.  **[code]** Design and implement the `--refresh` flag logic within `InitCommand`.
3.  **[code]** Create `internal/core/projectfiles/operations.go`: Define `Config` type (for `mcp.json`) and constants. Implement exported helper functions `EnsureMCPJSON(...)` and `EnsureD3GitignoreEntries(...)` by moving and adapting logic previously in `ProjectService` private methods.
4.  **[code]** Update `ProjectService.Init()` (`internal/project/project.go`) to call the new exported functions from the `projectfiles` package. Adjust `internal/cli/command/init.go` if necessary for instantiation or parameter passing changes.
5.  **[code]** Write dedicated, thorough unit tests for `projectfiles.EnsureMCPJSON` and `projectfiles.EnsureD3GitignoreEntries` in `internal/core/projectfiles/operations_test.go`, comprehensively mocking `ports.FileSystem`.
6.  **[code]** Update and simplify tests for `ProjectService.Init()` (`TestProject_Init` in `internal/project/project_test.go`). Focus on orchestration, init modes (clean/refresh/new), and mocking interactions with direct dependencies (other services) and the key outcomes of `projectfiles` helper calls (via `p.fs` mocks).
7.  **[test]** Run all unit tests for the modified command and services.
8.  **[code]** Update any relevant documentation to reflect the enhanced `d3 init` command, its `--refresh` flag, and the clarified `--clean` behavior.
9.  **[code]** Modify the `mcp_tool_init` tool to guide the user to use `d3 init` CLI command.
10. **[verify]** Manually test `d3 init` (no flags) in a new directory.
11. **[verify]** Manually test `d3 init --refresh` in an existing project:
    *   With some standard d3 directories/files missing.
    *   With `mcp.json` having some user-set values (e.g., a custom port).
    *   With an existing `.gitignore` file.
12. **[verify]** Manually test `d3 init --clean` and document its exact behavior.
13. **[commit]** Commit changes related to `d3 init` command enhancements and tests.
14. **[commit]** Commit changes related to `mcp_tool_init` deprecation and documentation updates.

## 3. Technical Constraints & Requirements

*   The command must operate correctly from any subdirectory, identifying the project root as the directory where `d3 init` is invoked.
*   The `--refresh` capability MUST be non-destructive to user files and preserve relevant user-configured values in `mcp.json`.
*   Changes must be compatible with the existing Cobra CLI framework and project service structure.
*   Clear error messages should be provided for issues like permission errors.
*   The behavior of the `--clean` flag needs to be clearly defined and distinct from `--refresh`. If its current behavior (removing existing files) is deemed too destructive, it might be a candidate for deprecation or re-scoping.

## 4. Considerations & Alternatives

*   **`--clean` flag behavior**:
    *   Current `init.go` suggests it removes existing files. This is risky.
    *   Option 1: Keep as is, but clearly document it's destructive.
    *   Option 2: Change `--clean` to mean "reset d3-managed files to default state" without touching user content not explicitly managed by d3's core structure. This is complex to define.
    *   Option 3: Deprecate `--clean` in favor of a more explicit "reset" command if truly needed, or rely on users managing their own files if they want a "from scratch" state after initial init.
    *   *Recommendation*: For now, the priority is `--refresh`. The `--clean` flag's utility and safety should be carefully evaluated. If its current "remove existing files" is confirmed, it should be clearly distinct from the non-destructive refresh.
*   **`mcp.json` Merging Logic**: This is critical. The logic must intelligently merge necessary d3 configurations without overwriting user-specific settings like port numbers if the server was already configured or running. A simple overwrite is not acceptable for refresh.
*   **Future Extensions**:
    *   `d3 init --template <name>` could initialize with specific rule sets or feature templates.
    *   Integration with version checking for the d3 tool itself.
