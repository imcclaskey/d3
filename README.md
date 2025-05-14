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
- **Consistent Documentation**: Automatically generate and maintain technical documentation (`define.md`, `design.md`, `deliver.json`) as you build.

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher (for building from source)
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

#### Building from Source

```bash
# Clone the repository
git clone https://github.com/imcclaskey/d3.git
cd d3

# Build and install
make install
```

### Basic Workflow

1.  **Initialize d3**:
    ```bash
    # Initialize d3 in your project directory
    d3 init
    ```

2.  **Start the MCP Server**:
    ```bash
    # Start the d3 MCP server
    d3 serve
    ```
    This command keeps running, listening for instructions from your AI assistant (e.g., within Cursor).

3.  **Interact via MCP Client (e.g., Cursor)**:
    *   Use your AI assistant, configured with d3 tools, to interact with the server.
    *   **Create a Feature**: Ask the AI to "create a new d3 feature named 'my-feature'".
        *   *Alternatively, use the CLI:* `d3 feature create my-feature`
    *   **Enter a Feature**: Once a feature exists, ask the AI to "enter the feature 'my-feature'".
        *   *Alternatively, use the CLI:* `d3 feature enter my-feature`
    *   **Move Through Phases**: Instruct the AI to "move to the define phase", "move to the design phase", or "move to the deliver phase".
        *   *Alternatively, use the CLI:* `d3 phase move define`, `d3 phase move design`, or `d3 phase move deliver`
    *   **Exit a Feature**: When done with a feature, or to switch, ask the AI to "exit the current feature".
        *   *Alternatively, use the CLI:* `d3 exit`

4.  **Develop within Phases**:
    *   **Define**: Work with the AI to populate `define.md` with requirements.
    *   **Design**: Collaborate with the AI to outline the technical plan in `design.md`.
    *   **Deliver**: Generate code with the AI, guided by the design, tracking progress in `deliver.json`.

## 📋 Development Phases

d3 enforces a structured workflow through three main phases:

### 1. Define Phase

Focus on the problem space, requirements, and user needs. Answer the "what" and "why" without diving into implementation details. Document everything in `define.md`.

### 2. Design Phase

Translate ideas into a technical blueprint. Design architecture, component interactions, and implementation steps in `design.md`. This becomes the roadmap for implementation.

### 3. Deliver Phase

Generate code following the technical plan, with progress tracked in `deliver.json`. Focus exclusively on writing high-quality, maintainable code that aligns with the established plan.

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
| `d3 serve`                    | Start the d3 MCP server for AI interaction                       |
| `d3 version`                  | Display the current version of d3                                |

### MCP Tool Functions (Used via AI Assistant)

| MCP Function          | Description                                                        |
|-----------------------|--------------------------------------------------------------------|
| `d3_feature_create`   | Create a new feature and set it as current context                 |
| `d3_feature_enter`    | Enter a feature context, resuming its last known phase             |
| `d3_feature_exit`     | Exit the current feature context, clearing active feature state    |
| `d3_phase_move`       | Move to a different phase (`define`, `design`, `deliver`)          |
| `d3_get_context`      | (Implicit) Get current feature/phase context                       |

## 📂 Project Structure

```
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
│   ├── .feature      # Current active feature name (internal)
│   ├── project.md        # Project overview and business objectives
│   └── tech.md           # Technology stack documentation
└── .cursor/              # Example client configuration (e.g., Cursor)
    └── rules/            # Client-side rules
        └── d3/           # d3-specific rules
            ├── core.gen.mdc     # Core rules for d3
            └── phase.gen.mdc    # Phase-specific rules (generated by d3 potentially)
```
*(Note: The exact structure of `deliver.json` and client rules might vary)*

## 🔄 How It Works

1.  **Initialization (`d3 init`)**:
    *   **Standard `d3 init`**: If run in a new project, it creates the `.d3` directory (with `features/`, `project.md`, `tech.md`), ensures `mcp.json` has the d3 server entry (preserving other content if `mcp.json` exists), sets up d3-specific `.gitignore` files, generates base Cursor rules, and clears any active feature session. If run in an already initialized project, it will inform the user that the project is already initialized and suggest using `--refresh` or `--clean`.
    *   **`d3 init --clean`**: Removes the entire `.d3/` directory if it exists (note: `mcp.json` in the project root is *not* removed). Then, it proceeds with a standard initialization as described above, including clearing any active feature session. This is useful for a fresh start of d3 structures within the project.
    *   **`d3 init --refresh`**: Updates an existing d3 project. It ensures essential directories (`.d3/`, `.d3/features/`), `mcp.json` entries, and `.gitignore` files are correctly set up or updated. It also refreshes Cursor rules. If the project was not previously initialized, `--refresh` behaves like a standard `d3 init`. This command does not clear the active feature session if the project was already initialized and is being refreshed.
2.  **Server Start**: `d3 serve` launches the MCP server, listening for client connections.
3.  **Client Connection**: An AI assistant (like Cursor's) connects to the MCP server.
4.  **Feature Management**:
    *   **Creation**: Using MCP tools (like `d3_feature_create`), the AI directs d3 to create a feature directory (`.d3/features/<feature-name>/`), its phase subdirectories (`define/`, `design/`, `deliver/`), initial phase files (`problem.md`, `plan.md`, `progress.yaml`), and a `.phase` file (e.g., containing 'define') to track the feature's current phase.
    *   **Entering**: Using `d3_feature_enter`, the AI directs d3 to set the specified feature as active. d3 reads the feature's `.phase` file to determine its current phase and updates `.d3/.feature` (by writing the feature name into it) to mark this feature as active.
    *   **Exiting**: Using `d3_feature_exit`, the AI directs d3 to clear/delete the `.d3/.feature` file.
5.  **Phase Management**: MCP tools (like `d3_phase_move`) update the content of the current feature's `.d3/features/<feature-name>/.phase` file and signal the client to adjust its behavior (e.g., load different rules or focus on specific phase files). The in-memory context within d3 is also updated.
6.  **AI Guidance**: The AI assistant, aware of the current d3 feature and phase (which d3 determines by reading the `.d3/.feature` file for the active feature name, and then reading the `.d3/features/<active_feature_name>/.phase` file for its current phase), provides contextually relevant assistance for populating `problem.md`, `plan.md`, or generating code tracked in `progress.yaml`.
7.  **Documentation**: Work done in each phase is captured in the corresponding files within the feature's phase directories.

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues or pull requests.

1. Fork the repository
2. Create your feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add some amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

<p align="center">
  Built with ❤️ for better AI collaboration
</p>