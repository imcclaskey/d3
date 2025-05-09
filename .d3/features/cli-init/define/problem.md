# Problem Statement

The existing `d3 init` CLI command (found in `internal/cli/command/init.go`) provides a baseline for initializing a d3 workspace. However, it needs enhancement to:
1.  Explicitly manage and correctly populate an `mcp.json` (or equivalent central configuration) file, ensuring the `workdir` accurately reflects the project's root (the directory where `init` is run).
2.  Offer a clear "refresh" mechanism that is non-destructive by default, ensuring all standard d3 files and directories are present without deleting user-generated content. This needs to be distinct from or an evolution of the current `--clean` flag's behavior.
3.  Fully consolidate all d3 environment initialization and setup logic, absorbing any responsibilities currently handled by the separate MCP `init` tool (though the MCP tool has no extra features not intended for the CLI), to provide a single, comprehensive CLI entry point.

# Feature Goals

*   Enhance the existing `d3 init` CLI command to be the definitive tool for d3 environment initialization and refresh.
*   Ensure the `d3 init` command correctly creates/updates an `mcp.json` file in the project root, with `workdir` set to the absolute path of this root directory.
*   Implement a "refresh" capability as the default behavior of `d3 init` (when not using `--clean`) that idempotently creates missing d3 scaffolding (files, directories) without deleting user content.
*   Refine the behavior of the `--clean` flag to ensure it correctly handles `mcp.json` and maintains its existing role for a clean slate, while integrating the new `mcp.json` logic.
*   Migrate all functionalities of the current MCP `init` tool into the `d3 init` CLI command, paving the way for the MCP tool's deprecation for this purpose.

# Core Requirements

*   The `d3 init` CLI command (as defined in `internal/cli/command/init.go`) MUST be the primary interface for initializing d3 environments.
*   The command MUST operate on the current working directory (as it currently does via `os.Getwd()`).
*   `d3 init` MUST create or update an `mcp.json` file in the project's root directory (where `init` is run).
*   The `mcp.json` file MUST contain a `workdir` key, set to the absolute path of the project root.
*   The `mcp.json` file MUST contain any other necessary d3 server/project configuration.
*   By default (i.e., without `--clean`), `d3 init` MUST act as a "refresh":
    *   It MUST create all standard d3 directory structures (e.g., `.d3/features`, `.cursor/rules/d3`) if they do not exist.
    *   It MUST NOT create feature-specific phase files (e.g., `problem.md`, `solution.md`) within the feature directory structures.
    *   It MUST create or update standard d3 configuration files (e.g., `.gitignore` with d3-specific entries) if they do not exist or are missing required entries.
    *   It MUST NOT delete or overwrite existing user-generated files or content within the standard d3 directories.
*   The `--clean` flag functionality MUST maintain its existing behavior of removing d3-managed files and structures, and additionally:
    *   It SHOULD remove existing d3-managed structures (e.g., `.d3` directory, `.cursor/rules/d3` directory) and d3-specific file entries (like those in `.gitignore`) before performing a fresh initialization.
    *   If an `mcp.json` file exists, `--clean` SHOULD overwrite any existing d3-related entries or the entire file if it's primarily d3-managed, with the new configuration, without further prompting.
    *   The `d3 init` command (especially with `--clean`) MUST incorporate all setup logic currently handled by the MCP `init` tool.

# Scope Exclusions

*   This feature will not manage the installation or lifecycle of the d3 server itself.
*   This feature will not handle migration of existing projects from other frameworks.
*   This feature will not provide UI/GUI-based initialization (it's a CLI command).

# (Optional) Unresolved Dependencies/Questions

*   What is the definitive list of all files, directories (beyond not creating phase files), and default contents (e.g., for `.gitignore`, `mcp.json` structure) that `d3 init` should manage for both fresh setup and refresh?
*   How should `d3 init` specifically behave if an `mcp.json` file exists but is malformed or contains conflicting/partial server information, *when not using the `--clean` flag*? (e.g., error out, attempt to merge, overwrite specific fields only).
*   Does the `projectSvc.Init(clean)` method (currently called by `init.go`) already handle the creation of all necessary files and directories (respecting the 'no phase files' rule and other specific file content requirements), or will that logic need to be added/modified there or in the command itself?
