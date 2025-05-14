# d3 - Define, Design, Deliver!

<div align="center">
  <img src="https://img.shields.io/badge/status-alpha-orange" alt="Status: Alpha">
  <img src="https://img.shields.io/badge/language-go-blue" alt="Language: Go">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT">
  <img src="https://github.com/imcclaskey/d3/workflows/CI/badge.svg" alt="CI Status">
</div>

<p align="center">
  <b>A structured workflow engine for AI-driven development within Cursor</b>
</p>

---

## 🧠 What is d3?

d3 is a CLI tool and Model Context Protocol (MCP) server designed to orchestrate intentional, AI-driven development workflows within environments like [Cursor](https://cursor.sh). By providing a structured, phase-based process managed via MCP tools, d3 acts as an agent control system, enhancing the AI pair programming experience. It guides the AI through distinct phases of software development: defining the problem, designing the solution, and delivering the code.

### Core Benefits

- **Structured AI Collaboration**: Move beyond chaotic, ad-hoc AI interactions to a predictable, phase-based workflow guided by the MCP server.
- **Separation of Concerns**: Keep problem definition (`define`), solution planning (`design`), and implementation (`deliver`) distinct and focused within each feature.
- **Optimized AI Context**: Provide the right context and rules to the AI agent at each phase via MCP for better, more relevant outcomes.
- **Consistent Documentation**: Automatically generate and maintain technical documentation (`problem.md`, `plan.md`, `progress.yaml`) as you build.

## 🚀 Quick Start

### Prerequisites

- An MCP-compatible client (like [Cursor IDE](https://cursor.sh/))

### Installation

#### Download from GitHub Releases

1. Go to the [Releases page](https://github.com/imcclaskey/d3/releases) of this repository
2. Download the binary for your platform:
   - macOS/Intel: `d3-darwin-amd64`
   - macOS/Apple Silicon: `d3-darwin-arm64`
   - Linux/Intel: `d3-linux-amd64`
   - Linux/ARM: `d3-linux-arm64`
   - Windows: `d3-windows-amd64.exe`
3. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x d3-darwin-arm64  # Example for Mac with Apple Silicon
   ```
4. Move the binary to a location in your PATH (Linux/macOS):
   ```bash
   mv d3-darwin-arm64 /usr/local/bin/d3
   ```

### Basic Workflow

1. **Initialize d3**:

    ```bash
    # Initialize d3 in your project directory
    d3 init
    ```

2. **Interact via MCP Client (e.g., Cursor)**:
    - Use your AI assistant, configured with d3 tools, to interact with the server.
    - **Create a Feature**: Ask the AI to "create a new d3 feature named 'my-feature'".
        - *Alternatively, use the CLI:* `d3 feature create my-feature`
    - **Enter a Feature**: Once a feature exists, ask the AI to "enter the feature 'my-feature'".
        - *Alternatively, use the CLI:* `d3 feature enter my-feature`
    - **Move Through Phases**: Instruct the AI to "move to the define phase", "move to the design phase", or "move to the deliver phase".
        - *Alternatively, use the CLI:* `d3 phase move define`, `d3 phase move design`, or `d3 phase move deliver`
    - **Exit a Feature**: When done with a feature, or to switch, ask the AI to "exit the current feature".
        - *Alternatively, use the CLI:* `d3 exit`

3. **Develop within Phases**:
    - **Define**: Work with the AI to populate `problem.md` with requirements.
    - **Design**: Collaborate with the AI to outline the technical plan in `plan.md`.
    - **Deliver**: Generate code with the AI, guided by the design, tracking progress in `progress.yaml`.

## 📋 Development Phases

d3 enforces a structured workflow through three main phases:

### 1. Define Phase

Focus on the problem space, requirements, and user needs. Answer the "what" and "why" without diving into implementation details. Document everything in `problem.md`.

### 2. Design Phase

Translate ideas into a technical blueprint. Design architecture, component interactions, and implementation steps in `plan.md`. This becomes the roadmap for implementation.

### 3. Deliver Phase

Generate code following the technical plan, with progress tracked in `progress.yaml`. Focus exclusively on writing high-quality, maintainable code that aligns with the established plan.

## 🛠️ Commands & MCP Tools

d3 primarily interacts via its MCP server, but retains a few core CLI commands.

### CLI Commands

| Command                       | Description                                                      |
|-------------------------------|------------------------------------------------------------------|
| `d3 init [--clean | --refresh]` | Initializes or updates the d3 project. `--clean` removes existing `.d3` and re-initializes. `--refresh` updates configuration in an existing project. |
| `d3 feature create <n>`    | Create a new feature and set it as the current context           |
| `d3 feature enter <n>`     | Enter a feature context, resuming its last known phase             |
| `d3 phase move <p>`        | Move the current feature to a different phase (define, design, deliver) |
| `d3 exit`                     | Exit the current feature context, clearing active feature state. |
| `d3 feature delete <n>`    | Delete a feature and its associated content                       |
| `d3 serve`                    | Start the d3 MCP server for AI interaction                       |
| `d3 version`                  | Display the current version of d3                                |

### MCP Tool Functions (Used via AI Assistant)

| MCP Function          | Description                                                        |
|-----------------------|--------------------------------------------------------------------|
| `d3_feature_create`   | Create a new feature and set it as current context                 |
| `d3_feature_enter`    | Enter a feature context, resuming its last known phase             |
| `d3_feature_exit`     | Exit the current feature context, clearing active feature state    |
| `d3_feature_delete`   | Delete a feature and its associated content (requires confirmation)|
| `d3_phase_move`       | Move to a different phase (`define`, `design`, `deliver`)          |
| `d3_init`             | Provides CLI guidance for initializing d3                          |

## 📂 Project Structure

```text
project/
├── .d3/                  # d3 configuration and feature documentation
│   ├── features/         # Feature-specific documentation
│   │   └── my-feature/   # Individual feature folder
│   │       ├── define/        # Define Phase artifacts
│   │       │   └── problem.md   # Problem definition and requirements
│   │       ├── design/        # Design Phase artifacts
│   │       │   └── plan.md      # Technical implementation plan
│   │       ├── deliver/       # Deliver Phase artifacts
│   │       │   └── progress.yaml# Implementation progress tracking
│   │       └── .phase        # Stores the current phase for this feature (e.g., "define", "design")
│   └── .feature           # Current active feature name (if any)
├── .cursor/              # Cursor IDE configuration
│   └── rules/            # Client-side rules
│       └── d3/           # d3-specific rules
│           ├── core.gen.mdc     # Core rules for d3
│           └── phase.gen.mdc    # Phase-specific rules (generated by d3)
└── .gitignore           # Will include entries for proper d3 file handling
```

## 🔄 How It Works

1. **Initialization (`d3 init`)**:
    - **Standard `d3 init`**: If run in a new project, it creates the `.d3` directory (with `features/`), ensures proper entries in `.gitignore` and `.cursorignore`, generates base Cursor rules, and clears any active feature session. If run in an already initialized project, it will suggest using `--refresh` or `--clean`.
    - **`d3 init --clean`**: Removes the entire `.d3/` directory if it exists, then proceeds with a standard initialization.
    - **`d3 init --refresh`**: Updates an existing d3 project, ensuring all necessary directories and configurations are properly set up.
2. **Server Start**: `d3 serve` launches the MCP server, listening for client connections.
3. **Client Connection**: An AI assistant (like Cursor's) connects to the MCP server.
4. **Feature Management**:
    - **Creation**: Using MCP tools (like `d3_feature_create`), the AI directs d3 to create a feature directory (`.d3/features/<feature-name>/`), its phase subdirectories (`define/`, `design/`, `deliver/`), initial phase files (`problem.md`, `plan.md`, `progress.yaml`), and a `.phase` file (set to 'define') to track the feature's current phase.
    - **Entering**: Using `d3_feature_enter`, the AI directs d3 to set the specified feature as active. d3 reads the feature's `.phase` file to determine its current phase and updates `.d3/.feature` with the feature name.
    - **Exiting**: Using `d3_feature_exit`, the AI directs d3 to clear/remove the `.d3/.feature` file.
5. **Phase Management**: MCP tools (like `d3_phase_move`) update the content of the current feature's `.phase` file and signal the client to adjust its behavior accordingly.
6. **AI Guidance**: The AI assistant, aware of the current d3 feature and phase, provides contextually relevant assistance for the current phase's tasks and documentation.
7. **Documentation**: Work done in each phase is captured in the corresponding files within the feature's phase directories.

<p align="center">
  Built with ❤️ for better AI collaboration
</p>