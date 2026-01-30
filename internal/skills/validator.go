package skills

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates operations against skill definitions
type Validator struct {
	allowedCommands map[string]bool
}

// NewValidator creates a new skills validator
func NewValidator() *Validator {
	return &Validator{
		allowedCommands: make(map[string]bool),
	}
}

// RegisterCommand registers a command as allowed
func (v *Validator) RegisterCommand(command string) {
	v.allowedCommands[command] = true
}

// RegisterCommands registers multiple commands as allowed
func (v *Validator) RegisterCommands(commands []string) {
	for _, cmd := range commands {
		v.allowedCommands[cmd] = true
	}
}

// IsCommandAllowed checks if a command is in the allowlist
func (v *Validator) IsCommandAllowed(command string) bool {
	return v.allowedCommands[command]
}

// ValidateCommand validates a command against the allowlist
func (v *Validator) ValidateCommand(command string) error {
	if !v.IsCommandAllowed(command) {
		return fmt.Errorf("command not allowed: %s", command)
	}
	return nil
}

// SanitizeInput sanitizes user input to prevent prompt injection
func SanitizeInput(input string) string {
	// Remove potential prompt injection patterns
	injectionPatterns := []string{
		`(?i)ignore\s+previous\s+instructions`,
		`(?i)forget\s+everything\s+above`,
		`(?i)you\s+are\s+now`,
		`(?i)system:`,
		`(?i)assistant:`,
		`(?i)human:`,
		`(?i)ai:`,
		`(?i)<\|im_start\|>`,
		`(?i)<\|im_end\|>`,
		`(?i)\[INST\]`,
		`(?i)\[/INST\]`,
	}

	sanitized := input
	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, "[FILTERED]")
	}

	return sanitized
}

// DetectPromptInjection detects potential prompt injection attempts
func DetectPromptInjection(input string) bool {
	// Patterns that indicate prompt injection
	injectionPatterns := []string{
		`(?i)ignore\s+previous\s+instructions`,
		`(?i)forget\s+everything`,
		`(?i)disregard\s+all\s+previous`,
		`(?i)you\s+are\s+now\s+\w+`,
		`(?i)act\s+as\s+if\s+you`,
		`(?i)pretend\s+to\s+be`,
		`(?i)from\s+now\s+on`,
		`(?i)system:\s*\w`,
		`(?i)assistant:\s*\w`,
		`(?i)<\|im_start\|>`,
		`(?i)\[INST\]`,
		`(?i)\\n\\nsystem`,
	}

	lowerInput := strings.ToLower(input)

	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		if re.MatchString(lowerInput) {
			return true
		}
	}

	return false
}

// GetAllowedDevOpsCommands returns the list of allowed Azure DevOps commands
func GetAllowedDevOpsCommands() []string {
	return []string{
		"devops_list_my_workitems",
		"devops_get_workitem",
		"devops_create_workitem",
		"devops_update_workitem",
		"devops_query_workitems",
		"devops_list_pipelines",
		"devops_run_pipeline",
		"devops_list_repos",
		"devops_list_boards",
	}
}

// GetAllowedTelegramCommands returns the list of allowed Telegram commands
func GetAllowedTelegramCommands() []string {
	return []string{
		"/start",
		"/help",
		"/status",
		"/workitems",
	}
}

// ValidateDevOpsWorkItemType validates work item type
func ValidateDevOpsWorkItemType(workItemType string) bool {
	allowedTypes := map[string]bool{
		"Task":       true,
		"Bug":        true,
		"User Story": true,
		"Feature":    true,
		"Epic":       true,
	}
	return allowedTypes[workItemType]
}

// ValidateDevOpsPriority validates priority value
func ValidateDevOpsPriority(priority int) bool {
	return priority >= 1 && priority <= 4
}

// ValidateDevOpsState validates work item state
func ValidateDevOpsState(state string) bool {
	allowedStates := map[string]bool{
		"New":      true,
		"Active":   true,
		"Resolved": true,
		"Closed":   true,
	}
	return allowedStates[state]
}
