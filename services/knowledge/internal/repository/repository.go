package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/lib/pq"

	"eomp/services/knowledge/internal/model"
)

// ErrVersionConflict indicates that an article changed after the caller read it.
var ErrVersionConflict = errors.New("knowledge article version conflict")

// Repository defines data access methods for Knowledge Service.
type Repository interface {
	// Categories
	ListCategories(ctx context.Context) ([]model.KnowledgeCategory, error)
	CreateCategory(ctx context.Context, cat *model.KnowledgeCategory) error

	// Articles
	ListArticles(ctx context.Context, query model.ArticleListQuery) (*model.ArticleListResponse, error)
	FindArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	FindArticleByIDForVisibility(ctx context.Context, id string, publishedOnly bool) (*model.KnowledgeArticle, error)
	FindArticleBySlug(ctx context.Context, slug string) (*model.KnowledgeArticle, error)
	FindArticleBySlugForVisibility(ctx context.Context, slug string, publishedOnly bool) (*model.KnowledgeArticle, error)
	CreateArticle(ctx context.Context, art *model.KnowledgeArticle) error
	UpdateArticle(ctx context.Context, art *model.KnowledgeArticle) error
	DeleteArticle(ctx context.Context, id string, expectedVersion int) error
	IncrementArticleViews(ctx context.Context, id string) error

	// Runbooks (SOP)
	ListRunbooks(ctx context.Context, category, search string, page, pageSize int) (*model.RunbookListResponse, error)
	FindRunbookByID(ctx context.Context, id string) (*model.KnowledgeRunbook, error)
	FindRunbookByCode(ctx context.Context, code string) (*model.KnowledgeRunbook, error)
	CreateRunbook(ctx context.Context, rb *model.KnowledgeRunbook) error

	// Search & Aggregation
	Search(ctx context.Context, keyword, category string, limit int, publishedOnly bool) ([]model.KnowledgeSearchResult, error)
	GetStats(ctx context.Context) (*model.KnowledgeStats, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository constructs a new PostgreSQL Knowledge repository.
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

// ----------------------------------------------------------------------
// Categories
// ----------------------------------------------------------------------

func (r *postgresRepository) ListCategories(ctx context.Context) ([]model.KnowledgeCategory, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	query := `
		SELECT id, name, code, icon, description, created_at, updated_at
		FROM knowledge_categories
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []model.KnowledgeCategory
	for rows.Next() {
		var cat model.KnowledgeCategory
		var desc sql.NullString
		if err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Code,
			&cat.Icon,
			&desc,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if desc.Valid {
			cat.Description = &desc.String
		}
		categories = append(categories, cat)
	}

	if categories == nil {
		categories = []model.KnowledgeCategory{}
	}
	return categories, nil
}

func (r *postgresRepository) CreateCategory(ctx context.Context, cat *model.KnowledgeCategory) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}

	query := `
		INSERT INTO knowledge_categories (id, name, code, icon, description, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		cat.Name,
		cat.Code,
		cat.Icon,
		cat.Description,
	).Scan(&cat.ID, &cat.CreatedAt, &cat.UpdatedAt)
}

// ----------------------------------------------------------------------
// Articles
// ----------------------------------------------------------------------

func (r *postgresRepository) ListArticles(ctx context.Context, query model.ArticleListQuery) (*model.ArticleListResponse, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	if query.PublishedOnly {
		whereClauses = append(whereClauses, "a.is_published = true")
	}
	args := []any{}
	argIndex := 1

	if query.Category != "" && query.Category != "All" && query.Category != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("(c.code = $%d OR c.name ILIKE $%d OR a.category_id::text = $%d)", argIndex, argIndex, argIndex))
		args = append(args, query.Category)
		argIndex++
	}

	if query.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(a.title ILIKE $%d OR a.summary ILIKE $%d OR a.content ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+query.Search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// 1. Total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM knowledge_articles a
		LEFT JOIN knowledge_categories c ON a.category_id = c.id
		WHERE %s
	`, whereSQL)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count articles: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			a.id, a.category_id, c.name, c.code, a.title, a.slug, a.summary, a.content,
			a.tags, a.author_id, a.author_name, a.department_id, a.view_count, a.helpful_count, a.is_published, a.version,
			a.created_at, a.updated_at
		FROM knowledge_articles a
		LEFT JOIN knowledge_categories c ON a.category_id = c.id
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []model.KnowledgeArticle
	for rows.Next() {
		var art model.KnowledgeArticle
		var catName, catCode sql.NullString
		var tags pq.StringArray

		if err := rows.Scan(
			&art.ID,
			&art.CategoryID,
			&catName,
			&catCode,
			&art.Title,
			&art.Slug,
			&art.Summary,
			&art.Content,
			&tags,
			&art.AuthorID,
			&art.AuthorName,
			&art.DepartmentID,
			&art.ViewCount,
			&art.HelpfulCount,
			&art.IsPublished,
			&art.Version,
			&art.CreatedAt,
			&art.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}

		if catName.Valid {
			art.CategoryName = &catName.String
		}
		if catCode.Valid {
			art.CategoryCode = &catCode.String
		}
		art.Tags = []string(tags)
		if art.Tags == nil {
			art.Tags = []string{}
		}

		articles = append(articles, art)
	}

	if articles == nil {
		articles = []model.KnowledgeArticle{}
	}

	return &model.ArticleListResponse{
		Data:       articles,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	return r.FindArticleByIDForVisibility(ctx, id, false)
}

func (r *postgresRepository) FindArticleByIDForVisibility(ctx context.Context, id string, publishedOnly bool) (*model.KnowledgeArticle, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	query := `
		SELECT 
			a.id, a.category_id, c.name, c.code, a.title, a.slug, a.summary, a.content,
			a.tags, a.author_id, a.author_name, a.department_id, a.view_count, a.helpful_count, a.is_published, a.version,
			a.created_at, a.updated_at
		FROM knowledge_articles a
		LEFT JOIN knowledge_categories c ON a.category_id = c.id
		WHERE a.id = $1 AND ($2 = false OR a.is_published = true)
	`
	var art model.KnowledgeArticle
	var catName, catCode sql.NullString
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, id, publishedOnly).Scan(
		&art.ID,
		&art.CategoryID,
		&catName,
		&catCode,
		&art.Title,
		&art.Slug,
		&art.Summary,
		&art.Content,
		&tags,
		&art.AuthorID,
		&art.AuthorName,
		&art.DepartmentID,
		&art.ViewCount,
		&art.HelpfulCount,
		&art.IsPublished,
		&art.Version,
		&art.CreatedAt,
		&art.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find article by id: %w", err)
	}

	if catName.Valid {
		art.CategoryName = &catName.String
	}
	if catCode.Valid {
		art.CategoryCode = &catCode.String
	}
	art.Tags = []string(tags)
	if art.Tags == nil {
		art.Tags = []string{}
	}

	return &art, nil
}

func (r *postgresRepository) FindArticleBySlug(ctx context.Context, slug string) (*model.KnowledgeArticle, error) {
	return r.FindArticleBySlugForVisibility(ctx, slug, false)
}

func (r *postgresRepository) FindArticleBySlugForVisibility(ctx context.Context, slug string, publishedOnly bool) (*model.KnowledgeArticle, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	query := `
		SELECT 
			a.id, a.category_id, c.name, c.code, a.title, a.slug, a.summary, a.content,
			a.tags, a.author_id, a.author_name, a.department_id, a.view_count, a.helpful_count, a.is_published, a.version,
			a.created_at, a.updated_at
		FROM knowledge_articles a
		LEFT JOIN knowledge_categories c ON a.category_id = c.id
		WHERE a.slug = $1 AND ($2 = false OR a.is_published = true)
	`
	var art model.KnowledgeArticle
	var catName, catCode sql.NullString
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, slug, publishedOnly).Scan(
		&art.ID,
		&art.CategoryID,
		&catName,
		&catCode,
		&art.Title,
		&art.Slug,
		&art.Summary,
		&art.Content,
		&tags,
		&art.AuthorID,
		&art.AuthorName,
		&art.DepartmentID,
		&art.ViewCount,
		&art.HelpfulCount,
		&art.IsPublished,
		&art.Version,
		&art.CreatedAt,
		&art.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find article by slug: %w", err)
	}

	if catName.Valid {
		art.CategoryName = &catName.String
	}
	if catCode.Valid {
		art.CategoryCode = &catCode.String
	}
	art.Tags = []string(tags)
	if art.Tags == nil {
		art.Tags = []string{}
	}

	return &art, nil
}

func (r *postgresRepository) CreateArticle(ctx context.Context, art *model.KnowledgeArticle) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}

	query := `
		INSERT INTO knowledge_articles (
			id, category_id, title, slug, summary, content, tags,
			author_id, author_name, department_id, view_count, helpful_count, is_published,
			created_at, updated_at
		)
		VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6,
			$7, $8, NULLIF($9, ''), 0, 0, $10,
			NOW(), NOW()
		)
		RETURNING id, version, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		art.CategoryID,
		art.Title,
		art.Slug,
		art.Summary,
		art.Content,
		pq.Array(art.Tags),
		art.AuthorID,
		art.AuthorName,
		art.DepartmentID,
		art.IsPublished,
	).Scan(&art.ID, &art.Version, &art.CreatedAt, &art.UpdatedAt)
}

func (r *postgresRepository) UpdateArticle(ctx context.Context, art *model.KnowledgeArticle) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}

	query := `
		UPDATE knowledge_articles
		SET category_id = $2, title = $3, summary = $4, content = $5, tags = $6,
		    is_published = $7, version = version + 1, updated_at = NOW()
		WHERE id = $1 AND version = $8
		RETURNING version, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		art.ID,
		art.CategoryID,
		art.Title,
		art.Summary,
		art.Content,
		pq.Array(art.Tags),
		art.IsPublished,
		art.Version,
	).Scan(&art.Version, &art.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrVersionConflict
	}
	return err
}

func (r *postgresRepository) DeleteArticle(ctx context.Context, id string, expectedVersion int) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}
	result, err := r.db.ExecContext(ctx, "DELETE FROM knowledge_articles WHERE id = $1 AND version = $2", id, expectedVersion)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect article delete result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (r *postgresRepository) IncrementArticleViews(ctx context.Context, id string) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}
	_, err := r.db.ExecContext(ctx, "UPDATE knowledge_articles SET view_count = view_count + 1 WHERE id = $1", id)
	return err
}

// ----------------------------------------------------------------------
// Runbooks
// ----------------------------------------------------------------------

func (r *postgresRepository) ListRunbooks(ctx context.Context, category, search string, page, pageSize int) (*model.RunbookListResponse, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	whereClauses := []string{"is_active = true"}
	args := []any{}
	argIndex := 1

	if category != "" && category != "All" && category != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("category ILIKE $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+search+"%")
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM runbooks WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count runbooks: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, code, title, category, description, prerequisites, steps_json, rollback_steps, author_name, is_active, created_at, updated_at
		FROM runbooks
		WHERE %s
		ORDER BY code ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query runbooks: %w", err)
	}
	defer rows.Close()

	var runbooks []model.KnowledgeRunbook
	for rows.Next() {
		var rb model.KnowledgeRunbook
		var stepsRaw []byte

		if err := rows.Scan(
			&rb.ID,
			&rb.Code,
			&rb.Title,
			&rb.Category,
			&rb.Description,
			&rb.Prerequisites,
			&stepsRaw,
			&rb.RollbackSteps,
			&rb.AuthorName,
			&rb.IsActive,
			&rb.CreatedAt,
			&rb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan runbook: %w", err)
		}

		if len(stepsRaw) > 0 {
			_ = json.Unmarshal(stepsRaw, &rb.StepsJSON)
		}
		if rb.StepsJSON == nil {
			rb.StepsJSON = []model.RunbookStep{}
		}

		runbooks = append(runbooks, rb)
	}

	if runbooks == nil {
		runbooks = []model.KnowledgeRunbook{}
	}

	return &model.RunbookListResponse{
		Data:       runbooks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindRunbookByID(ctx context.Context, id string) (*model.KnowledgeRunbook, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	query := `
		SELECT id, code, title, category, description, prerequisites, steps_json, rollback_steps, author_name, is_active, created_at, updated_at
		FROM runbooks
		WHERE id = $1
	`
	var rb model.KnowledgeRunbook
	var stepsRaw []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rb.ID,
		&rb.Code,
		&rb.Title,
		&rb.Category,
		&rb.Description,
		&rb.Prerequisites,
		&stepsRaw,
		&rb.RollbackSteps,
		&rb.AuthorName,
		&rb.IsActive,
		&rb.CreatedAt,
		&rb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find runbook by id: %w", err)
	}

	if len(stepsRaw) > 0 {
		_ = json.Unmarshal(stepsRaw, &rb.StepsJSON)
	}
	if rb.StepsJSON == nil {
		rb.StepsJSON = []model.RunbookStep{}
	}

	return &rb, nil
}

func (r *postgresRepository) FindRunbookByCode(ctx context.Context, code string) (*model.KnowledgeRunbook, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	query := `
		SELECT id, code, title, category, description, prerequisites, steps_json, rollback_steps, author_name, is_active, created_at, updated_at
		FROM runbooks
		WHERE code = $1
	`
	var rb model.KnowledgeRunbook
	var stepsRaw []byte

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&rb.ID,
		&rb.Code,
		&rb.Title,
		&rb.Category,
		&rb.Description,
		&rb.Prerequisites,
		&stepsRaw,
		&rb.RollbackSteps,
		&rb.AuthorName,
		&rb.IsActive,
		&rb.CreatedAt,
		&rb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find runbook by code: %w", err)
	}

	if len(stepsRaw) > 0 {
		_ = json.Unmarshal(stepsRaw, &rb.StepsJSON)
	}
	if rb.StepsJSON == nil {
		rb.StepsJSON = []model.RunbookStep{}
	}

	return &rb, nil
}

func (r *postgresRepository) CreateRunbook(ctx context.Context, rb *model.KnowledgeRunbook) error {
	if r.db == nil {
		return errors.New("database connection not available")
	}

	stepsBytes, err := json.Marshal(rb.StepsJSON)
	if err != nil {
		stepsBytes = []byte("[]")
	}

	query := `
		INSERT INTO runbooks (
			id, code, title, category, description, prerequisites, steps_json,
			rollback_steps, author_name, is_active, created_at, updated_at
		)
		VALUES (
			uuid_generate_v4(), $1, $2, $3, $4, $5, $6,
			$7, $8, $9, NOW(), NOW()
		)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		rb.Code,
		rb.Title,
		rb.Category,
		rb.Description,
		rb.Prerequisites,
		stepsBytes,
		rb.RollbackSteps,
		rb.AuthorName,
		rb.IsActive,
	).Scan(&rb.ID, &rb.CreatedAt, &rb.UpdatedAt)
}

// ----------------------------------------------------------------------
// Search & Aggregation
// ----------------------------------------------------------------------

func (r *postgresRepository) Search(ctx context.Context, keyword, category string, limit int, publishedOnly bool) ([]model.KnowledgeSearchResult, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	if limit <= 0 {
		limit = 10
	}

	var results []model.KnowledgeSearchResult
	pattern := "%" + keyword + "%"

	// 1. Search Articles
	artQuery := `
		SELECT 
			a.id, 'article' as type, a.title, a.summary, COALESCE(c.name, 'General') as category,
			a.slug, a.view_count, a.tags,
			CASE 
				WHEN a.title ILIKE $1 THEN 0.98
				WHEN a.summary ILIKE $1 THEN 0.85
				ELSE 0.72
			END as score,
			a.updated_at
		FROM knowledge_articles a
		LEFT JOIN knowledge_categories c ON a.category_id = c.id
		WHERE ($3 = false OR a.is_published = true)
		  AND (a.title ILIKE $1 OR a.summary ILIKE $1 OR a.content ILIKE $1)
		ORDER BY score DESC, a.view_count DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, artQuery, pattern, limit, publishedOnly)
	if err != nil {
		return nil, fmt.Errorf("search knowledge articles: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var res model.KnowledgeSearchResult
		var tags pq.StringArray
		var updatedAt time.Time

		if err := rows.Scan(
			&res.ID,
			&res.Type,
			&res.Title,
			&res.Snippet,
			&res.Category,
			&res.SlugOrCode,
			&res.ViewCount,
			&tags,
			&res.Score,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan knowledge article search result: %w", err)
		}
		res.Tags = []string(tags)
		res.UpdatedTime = updatedAt.Format(time.RFC3339)
		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate knowledge article search results: %w", err)
	}

	// 2. Search Runbooks
	rbQuery := `
		SELECT 
			id, 'runbook' as type, title, description, category,
			code, 0 as view_count,
			CASE 
				WHEN title ILIKE $1 THEN 0.95
				WHEN description ILIKE $1 THEN 0.82
				ELSE 0.70
			END as score,
			updated_at
		FROM runbooks
		WHERE is_active = true AND (title ILIKE $1 OR description ILIKE $1 OR code ILIKE $1)
		ORDER BY score DESC
		LIMIT $2
	`
	rbRows, err := r.db.QueryContext(ctx, rbQuery, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search runbooks: %w", err)
	}
	defer rbRows.Close()
	for rbRows.Next() {
		var res model.KnowledgeSearchResult
		var updatedAt time.Time

		if err := rbRows.Scan(
			&res.ID,
			&res.Type,
			&res.Title,
			&res.Snippet,
			&res.Category,
			&res.SlugOrCode,
			&res.ViewCount,
			&res.Score,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan runbook search result: %w", err)
		}
		res.Tags = []string{"SOP", "Runbook"}
		res.UpdatedTime = updatedAt.Format(time.RFC3339)
		results = append(results, res)
	}
	if err := rbRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runbook search results: %w", err)
	}

	if results == nil {
		results = []model.KnowledgeSearchResult{}
	}
	return results, nil
}

func (r *postgresRepository) GetStats(ctx context.Context) (*model.KnowledgeStats, error) {
	if r.db == nil {
		return nil, errors.New("database connection not available")
	}

	var stats model.KnowledgeStats
	queries := []struct {
		query string
		dest  *int
	}{
		{"SELECT COUNT(*) FROM knowledge_articles WHERE is_published = true", &stats.TotalArticles},
		{"SELECT COUNT(*) FROM knowledge_categories", &stats.TotalCategories},
		{"SELECT COUNT(*) FROM runbooks WHERE is_active = true", &stats.TotalRunbooks},
		{"SELECT COALESCE(SUM(view_count), 0) FROM knowledge_articles", &stats.TotalViews},
	}
	for _, item := range queries {
		if err := r.db.QueryRowContext(ctx, item.query).Scan(item.dest); err != nil {
			return nil, fmt.Errorf("query knowledge statistics: %w", err)
		}
	}

	return &stats, nil
}
