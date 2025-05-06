# Technical Design Plan: Add Unit Tests

## 1. Technical Approach Overview

This feature implementation will establish a comprehensive unit testing suite by:

*   **Refactoring for Testability:** Core components are refactored to improve testability, primarily through Dependency Injection, focusing on abstracting direct file system access and injecting dependencies.
*   **Dependency Injection Pattern:** Go interfaces are defined for service dependencies (e.g., `FileSystem`, `RuleGenerator`) and system interactions. Components accept these interfaces via constructors or parameters, allowing mock/fake implementations during tests.
*   **Test Implementation Strategy:** Unit tests are added incrementally on a package-by-package basis after the necessary initial refactoring is complete. Tests cover happy paths and relevant edge cases.
*   **Testing Framework:** The standard Go `testing` package is used. Mocks/fakes are implemented using standard Go interfaces.
*   **Go Testing Conventions:** Adhering to standard Go testing practices (`*_test.go`, `TestXxx`, `t.Run`, table-driven tests).
*   **CI Integration:** Test execution is integrated into the CI pipeline (GitHub Actions).

## 2. Implementation Steps

1.  **Refactor Core Components for Testability:**
    *   **Define `FileSystem` Interface:** Create an interface in a shared package (e.g., `internal/core/ports/FileSystem`) defining common file operations (`Stat`, `ReadFile`, `WriteFile`, `MkdirAll`, `ReadDir`, `Create`).
    *   **Implement Real `FileSystem`:** Create a concrete implementation of `FileSystem` (e.g., in `internal/core/ports`) that wraps the standard `os` package functions.
    *   **Refactor `internal/core/session.Storage`:** Modify `NewStorage` to accept the `ports.FileSystem` interface. Update methods to use the interface.
    *   **Refactor `internal/core/feature.Service`:** Modify `NewService` to accept the `ports.FileSystem` interface. Update methods to use the interface.
    *   **Define `RuleGenerator` Interface:** Create an interface *within* the `internal/core/rules` package, matching the methods of `rules.RuleGenerator`.
    *   **Refactor `internal/core/rules.Service`:** Modify `NewService` to accept `ports.FileSystem` and the local `rules.RuleGenerator` interface. Update `RefreshRules`.
    *   **Refactor `internal/core/phase.EnsurePhaseFiles`:** Modify the function signature to accept the `ports.FileSystem` interface. Update implementation.
    *   **Define Service Interfaces in `project`:** Define interfaces *within* the `internal/project` package for the dependencies `Project` needs (e.g., `project.SessionStorage`, `project.RuleService`, `project.FeatureService`).
    *   **Refactor `internal/project.Project`:** Modify `New` to accept these `project`-local interfaces. Modify `ChangePhase` to handle passing the `FileSystem` interface down.
    *   **Update Service Implementations:** Ensure `session.Storage`, `rules.Service`, `feature.Service` structs implement the corresponding interfaces defined in the `project` package.
    *   **Update Instantiation:** Modify main application setup (`d3/main.go` or `internal/cli`) to create the real `FileSystem`, wrap concrete services if necessary to satisfy interfaces, and inject dependencies correctly.
2.  **Implement Unit Tests for `internal/project`:**
    *   Write tests for `Project` methods, using mock/fake implementations of `SessionStorage`, `RuleService`, `FeatureService`, and potentially `FileSystem` passed via interfaces.
3.  **Implement Unit Tests for `internal/core/session`:**
    *   Write tests for `session.Storage` methods, using a mock `FileSystem`.
    *   Write tests for `Phase` methods and `ParsePhase`.
4.  **Implement Unit Tests for `internal/core/rules`:**
    *   Write tests for `rules.RuleGenerator` methods.
    *   Write tests for `rules.Service` methods, using mock `FileSystem` and `RuleGenerator`.
5.  **Implement Unit Tests for `internal/core/feature`:**
    *   Write tests for `feature.Service` methods, using a mock `FileSystem`.
6.  **Implement Unit Tests for `internal/core/phase`:**
    *   Write tests for `phase.EnsurePhaseFiles`, using a mock `FileSystem`.
7.  **Implement Unit Tests for `internal/cli`:**
    *   Write tests for individual commands (e.g., `internal/cli/command/create.go`), likely mocking the `Project` service they interact with.
    *   Optionally, add integration-style tests for `cli.Execute` using Cobra's testing helpers if deemed necessary.
8.  **Implement Unit Tests for `internal/mcp`:**
    *   Write tests for MCP tool handlers (e.g., `HandleMove`), passing in a mock `Project`.
9.  **Integrate Tests into CI Pipeline:**
    *   Modify the relevant GitHub Actions workflow to add a step executing `go test ./...` before build/release jobs.

## 3. Technical Constraints & Requirements

*   **Go Standard Library:** Tests primarily rely on the standard Go `testing` package.
*   **Go Conventions:** Tests must follow standard Go testing conventions and best practices.
*   **CI Integration:** Tests MUST run successfully in the CI environment (GitHub Actions) before code is merged or deployed.
*   **Maintainability:** Tests should be clear, concise, and maintainable.
*   **Performance:** Unit tests should execute quickly.

## 4. Considerations

*   **Refactoring Impact:** The initial refactoring (Step 1) touches several core packages and introduces new interfaces. Careful implementation and testing are required.
*   **Mocking Approach:** Standard Go interfaces with hand-written fakes/mocks are the primary mechanism. Avoid external mocking libraries.
*   **Test Coverage:** Goal is significant coverage, focusing on critical logic.
