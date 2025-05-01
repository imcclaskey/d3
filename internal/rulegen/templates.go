package rulegen

// Templates contains the embedded rule templates
// This allows the application to run without external template files
var Templates = map[string]string{
	"setup":         setupTemplate,
	"ideation":      ideationTemplate,
	"instruction":   instructionTemplate,
	"implementation": implementationTemplate,
}

// GetTemplate returns the template content for the given phase
func GetTemplate(phase string) (string, bool) {
	content, exists := Templates[phase]
	return content, exists
} 