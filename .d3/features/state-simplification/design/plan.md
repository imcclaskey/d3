# Technical Design: State Simplification

## 1. Technical Approach Overview

This plan outlines the technical steps to simplify state management within the d3 framework. The core idea is to move from the current `state.yaml` and in-memory state tracking towards a file-system-first approach for managing the active feature and its current phase.

*   **Phase Management**: Each feature directory (e.g., `.d3/features/my-feature/`) will contain a `.phase` file. This plain text file will store the name of the current phase (e.g., "define", "design", "deliver"). This replaces the phase information previously stored in `state.yaml`.
*   **Active Feature Tracking**: A single file named `.feature` in the `.d3/` directory will store the name of the currently active feature. This replaces the `.active_feature` file (or any other previous mechanism like `.session`).
*   **Source of Truth**: The `.feature` file and the `.phase` file within the active feature's directory will become the definitive source of truth for the current operational context (active feature and its phase).
*   **In-Memory State Removal**: The `Project` struct in `internal/project/project.go` will no longer maintain `CurrentFeature` and `CurrentPhase` in its in-memory `State`. Instead, these values will be read from the respective files (`.d3/.feature` and `.d3/features/<active_feature>/.phase`) on demand.
*   **Service Layer Adjustments**: The `FeatureServicer` (`internal/core/feature/feature.go`) will be updated to manage the `.phase` files and the `.d3/.feature` file. The `ProjectService` (`internal/project/project.go`) will be refactored to use these new `FeatureServicer` methods and remove its direct in-memory state management.
*   **Gitignore**: Both `.d3/.feature` and `*.phase` (within feature directories) will be gitignored.

## 2. Delivery Steps

### Phase 1: Core Feature Service Refactoring

1.  **[code]** Modify `FeatureServicer` (`internal/core/feature/feature.go`):
    *   Update `CreateFeature` to create a `.phase` file (e.g., with "define" as default content) instead of `state.yaml`.
    *   Implement `GetFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error)` to read the phase from `.d3/features/<featureName>/.phase`.
    *   Implement `SetFeaturePhase(ctx context.Context, featureName string, newPhase phase.Phase) error` to write the new phase to `.d3/features/<featureName>/.phase`.
    *   Ensure `DeleteFeature` removes the `.phase` file.
    *   Remove all logic related to `state.yaml` for phase management (parsing, writing, struct definitions if any).
    *   Implement `SetActiveFeature(featureName string) error` to write `featureName` to `.d3/.feature`.
    *   Implement `GetActiveFeature() (string, error)` to read the feature name from `.d3/.feature`.
    *   Implement `ClearActiveFeature() error` to delete or empty `.d3/.feature`.
2.  **[code]** Add/Update unit tests for `FeatureServicer` to cover new `.phase` and `.feature` file interactions, including edge cases (file not found, invalid content).
3.  **[test]** Run all unit tests for `internal/core/feature/`.
4.  **[commit]** Commit changes for "Refactor FeatureService for .phase and .feature files".

### Phase 2: Project Service Refactoring

1.  **[code]** Modify `ProjectService` (`internal/project/project.go`):
    *   Remove `CurrentFeature` and `CurrentPhase` fields from the internal `State` struct.
    *   Create internal helper functions within `project.go` (or as private methods on `Project`):
        *   `getActiveFeatureName(ctx context.Context) (string, error)` that calls `features.GetActiveFeature()`.
        *   `getActiveFeaturePhase(ctx context.Context) (phase.Phase, error)` that calls `getActiveFeatureName()` then `features.GetFeaturePhase()`.
        *   `getFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error)` that calls `features.GetFeaturePhase()`.
    *   Update `Init`: No longer clears/sets in-memory `CurrentFeature`/`CurrentPhase`. `features.ClearActiveFeature()` will be called.
    *   Update `CreateFeature`:
        *   No longer sets `p.state.CurrentFeature` or `p.state.CurrentPhase`.
        *   Calls `features.SetActiveFeature(featureName)` after successful feature creation by `features.CreateFeature` (which now creates `.phase`).
        *   `rules.RefreshRules` will now be called with feature and phase obtained via the new helper functions.
    *   Update `ChangePhase`:
        *   Use `getActiveFeatureName()` to get `currentFeatureName`.
        *   Use `getActiveFeaturePhase()` to get `currentInMemoryPhase` (rename variable to reflect source).
        *   Calls `features.SetFeaturePhase(ctx, currentFeatureName, targetPhase)`.
        *   No longer sets `p.state.CurrentPhase`.
        *   `rules.RefreshRules` will use new helpers.
    *   Update `EnterFeature`:
        *   Calls `features.SetActiveFeature(featureName)`.
        *   Phase is implicitly read by `rules.RefreshRules` via new helpers or directly if needed.
        *   No longer sets `p.state.CurrentFeature` or `p.state.CurrentPhase`.
    *   Update `ExitFeature`:
        *   Use `getActiveFeatureName()` to determine `exitedFeatureName`.
        *   Calls `features.ClearActiveFeature()`.
        *   No longer sets `p.state.CurrentFeature` or `p.state.CurrentPhase`.
    *   Update `DeleteFeature`:
        *   `features.DeleteFeature` will now handle clearing `.feature` if the deleted feature was active.
        *   No longer updates `p.state.CurrentFeature` or `p.state.CurrentPhase`.
        *   Logic for `rules.ClearGeneratedRules` if active context was cleared remains, but context determination changes.
2.  **[code]** Add/Update unit tests for `ProjectService` to reflect removal of in-memory state and reliance on `FeatureServicer` for state. Mock `FeatureServicer` calls appropriately.
3.  **[test]** Run all unit tests for `internal/project/`.
4.  **[commit]** Commit changes for "Refactor ProjectService to remove in-memory state".

### Phase 3: Rules Service Integration & Gitignore

1.  **[code]** Review `RulesServicer` (`internal/core/rules/rules.go`):
    *   Ensure `RefreshRules(featureName string, phaseName string)` is correctly called by `ProjectService` methods with feature/phase derived from file system. The signature of `RefreshRules` itself likely doesn't need to change, but its callers in `ProjectService` do.
2.  **[code]** Update `.d3/.gitignore` (or main `.gitignore`):
    *   Add entry: `/.feature` (to ignore `.d3/.feature`)
    *   Add entry: `features/*/.phase` (to ignore `.phase` files in any feature directory)
    *   Alternatively, if a single `.d3/.gitignore` is preferred, entries would be `.feature` and `features/*/.phase`.
3.  **[verify]** Manually check that `.d3/.feature` and feature-specific `.phase` files are correctly ignored by Git after changes.
4.  **[code]** Update relevant CLI command tests (`cmd/`) to ensure they work with the new state mechanism. This might involve adjusting test setup or assertions.
5.  **[test]** Run full integration test suite (e.g., `make test` or equivalent).
6.  **[commit]** Commit changes for "Integrate RulesService, update gitignore, and finalize tests".

### Phase 4: Cleanup and Verification

1.  **[code]** Globally search codebase for any remaining direct use of `state.yaml` for phase information and remove/refactor.
2.  **[code]** Remove old `state.yaml` files from existing feature directories in `.d3/features/` if any. (This might be a manual step or a small script if many exist, but for planning, we'll list it as code).
3.  **[verify]** Manually test key CLI flows:
    *   `d3 init`
    *   `d3 feature create test-feat`
    *   `d3 phase move design` (while in test-feat)
    *   `d3 feature enter other-feat` (assuming another feature exists or is created)
    *   `d3 feature exit`
    *   `d3 feature delete test-feat`
4.  **[commit]** Commit "Final cleanup and verification of state simplification".

## 3. Technical Constraints & Requirements

*   Must adhere to the definitions in `problem.md`.
*   All file reads/writes for `.feature` and `.phase` must be robust, with proper error handling for missing files or malformed content.
*   Changes should not negatively impact the performance of CLI operations noticeably. Direct file I/O for small state files is generally acceptable.
*   The `.phase` file content should be simple strings like "define", "design", "deliver".
*   The `.feature` file content should be the simple string name of the feature.

## 4. Considerations & Alternatives

*   **Alternative to direct file read**: Could introduce a lightweight caching layer in `ProjectService` if file reads become a performance concern, but the `problem.md` explicitly discourages new caching layers. We will proceed with direct reads first.
*   **Error Handling**: Define consistent error messages for scenarios like:
    *   `.d3/.feature` not found or empty (implies no active feature).
    *   `.d3/features/<active_feature>/.phase` not found or empty/invalid.
*   **Migration**: For existing projects, a one-time migration step might be needed to convert existing `state.yaml` to `.phase` files and create the `.feature` file. This plan focuses on the core changes; migration can be a follow-up task if deemed necessary for backward compatibility with very old project states. For now, we assume new projects or that `d3 init --clean` would reset to the new structure.
*   **Concurrency**: Current CLI operations are typically single-threaded. If d3 ever introduces concurrent operations modifying state, file-based locking or more robust transactionality might be needed. This is out of scope for the current simplification.
