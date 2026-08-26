package prompt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eomp/services/ai/internal/model"
)

// SystemPrompt defines the foundational persona and constraints for the enterprise AI assistant.
const SystemPrompt = `You are the EOMP AI Operations Copilot, an enterprise IT service management (ITSM) and operations assistant.
Your purpose is to help employees, IT support staff, and DevOps engineers diagnose issues, troubleshoot infrastructure, and apply standard operating procedures (SOP).

OPERATIONAL RULES:
1. Always base your answers on verified organizational documentation and SOP runbooks when available.
2. AI is an advisory tool: never claim to have performed irreversible actions (such as deleting data, approving requests, or revoking credentials) directly without human authorization.
3. Be professional, clear, concise, structured (use markdown headers, bold text, step-by-step lists).
4. When relevant documentation is cited, reference the runbook code or article title.`

// FormatRAGPrompt injects retrieved citations/documents into the system or user prompt.
func FormatRAGPrompt(userQuery string, citations []model.Citation) string {
	if len(citations) == 0 {
		return userQuery
	}

	var sb strings.Builder
	sb.WriteString("Retrieved Knowledge Base & Runbook Context:\n")
	for i, c := range citations {
		sb.WriteString(fmt.Sprintf("[%d] %s (Type: %s, Category: %s, Relevance: %.2f)\n", i+1, c.Title, c.Type, c.Category, c.Score))
	}
	sb.WriteString("\nUser Query: ")
	sb.WriteString(userQuery)
	sb.WriteString("\n\nPlease provide a clear resolution following standard IT procedures.")
	return sb.String()
}

// HelpdeskTriagePrompt formats a ticket for categorization and resolution suggestion.
func HelpdeskTriagePrompt(title, description string) string {
	return fmt.Sprintf(`You are an expert ITIL Service Desk Lead. Analyze the following IT helpdesk ticket and output ONLY valid JSON.

Ticket Title: %s
Ticket Description: %s

Output must match this exact JSON schema:
{
  "suggested_category": "Network & Access | IT Security & Access | Hardware & Equipment | DevOps & Infrastructure | Software & Productivity | IT Support",
  "priority": "LOW | MEDIUM | HIGH | URGENT",
  "summary": "One-sentence summary of the core incident",
  "root_cause": "Probable technical root cause",
  "suggested_resolution": "Step-by-step actionable troubleshooting instructions",
  "confidence": 0.95
}

DO NOT include any markdown code fences or other text outside the JSON object.`, title, description)
}

// ParseTriageJSON parses the LLM output into a model.TicketAnalysis.
func ParseTriageJSON(raw string, title string) (*model.TicketAnalysis, error) {
	// Clean markdown fences if any
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var parsed struct {
		SuggestedCategory   string  `json:"suggested_category"`
		Priority            string  `json:"priority"`
		Summary             string  `json:"summary"`
		RootCause           string  `json:"root_cause"`
		SuggestedResolution string  `json:"suggested_resolution"`
		Confidence          float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse triage JSON: %w (raw output: %s)", err, raw)
	}

	if parsed.Confidence <= 0 || parsed.Confidence > 1.0 {
		parsed.Confidence = 0.92
	}
	if parsed.Priority == "" {
		parsed.Priority = "MEDIUM"
	}
	if parsed.SuggestedCategory == "" {
		parsed.SuggestedCategory = "IT Support"
	}

	return &model.TicketAnalysis{
		TicketID:            fmt.Sprintf("AI-TK-%d", time.Now().Unix()%10000),
		SuggestedCategory:   parsed.SuggestedCategory,
		Priority:            parsed.Priority,
		Summary:             parsed.Summary,
		RootCause:           parsed.RootCause,
		SuggestedResolution: parsed.SuggestedResolution,
		Confidence:          parsed.Confidence,
		RequiresHumanReview: true,
		CreatedAt:           time.Now(),
	}, nil
}
