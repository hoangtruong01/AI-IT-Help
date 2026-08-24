package model

import "time"

// Message represents a single chat conversation message.
type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a user request to the AI Assistant.
type ChatRequest struct {
	SessionID string    `json:"session_id,omitempty"`
	Messages  []Message `json:"messages"`
}

// ChatResponse represents AI Assistant's answer with cited knowledge sources.
type ChatResponse struct {
	Answer       string     `json:"answer"`
	Citations    []Citation `json:"citations,omitempty"`
	Confidence   float64    `json:"confidence"`
	TokensUsed   int        `json:"tokens_used"`
	FallbackMode bool       `json:"fallback_mode,omitempty"`
}

// Citation references a knowledge base article or runbook used for RAG grounding.
type Citation struct {
	ArticleID string  `json:"article_id"`
	Title     string  `json:"title"`
	Score     float64 `json:"score"`
	Category  string  `json:"category,omitempty"`
	Type      string  `json:"type,omitempty"` // "article" or "runbook"
}

// TicketAnalysis represents AI-powered categorization and suggested resolution for helpdesk tickets.
type TicketAnalysis struct {
	TicketID            string     `json:"ticket_id"`
	SuggestedCategory   string     `json:"suggested_category"`
	Priority            string     `json:"priority"` // "LOW", "MEDIUM", "HIGH", "URGENT"
	Summary             string     `json:"summary"`
	RootCause           string     `json:"root_cause"`
	SuggestedResolution string     `json:"suggested_resolution"`
	Confidence          float64    `json:"confidence"`
	Citations           []Citation `json:"citations,omitempty"`
	RequiresHumanReview bool       `json:"requires_human_review"` // Rule: AI cannot take authoritative action
	CreatedAt           time.Time  `json:"created_at"`
}

// VectorDocument represents a chunk of text stored in Qdrant vector database.
type VectorDocument struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}
