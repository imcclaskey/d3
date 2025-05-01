---
description: Phase-specific rules for i3 framework (Setup Phase)
globs: 
alwaysApply: true
---

# i3 Phase: Setup

## 1. Purpose
You are assisting with the initial setup phase of the i3 framework. This foundational phase establishes the project context and prepares for future feature development.

In this phase, you should:

1. Help the user define their project in project.md
2. Help the user define their tech stack in tech.md
3. Guide the user on how to create features using the i3 create command

## 2. Required Output Format

The primary artifacts of this phase are:

1. **[project.md](mdc:.i3/project.md)**
   * Project name and description
   * Business objectives
   * Target users/audience
   * Key features and functionality overview
   * Any critical constraints or requirements
   * Future growth areas and evolution plans

2. **[tech.md](mdc:.i3/tech.md)**
   * Programming languages
   * Frameworks and libraries
   * Database technologies
   * Infrastructure and deployment details
   * Development tools and workflows
   * Architecture patterns and practices

## 3. Forbidden Actions

During the Setup phase, you **MUST NOT**:

* Start implementing specific features without proper context creation
* Make technical decisions that would be better decided during feature phases
* Skip the creation of required documentation files
* Work on detailed implementation before the project scope is defined
* Suggest changing established technology choices without explicit user direction
* Ignore existing code patterns and conventions

## 4. Operational Context & Workflow

**A. Guiding Principles:**

* **Collaborative Onboarding:** Work with the user to understand their project needs and guide them through the setup process
* **Focus on Clarity:** Ensure project and technology documentation is clear, concise, and comprehensive
* **Preparation for Features:** Set the stage for successful feature development by establishing solid foundational context
* **Respect Existing Work:** Recognize and document established patterns in existing codebases
* **Balance Current vs. Future:** Document both what exists today and what needs to evolve

**B. Working with Setup Files:**

* **Check Existence:** Verify if project.md and tech.md already exist
* **Template Creation:** If files don't exist, help create them with appropriate sections
* **Iterative Refinement:** Work with the user to refine content until it accurately represents the project
* **Command Guidance:** Provide clear instructions on using i3 commands, especially for feature creation

**C. Handling Existing Codebases:**

* **Codebase Analysis:** Prioritize analyzing the existing codebase to identify current technologies and patterns
* **Tech Documentation Approach:** For tech.md, favor documenting what's observed in the codebase as technology choices are likely established
* **Project Evolution Focus:** For project.md, balance documenting current functionality with the growth and changes that motivated i3 adoption
* **Context Gathering:** Proactively search the codebase for:
  * Package/dependency information files (package.json, go.mod, requirements.txt, etc.)
  * README files, documentation, and comments
  * Main entry points and core architecture files
  * Common patterns in directory structure and file organization
* **Tech Stack Inference:** Draw conclusions about the tech stack from actual code rather than making assumptions

**D. Completion Indicators:**

* Both project.md and tech.md are complete with substantial content
* The user understands how to create and work with features
* The project context is well-defined and ready for feature development
* Existing code patterns and standards are acknowledged and documented
* Growth areas that motivated i3 adoption are clearly identified

## 5. Required Files

* **[project.md](mdc:.i3/project.md)** - Project description, including both current state and evolution plans
* **[tech.md](mdc:.i3/tech.md)** - Technical stack details, primarily derived from existing codebase analysis 