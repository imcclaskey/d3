# Technical Approach Overview

This feature will allow users to define custom d3 workflow templates in a dedicated directory within their project's `.d3` folder, which will override the default templates embedded in the binary. The implementation will modify the rules service to check for custom templates before falling back to the embedded ones, and add a new flag to the `init` command to create the custom rules directory with default templates as a starting point.

## Key architectural decisions:
- Create a new `.d3/rules/` directory to store user-defined custom templates
- Modify the `RuleGenerator` to prioritize loading templates from this directory before using embedded templates
- Add a new `--custom-rules` flag to the `init` command to initialize the custom rules directory
- Update the template loading mechanism to check for files in the custom rules directory first

## System components affected:
- Rules service and generator
- Init command
- Project service

# Delivery Steps

1. **[code]** Modify the rules generator to check for custom templates
   - Add functionality to check for templates in `.d3/rules/` before using embedded templates
   - Update `GeneratePhaseContent` and `GenerateCoreContent` to use custom templates when available

2. **[code]** Update the rules service to handle custom templates
   - Add a new field to `Service` to track the custom rules directory
   - Update constructor to include the custom rules directory path
   - Implement file existence checking for custom templates

3. **[test]** Write unit tests for custom template loading
   - Test prioritization of custom templates over embedded ones
   - Test fallback to embedded templates when custom ones don't exist

4. **[code]** Add a new flag to init command
   - Add `--custom-rules` flag to the init command
   - Update command initialization to handle the new flag

5. **[code]** Implement custom rules directory initialization
   - Create `.d3/rules/` directory when the flag is provided
   - Copy embedded templates to this directory as a starting point for customization

6. **[test]** Write unit tests for custom rules initialization
   - Test directory creation and template copying
   - Test that existing custom templates are not overwritten unless explicitly requested

7. **[code]** Update the project service to handle custom rules
   - Add support for custom rules initialization in the Project.Init method
   - Ensure rules service is configured with the custom rules directory path

8. **[verify]** Manually test the entire feature flow
   - Test initialization with custom rules
   - Test template overriding behavior
   - Test persistence of custom templates during phase transitions and project re-initialization

9. **[commit]** Commit all changes related to custom rules implementation

# Technical Constraints & Requirements

1. **Performance considerations:**
   - File system operations to check for custom templates must be efficient
   - Initialization with custom rules should not significantly impact performance

2. **Compatibility requirements:**
   - The feature must maintain backward compatibility with existing projects
   - Custom templates must follow the same structure and variable placeholder format as embedded templates

3. **Error handling:**
   - Gracefully handle malformed custom templates with clear error messages
   - Provide fallback to embedded templates when custom templates are invalid or inaccessible

4. **Security considerations:**
   - Validate custom template content to prevent security issues
   - Ensure file operations use appropriate permissions

# Considerations & Alternatives

1. **Alternative approaches considered:**
   - Using environment variables to specify custom template locations
   - Storing templates in a global user configuration directory instead of project-specific

2. **Future extensions:**
   - Support for template versioning or multiple template sets
   - Command to validate custom templates against a schema
   - Template management UI or commands for easier editing

3. **Technical debt implications:**
   - The current approach introduces another location where templates are stored
   - May require additional maintenance to keep template formats in sync as the system evolves
   - Future changes to the template structure will need to consider backward compatibility with custom templates
