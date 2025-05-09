1.  **Problem Statement**
    *   The current state management for features and phases in the `d3` project is overly complex. It relies on a `state.yaml` file within each feature directory to store the feature's current phase and on in-memory state tracking within `project.go`. This dual system can lead to synchronization issues, increased cognitive load for developers, and makes the system harder to maintain and debug.

2.  **Feature Goals**
    *   Simplify phase state persistence by replacing the `state.yaml` file with a simpler, gitignored `.phase` file within each feature's directory.
    *   Eliminate the need for in-memory state tracking of the current feature and phase in `project.go`.
    *   Establish the file system (specifically, a `.feature` file at the project level and the `.phase` file within the active feature's directory) as the single source of truth for the current operational context.
    *   Reduce the overall complexity of state management logic within the `d3` framework.

3.  **Core Requirements**
    *   The system MUST store the current phase of a given feature in a plain text file named `.phase` located directly within that feature's directory (e.g., `.d3/features/feature-name/.phase`).
    *   The content of the `.phase` file MUST be the string name of the current phase (e.g., "define", "design", "deliver").
    *   The `.phase` file MUST be added to the project's `.d3/.gitignore` file to ensure it is not tracked by version control.
    *   The system MUST determine the globally active feature by reading the content of the `.feature` file located in the `.d3` directory.
    *   The system MUST determine the phase of the currently active feature by reading the content of the `.phase` file within that active feature's directory.
    *   All internal logic that currently relies on in-memory variables for `CurrentFeature` and `CurrentPhase` (e.g., in `project.go`) MUST be refactored to derive this information by reading from the `.feature` and relevant `.phase` files at the time it is needed.
    *   Commands related to feature and phase lifecycle (e.g., creating features, changing phases, entering features, exiting features) MUST be updated to write to and read from the new `.phase` and `.feature` file-based mechanism.
    *   The system MUST gracefully handle scenarios where the `.feature` file or a feature's `.phase` file is missing or contains invalid data (e.g., by reporting a clear error or defaulting to a known safe state if appropriate).

4.  **Scope Exclusions**
    *   Changes to the `feature.Service` or `phase.Service` interfaces beyond what is strictly necessary to support reading/writing the `.phase` file and removing the use of `state.yaml` for phase management.
    *   Alterations to the rule generation or refresh mechanisms, other than ensuring they correctly receive the feature and phase context derived from the new file-based state.
    *   Introduction of any new in-memory caching layers for feature or phase state; the explicit goal is to remove, not replace, reliance on in-memory state for this information.
    *   Modifications to user-facing CLI output formats or messages, unless directly resulting from changes in how state is determined (e.g., new error messages for missing state files).
    *   Any architectural refactoring of the `Project` struct or its dependencies not directly related to the removal of in-memory `CurrentFeature` and `CurrentPhase` fields and the logic that manages them.
    *   Management of any other configuration data that might currently reside in `state.yaml` files, if such data exists beyond just the phase. This initiative is solely focused on the phase information.
