# Problem Statement
Users cannot easily customize d3 workflow templates (e.g., for define, design, deliver phases) because these templates are embedded within the compiled binary. This prevents users from persisting their template modifications across d3 project initializations or refreshes.

# Feature Goals
- Enable users to define and manage their own custom d3 workflow templates.
- Ensure that user-defined custom templates are prioritized and used by the d3 tool when present.
- Allow custom templates to be persisted and automatically applied during phase transitions or project (re)initialization.
- Retain the default, in-binary templates as a fallback when no user-defined custom templates are available.

# Core Requirements
- Users MUST be able to store their custom workflow templates in a dedicated directory within their project's `.d3` folder (e.g., `.d3/rules/`).
- The d3 tool MUST check for the existence of user-defined templates in the specified directory before using the default in-binary templates.
- If custom templates are found in the designated user directory, the d3 tool MUST copy these templates into the `.cursor/d3/rules` directory for active use, overwriting any existing files derived from the default templates.
- If no custom templates are found in the user-defined directory, the d3 tool MUST use the default in-binary templates and copy them to `.cursor/d3/rules` as per the current behavior.
- Changes made by users to their custom templates in the `.d3/rules/` directory MUST be reflected the next time a relevant d3 command is run that would involve template copying (e.g., phase change, init).
- The d3 `init` command MUST support an optional flag (e.g., `--custom-rules=true` or `--with-custom-rules`) that, when supplied, triggers the creation of the `.d3/rules/` directory and populates it with copies of the default in-binary templates as a starting point for user customization.

# Scope Exclusions
- This feature does not include a UI or interactive management tool for custom templates within the d3 CLI itself. Template management is file-system based.
- This feature will not support versioning or branching of custom rule sets beyond what is offered by the user's own version control system (e.g., git) on their `.d3` directory.
- Migration of existing, modified `.cursor/d3/rules` files to the new custom rules directory is not an automatic process.
