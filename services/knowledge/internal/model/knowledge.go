package model

import (
	"time"
)

// KnowledgeCategory represents an organizational category for articles and runbooks.
type KnowledgeCategory struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Icon        string    `json:"icon"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateCategoryRequest DTO for creating a new category.
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Icon        string  `json:"icon"`
	Description *string `json:"description,omitempty"`
}

// KnowledgeArticle represents a standard operating procedure or documentation article.
type KnowledgeArticle struct {
	ID           string    `json:"id"`
	CategoryID   string    `json:"category_id"`
	CategoryName *string   `json:"category_name,omitempty"`
	CategoryCode *string   `json:"category_code,omitempty"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Summary      string    `json:"summary"`
	Content      string    `json:"content"`
	Tags         []string  `json:"tags"`
	AuthorID     string    `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	ViewCount    int       `json:"view_count"`
	HelpfulCount int       `json:"helpful_count"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateArticleRequest DTO for authoring a new article.
type CreateArticleRequest struct {
	CategoryID  string   `json:"category_id"`
	Title       string   `json:"title"`
	Slug        *string  `json:"slug,omitempty"`
	Summary     string   `json:"summary"`
	Content     string   `json:"content"`
	Tags        []string `json:"tags,omitempty"`
	AuthorID    *string  `json:"author_id,omitempty"`
	AuthorName  *string  `json:"author_name,omitempty"`
	IsPublished *bool    `json:"is_published,omitempty"`
}

// UpdateArticleRequest DTO for updating an existing article.
type UpdateArticleRequest struct {
	CategoryID  *string   `json:"category_id,omitempty"`
	Title       *string   `json:"title,omitempty"`
	Summary     *string   `json:"summary,omitempty"`
	Content     *string   `json:"content,omitempty"`
	Tags        *[]string `json:"tags,omitempty"`
	IsPublished *bool     `json:"is_published,omitempty"`
}

// RunbookStep represents an individual step in a Standard Operating Procedure (SOP).
type RunbookStep struct {
	Step     int    `json:"step"`
	Action   string `json:"action"`
	Command  string `json:"command,omitempty"`
	Expected string `json:"expected,omitempty"`
}

// KnowledgeRunbook represents an actionable, step-by-step IT operational runbook.
type KnowledgeRunbook struct {
	ID            string        `json:"id"`
	Code          string        `json:"code"`
	Title         string        `json:"title"`
	Category      string        `json:"category"`
	Description   string        `json:"description"`
	Prerequisites string        `json:"prerequisites"`
	StepsJSON     []RunbookStep `json:"steps_json"`
	RollbackSteps string        `json:"rollback_steps"`
	AuthorName    string        `json:"author_name"`
	IsActive      bool          `json:"is_active"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// CreateRunbookRequest DTO for creating an SOP Runbook.
type CreateRunbookRequest struct {
	Code          string        `json:"code"`
	Title         string        `json:"title"`
	Category      string        `json:"category"`
	Description   string        `json:"description"`
	Prerequisites string        `json:"prerequisites"`
	StepsJSON     []RunbookStep `json:"steps_json"`
	RollbackSteps string        `json:"rollback_steps"`
	AuthorName    *string       `json:"author_name,omitempty"`
}

// DocumentEmbedding represents a vector chunk reference for semantic search.
type DocumentEmbedding struct {
	ID          string    `json:"id"`
	ArticleID   string    `json:"article_id"`
	ChunkIndex  int       `json:"chunk_index"`
	ChunkText   string    `json:"chunk_text"`
	EmbeddingID string    `json:"embedding_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// KnowledgeSearchResult represents a ranked search result item.
type KnowledgeSearchResult struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // "article" or "runbook"
	Title       string   `json:"title"`
	Snippet     string   `json:"snippet"`
	Category    string   `json:"category"`
	Score       float64  `json:"score"`
	Tags        []string `json:"tags,omitempty"`
	SlugOrCode  string   `json:"slug_or_code"`
	ViewCount   int      `json:"view_count,omitempty"`
	UpdatedTime string   `json:"updated_time"`
}

// ArticleListQuery query parameters for listing articles.
type ArticleListQuery struct {
	Category string
	Search   string
	Page     int
	PageSize int
}

// ArticleListResponse paginated list of articles.
type ArticleListResponse struct {
	Data       []KnowledgeArticle `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// RunbookListResponse list of runbooks.
type RunbookListResponse struct {
	Data       []KnowledgeRunbook `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// KnowledgeStats dashboard KPIs.
type KnowledgeStats struct {
	TotalArticles   int `json:"total_articles"`
	TotalCategories int `json:"total_categories"`
	TotalRunbooks   int `json:"total_runbooks"`
	TotalViews      int `json:"total_views"`
}
