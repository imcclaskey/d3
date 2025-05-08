# Technical Design: Enter-Exit Feature

## 1. Technical Approach Overview
*High-level summary of the implementation strategy*
*Key architectural decisions and their rationale*
*System components affected and their integration points*

## 2. Implementation Steps
*   **Task 1: Core State & Feature Service Refactoring for Robust Phase Management**
    *   **Task 1.1:** (Completed) Rename MCP tool `d3_create_feature` to `d3_feature_create` and CLI command `d3 create feature` to `d3 feature create`.
    *   **Task 1.2:** (Completed) Refactor `session.SessionState` (persisted in `session.yml` via `session.Storage`):
        *   Remove the `CurrentPhase` field. `session.yml` now only stores `CurrentFeature`.
        *   `session.Storage.Load()` and `session.Storage.Save()` in `internal/core/session/session.go` adjusted.
    *   **Task 1.3: Enhance `FeatureServicer` for `state.yaml` Management**
        *   In `internal/project/project.go` (where `FeatureServicer` interface is defined):
            *   Add method: `GetFeaturePhase(ctx context.Context, featureName string) (session.Phase, error)`
            *   Add method: `SetFeaturePhase(ctx context.Context, featureName string, phase session.Phase) error`
        *   In `internal/core/feature/feature.go` (implementation of `feature.Service`):
            *   Define a local struct (e.g., `featureStateData { LastActivePhase session.Phase \`yaml:\"active_phase\"\` }`) for `state.yaml`.
            *   Implement `GetFeaturePhase`: Reads `active_phase` from `.d3/features/<featureName>/state.yaml`. If `state.yaml` doesn't exist, it creates it with a default phase (e.g., "define") and returns that phase.
            *   Implement `SetFeaturePhase`: Writes the provided phase to `.d3/features/<featureName>/state.yaml` as `active_phase`.
            *   Modify the existing `CreateFeature` method in `feature.Service`: Ensure it internally creates the `.d3/features/<featureName>/state.yaml` file with a default phase (e.g., "define") upon successful feature directory creation.
    *   **Task 1.4: Refactor `Project.ChangePhase` to use `FeatureServicer`**
        *   Modify `Project.ChangePhase` in `internal/project/project.go`:
            *   It will call `p.features.SetFeaturePhase(p.state.CurrentFeature, targetPhase)` to persist the phase change to the feature's `state.yaml`.
            *   It will update the in-memory `p.state.CurrentPhase`.
            *   It will continue to save `session.yml` (via `p.session.Save()`) to update `LastModified` and ensure `CurrentFeature` is correctly persisted.
    *   **Task 1.5: Refactor `Project.CreateFeature` to use enhanced `FeatureServicer`**
        *   Modify `Project.CreateFeature` in `internal/project/project.go`:
            *   The call to `p.features.CreateFeature(...)` (the service method) is now responsible for creating the feature directory AND its initial `state.yaml`.
            *   Ensure `Project.CreateFeature` correctly sets the in-memory `p.state.CurrentFeature` (to new feature name) and `p.state.CurrentPhase` (to "define", or whatever the service defaults to).
            *   Ensure it updates `session.yml` (via `p.session.Save()`) to set `CurrentFeature = newFeatureName`.
    *   **Task 1.6:** Update all relevant internal code references and documentation to reflect these naming changes and significant core state/feature service refactorings.
    *   **Task 1.7:** Run all existing tests. Update tests and add new ones to cover the refactored `FeatureServicer` methods, and the modified behaviors of `Project.CreateFeature` and `Project.ChangePhase`.

*   **Task 2: Implement Core `EnterFeature` Logic (using enhanced `FeatureServicer`)**
    *   **Task 2.1:** Define/Modify the `EnterFeature(ctx context.Context, featureName string) (*FeatureStatus, error)` function in `internal/project/project.go`. This function will:
        *   Call `phase, err := p.features.GetFeaturePhase(ctx, featureName)` to get the feature's last active phase (this will also handle creating `state.yaml` with a default if it's a new/bare feature directory).
        *   If successful, update `session.yml` (via `p.session.Save()`) to set `CurrentFeature = featureName`.
        *   Update in-memory `p.state.CurrentFeature = featureName` and `p.state.CurrentPhase = phase`.
        *   Implement logic to load/activate d3 rules.
        *   Define and return a `FeatureStatus` struct or an error.

*   **Task 3: Create MCP Tool `d3_feature_enter`** (No change in definition, relies on updated Task 2)
*   **Task 4: Create CLI Command `d3 feature enter <feature_name>`** (No change in definition, relies on updated Task 3)

*   **Task 5: Implement Core `ExitFeature` Logic (relying on refactored state)**
    *   **Task 5.1:** Define/Modify `ExitFeature(ctx context.Context) (*Result, error)` in `internal/project/project.go`. This function will:
        *   Get `activeFeatureName = p.state.CurrentFeature`. If no feature is active, return.
        *   Update `session.yml` (via `p.session.Save()`) to set `CurrentFeature = ""`.
        *   Clear in-memory `p.state.CurrentFeature` and `p.state.CurrentPhase`.
        *   Update/clear d3 rules.
        *   Return a `Result` struct or an error.

*   **Task 6: Create MCP Tool `d3_feature_exit`** (No change in definition, relies on updated Task 5)
*   **Task 7: Create CLI Command `d3 exit`** (No change in definition, relies on updated Task 6)

*   **Task 8: Add Unit Tests for New Commands and Tool Handlers**
    *   **Task 8.1:** Add tests in `internal/mcp/tools/tools_test.go` for the `HandleFeatureEnter` tool handler.
        *   Verify correct parameter extraction (`feature_name`).
        *   Verify correct call to `ProjectService.EnterFeature`.
        *   Verify correct formatting of success and error results.
    *   **Task 8.2:** Add tests in `internal/mcp/tools/tools_test.go` for the `HandleFeatureExit` tool handler.
        *   Verify correct call to `ProjectService.ExitFeature`.
        *   Verify correct formatting of success and error results.
    *   **Task 8.3:** Create `internal/cli/command/feature_enter_test.go` and add tests for `NewFeatureEnterCommand`.
        *   Verify command setup (Use, Short, Args).
        *   Test the `run` logic, mocking the MCP call placeholder.
    *   **Task 8.4:** Create `internal/cli/command/exit_test.go` and add tests for `NewExitCommand`.
        *   Verify command setup (Use, Short, Args).
        *   Test the `run` logic, mocking the MCP call placeholder.

*Sequenced list of concrete implementation tasks for enter-exit*

## 3. Technical Constraints & Requirements
*Performance expectations or requirements*
*Compatibility and dependency considerations*
*Security, privacy, or compliance requirements*
*Critical error handling and edge cases*

## 4. Considerations & Alternatives
*Briefly discuss alternative approaches considered*
*Note potential future extensions or improvements*
*Highlight technical debt implications*