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

The following refactoring steps are designed to be completed sequentially, with the aim of having a buildable state (`make build`) after each major step.

1.  **Establish `FileSystem` Port and Adapter:**
    *   In a shared package (e.g., `internal/core/ports`), define a `FileSystem` interface (e.g., `ports.FileSystem`) with common file operations (`Stat`, `ReadFile`, `WriteFile`, `MkdirAll`, `ReadDir`, `Create`).
    *   In the same package, implement a concrete `RealFileSystem` (e.g., `ports.RealFileSystem`) that wraps the standard `os` package functions.
    *   In the main application setup (e.g., `d3/main.go` or `internal/cli`), instantiate this `RealFileSystem`. This instance will be used for subsequent dependency injections.
    *(This step ensures the FileSystem provider is ready and the application builds).*

2.  **Inject `FileSystem` into `internal/core/session.Storage`:**
    *   Modify `session.NewStorage` (or equivalent constructor) to accept an argument of type `ports.FileSystem`.
    *   Update methods within `session.Storage` to use the injected `FileSystem` interface instead of direct `os` calls.
    *   Update the main application setup to pass the `RealFileSystem` instance (from Step 1) when creating `session.Storage`.
    *(This step refactors session.Storage and ensures the application builds).*

3.  **Inject `FileSystem` into `internal/core/feature.Service`:**
    *   Modify `feature.NewService` to accept `ports.FileSystem`.
    *   Update methods within `feature.Service` to use the injected `FileSystem`.
    *   Update the main application setup to pass the `RealFileSystem` instance when creating `feature.Service`.
    *(Ensures buildable state).*

4.  **Refactor `internal/core/rules.Service` for Dependency Injection:**
    *   Define a `Generator` interface *within* the `internal/core/rules` package (e.g., `type Generator interface { GenerateRuleStubs(...) ... }`) to represent the rule generation capabilities currently provided by `rules.RuleGenerator`.
    *   Ensure the existing concrete `rules.RuleGenerator` struct implements this new `rules.Generator` interface.
    *   Modify `rules.NewService` to accept `ports.FileSystem` and `rules.Generator` as arguments.
    *   Update methods within `rules.Service` to use the injected `FileSystem` and `Generator`.
    *   Update the main application setup:
        *   Instantiate `rules.RuleGenerator` (if it has its own dependencies, ensure they are provided or refactored similarly if they involve `os` calls).
        *   Pass the `RealFileSystem` instance and the `rules.RuleGenerator` instance (as `rules.Generator`) when creating `rules.Service`.
    *(Ensures buildable state).*

5.  **Refactor `internal/project.Project` for Dependency Injection and `FileSystem` Usage:**
    *   Define service-specific interfaces *within* the `internal/project` package for the services `Project` depends on (e.g., `type Storage interface { GetSession() ... }` for `session.Storage`, `type FeatureService interface { ... }`, `type RulesService interface { ... }`).
    *   Ensure the existing `internal/core/session.Storage`, `internal/core/feature.Service`, and `internal/core/rules.Service` structs (and their constructors) are compatible with and satisfy these new `project`-local interfaces. This might involve ensuring method signatures match.
    *   Modify `project.New` (or equivalent constructor for `Project`) to accept arguments for these `project`-local service interfaces (e.g., `project.Storage`, `project.FeatureService`, `project.RulesService`) AND an argument of type `ports.FileSystem`. The `FileSystem` is for `Project`'s direct use or for passing to callees like `EnsurePhaseFiles`.
    *   Update methods within `project.Project` to use the injected service interfaces and the injected `FileSystem`.
    *   Update the main application setup to:
        *   Pass the already-refactored `session.Storage`, `feature.Service`, and `rules.Service` instances (which now satisfy the `project`-local interfaces) when creating `project.Project`.
        *   Pass the `RealFileSystem` instance to `project.New`.
    *(Ensures buildable state).*

6.  **Inject `FileSystem` into `internal/core/phase.EnsurePhaseFiles`:**
    *   Modify the function signature of `phase.EnsurePhaseFiles` to accept `ports.FileSystem` as an argument.
    *   Update the implementation of `phase.EnsurePhaseFiles` to use the injected `FileSystem`.
    *   Update all call sites of `phase.EnsurePhaseFiles` (e.g., within `project.Project.ChangePhase`) to pass the required `ports.FileSystem` instance. (`project.Project` should now have access to `FileSystem` via DI from Step 5).
    *   **IMPLEMENTED:** Refactored `phase.EnsurePhaseFiles` to use a deterministic ordering of phases rather than relying on Go's map iteration order. This ensures consistent behavior across test runs and in production, utilizing a predefined slice of phases (`Define`, `Design`, `Deliver`). 
    *(Ensures buildable state).*

7.  **Implement Unit Tests for `internal/project`:**
    *   Write tests for `Project` methods, using `gomock`-generated mocks for its dependencies (`StorageService`, `FeatureServicer`, `RulesServicer`, `FileSystem`) passed via interfaces.

8.  **Implement Unit Tests for `internal/core/session`:**
    *   Write tests for `session.Storage` methods, using a `gomock`-generated mock `FileSystem`.
    *   Write tests for `Phase` methods and `ParsePhase`.

9.  **Implement Unit Tests for `internal/core/rules`:**
    *   Write tests for `rules.Service` methods, using `gomock`-generated mocks for `FileSystem` and `Generator` interfaces.
    *   (Optional: Write tests for `rules.RuleGenerator` methods separately).

10. **Implement Unit Tests for `internal/core/feature`:**
    *   Write tests for `feature.Service` methods, using a `gomock`-generated mock `FileSystem`.

11. **Implement Unit Tests for `internal/core/phase`:**
    *   Write tests for `phase.EnsurePhaseFiles`, using a `gomock`-generated mock `FileSystem`.

**11.5. Define `ProjectService` Interface and Refactor CLI/MCP for DI:**
    *   In `internal/project/project.go`, define a `ProjectService` interface encompassing methods needed by CLI commands and MCP tools.
    *   Ensure `project.Project` implements this interface.
    *   Use `gomock` to generate a mock for `ProjectService` (e.g., into `internal/project/mocks/`).
    *   Refactor CLI commands (e.g., `InitCommand`, `CreateCommand`) and MCP tool handlers to accept and use an instance of `ProjectService` for their core logic, allowing mock injection for tests.

12. **Implement Unit Tests for `internal/cli`:**
    *   Write tests for individual command logic (e.g., the `run` methods of `InitCommand`, `CreateCommand`), injecting a `gomock`-generated `MockProjectService`.
    *   Optionally, add integration-style tests for `cli.Execute` using Cobra's testing helpers if deemed necessary.

13. **Implement Unit Tests for `internal/mcp`:**
    *   Write tests for MCP tool handlers (e.g., `HandleMove`), injecting a `gomock`-generated `MockProjectService`.

14. **Integrate Tests into CI Pipeline:**
    *   Modify the relevant GitHub Actions workflow to add a step executing `go test ./...` before build/release jobs.
    *   **IMPLEMENTED:** Created a new CI workflow in `.github/workflows/ci.yml` that runs tests on pull requests and pushes to the main branch. Updated the existing release workflow to run tests before building. Both workflows use `make test` to ensure consistent test execution between local and CI environments.
    *   **ENHANCED:** Added essentials for test visibility in CI workflows:
        * Coverage summary displayed directly in GitHub workflow output
        * GitHub Actions workflow status badge in README to show build/test status
        * Coverage reports saved as artifacts for future reference

## 3. Technical Constraints & Requirements

*   **Go Standard Library:** Tests primarily rely on the standard Go `testing` package.
*   **Go Conventions:** Tests must follow standard Go testing conventions and best practices.
*   **CI Integration:** Tests MUST run successfully in the CI environment (GitHub Actions) before code is merged or deployed.
*   **Maintainability:** Tests should be clear, concise, and maintainable.
*   **Performance:** Unit tests should execute quickly.

## 4. Considerations

*   **Refactoring Impact:** The initial refactoring (Steps 1-6) touches several core packages and introduces new interfaces and DI patterns. While the aim is for each step to be buildable, careful implementation and verification at each stage are required.
*   **Mocking Approach:** Mocks are generated using the `gomock` library (`github.com/golang/mock/mockgen`) based on defined Go interfaces. `go:generate` directives are used to trigger mock generation.
*   **Test Coverage:** Goal is significant coverage, focusing on critical logic.
*   **Deterministic Behavior:** Special attention must be paid to ensuring deterministic behavior in code that might otherwise exhibit non-determinism (e.g., map iteration in Go). Functions should be designed to produce consistent outputs for consistent inputs to facilitate reliable testing.
