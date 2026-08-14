package prompt

import "fmt"

// SystemPrompt defines the foundational persona and constraints for the enterprise AI assistant.
const SystemPrompt = `You are the EOMP AI Operations Assistant.
Your purpose is to help employees, IT staff, and managers resolve operational inquiries efficiently.

OPERATIONAL RULES:
1. Always base your answers on verified organizational documentation when available.
2. AI is an advisory and assistant tool: never claim to have performed irreversible actions (such as deleting data, approving financial requests, or revoking credentials) directly.
3. Be professional, clear, concise, and structured.`

// HelpdeskTriagePrompt formats a ticket for categorization and resolution suggestion.
func HelpdeskTriagePrompt(title, description string) string {
	return fmt.Sprintf(`Please analyze the following IT helpdesk ticket:
Title: %s
Description: %s

Provide:
1. Category (Hardware, Software, Network, Access, HR, Other)
2. Priority (Low, Medium, High, Critical)
3. One-sentence summary
4. Suggested troubleshooting steps for the technician`, title, description)
}
