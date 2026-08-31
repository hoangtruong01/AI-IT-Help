package service

import (
	"context"
	stdErrors "errors"
	"fmt"
	"regexp"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/knowledge/internal/model"
	"eomp/services/knowledge/internal/repository"
)

// KnowledgeService defines business logic operations for Knowledge Base and SOP Runbooks.
type KnowledgeService interface {
	ListCategories(ctx context.Context) ([]model.KnowledgeCategory, error)
	CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.KnowledgeCategory, error)

	ListArticles(ctx context.Context, query model.ArticleListQuery) (*model.ArticleListResponse, error)
	GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error)
	GetArticleBySlug(ctx context.Context, slug string) (*model.KnowledgeArticle, error)
	CreateArticle(ctx context.Context, req *model.CreateArticleRequest) (*model.KnowledgeArticle, error)
	UpdateArticle(ctx context.Context, id string, req *model.UpdateArticleRequest) (*model.KnowledgeArticle, error)
	DeleteArticle(ctx context.Context, id string, expectedVersion int) error

	ListRunbooks(ctx context.Context, category, search string, page, pageSize int) (*model.RunbookListResponse, error)
	GetRunbookByID(ctx context.Context, id string) (*model.KnowledgeRunbook, error)
	GetRunbookByCode(ctx context.Context, code string) (*model.KnowledgeRunbook, error)
	CreateRunbook(ctx context.Context, req *model.CreateRunbookRequest) (*model.KnowledgeRunbook, error)

	Search(ctx context.Context, keyword, category string, limit int) ([]model.KnowledgeSearchResult, error)
	GetStats(ctx context.Context) (*model.KnowledgeStats, error)
}

type knowledgeService struct {
	repo repository.Repository
}

// NewKnowledgeService constructs a new KnowledgeService instance.
func NewKnowledgeService(repo repository.Repository) KnowledgeService {
	return &knowledgeService{repo: repo}
}

func (s *knowledgeService) ListCategories(ctx context.Context) ([]model.KnowledgeCategory, error) {
	return s.repo.ListCategories(ctx)
}

func (s *knowledgeService) CreateCategory(ctx context.Context, req *model.CreateCategoryRequest) (*model.KnowledgeCategory, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, errors.BadRequest("Category name is required")
	}
	if strings.TrimSpace(req.Code) == "" {
		req.Code = generateSlug(req.Name)
	}
	if strings.TrimSpace(req.Icon) == "" {
		req.Icon = "i-lucide-folder"
	}

	cat := &model.KnowledgeCategory{
		Name:        req.Name,
		Code:        strings.ToLower(req.Code),
		Icon:        req.Icon,
		Description: req.Description,
	}

	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, errors.Internal(ctx, "knowledge create category", err)
	}
	return cat, nil
}

func (s *knowledgeService) ListArticles(ctx context.Context, query model.ArticleListQuery) (*model.ArticleListResponse, error) {
	return s.repo.ListArticles(ctx, query)
}

func (s *knowledgeService) GetArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.BadRequest("Article ID is required")
	}
	art, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "knowledge find article by ID", err)
	}
	if art == nil {
		return nil, errors.NotFound(fmt.Sprintf("Article %s not found", id))
	}

	// Increment view count asynchronously
	_ = s.repo.IncrementArticleViews(ctx, id)

	return art, nil
}

func (s *knowledgeService) GetArticleBySlug(ctx context.Context, slug string) (*model.KnowledgeArticle, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, errors.BadRequest("Article slug is required")
	}
	art, err := s.repo.FindArticleBySlug(ctx, slug)
	if err != nil {
		return nil, errors.Internal(ctx, "knowledge find article by slug", err)
	}
	if art == nil {
		return nil, errors.NotFound(fmt.Sprintf("Article '%s' not found", slug))
	}

	_ = s.repo.IncrementArticleViews(ctx, art.ID)
	return art, nil
}

func (s *knowledgeService) CreateArticle(ctx context.Context, req *model.CreateArticleRequest) (*model.KnowledgeArticle, error) {
	if req == nil {
		return nil, errors.BadRequest("Request payload cannot be empty")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.BadRequest("Article title is required")
	}
	if strings.TrimSpace(req.CategoryID) == "" {
		return nil, errors.BadRequest("Category ID is required")
	}
	if strings.TrimSpace(req.Content) == "" {
		return nil, errors.BadRequest("Article content is required")
	}

	slug := ""
	if req.Slug != nil && strings.TrimSpace(*req.Slug) != "" {
		slug = generateSlug(*req.Slug)
	} else {
		slug = generateSlug(req.Title)
	}

	summary := req.Summary
	if strings.TrimSpace(summary) == "" {
		if len(req.Content) > 200 {
			summary = req.Content[:200] + "..."
		} else {
			summary = req.Content
		}
	}

	authorID := "system"
	if req.AuthorID != nil && *req.AuthorID != "" {
		authorID = *req.AuthorID
	}
	authorName := "IT Support"
	if req.AuthorName != nil && *req.AuthorName != "" {
		authorName = *req.AuthorName
	}

	isPublished := true
	if req.IsPublished != nil {
		isPublished = *req.IsPublished
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	art := &model.KnowledgeArticle{
		CategoryID:  req.CategoryID,
		Title:       req.Title,
		Slug:        slug,
		Summary:     summary,
		Content:     req.Content,
		Tags:        tags,
		AuthorID:    authorID,
		AuthorName:  authorName,
		IsPublished: isPublished,
	}

	if err := s.repo.CreateArticle(ctx, art); err != nil {
		return nil, errors.Internal(ctx, "knowledge create article", err)
	}

	return art, nil
}

func (s *knowledgeService) UpdateArticle(ctx context.Context, id string, req *model.UpdateArticleRequest) (*model.KnowledgeArticle, error) {
	if req == nil {
		return nil, errors.BadRequest("Request payload cannot be empty")
	}
	if strings.TrimSpace(id) == "" {
		return nil, errors.BadRequest("Article ID is required")
	}
	if req.Version <= 0 {
		return nil, errors.BadRequest("version is required for optimistic concurrency control")
	}

	art, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "knowledge find article for update", err)
	}
	if art == nil {
		return nil, errors.NotFound(fmt.Sprintf("Article %s not found", id))
	}
	if art.Version != req.Version {
		return nil, errors.Conflict("article was modified by another request; reload and retry")
	}

	if req.CategoryID != nil && *req.CategoryID != "" {
		art.CategoryID = *req.CategoryID
	}
	if req.Title != nil && *req.Title != "" {
		art.Title = *req.Title
	}
	if req.Summary != nil && *req.Summary != "" {
		art.Summary = *req.Summary
	}
	if req.Content != nil && *req.Content != "" {
		art.Content = *req.Content
	}
	if req.Tags != nil {
		art.Tags = *req.Tags
	}
	if req.IsPublished != nil {
		art.IsPublished = *req.IsPublished
	}

	if err := s.repo.UpdateArticle(ctx, art); err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return nil, errors.Conflict("article was modified by another request; reload and retry")
		}
		return nil, errors.Internal(ctx, "knowledge update article", err)
	}

	return art, nil
}

func (s *knowledgeService) DeleteArticle(ctx context.Context, id string, expectedVersion int) error {
	if strings.TrimSpace(id) == "" {
		return errors.BadRequest("Article ID is required")
	}
	if expectedVersion <= 0 {
		return errors.BadRequest("version is required for optimistic concurrency control")
	}
	art, err := s.repo.FindArticleByID(ctx, id)
	if err != nil {
		return errors.Internal(ctx, "knowledge find article for delete", err)
	}
	if art == nil {
		return errors.NotFound(fmt.Sprintf("Article %s not found", id))
	}
	if art.Version != expectedVersion {
		return errors.Conflict("article was modified by another request; reload and retry")
	}
	if err := s.repo.DeleteArticle(ctx, id, expectedVersion); err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return errors.Conflict("article was modified by another request; reload and retry")
		}
		return errors.Internal(ctx, "knowledge delete article", err)
	}
	return nil
}

func (s *knowledgeService) ListRunbooks(ctx context.Context, category, search string, page, pageSize int) (*model.RunbookListResponse, error) {
	return s.repo.ListRunbooks(ctx, category, search, page, pageSize)
}

func (s *knowledgeService) GetRunbookByID(ctx context.Context, id string) (*model.KnowledgeRunbook, error) {
	rb, err := s.repo.FindRunbookByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "knowledge find runbook by ID", err)
	}
	if rb == nil {
		return nil, errors.NotFound(fmt.Sprintf("Runbook %s not found", id))
	}
	return rb, nil
}

func (s *knowledgeService) GetRunbookByCode(ctx context.Context, code string) (*model.KnowledgeRunbook, error) {
	rb, err := s.repo.FindRunbookByCode(ctx, code)
	if err != nil {
		return nil, errors.Internal(ctx, "knowledge find runbook by code", err)
	}
	if rb == nil {
		return nil, errors.NotFound(fmt.Sprintf("Runbook with code '%s' not found", code))
	}
	return rb, nil
}

func (s *knowledgeService) CreateRunbook(ctx context.Context, req *model.CreateRunbookRequest) (*model.KnowledgeRunbook, error) {
	if req == nil {
		return nil, errors.BadRequest("Request payload cannot be empty")
	}
	if strings.TrimSpace(req.Code) == "" {
		return nil, errors.BadRequest("Runbook code is required (e.g. RB-NET-05)")
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.BadRequest("Runbook title is required")
	}
	if strings.TrimSpace(req.Category) == "" {
		req.Category = "General"
	}

	authorName := "IT Operations"
	if req.AuthorName != nil && *req.AuthorName != "" {
		authorName = *req.AuthorName
	}

	rb := &model.KnowledgeRunbook{
		Code:          strings.ToUpper(strings.TrimSpace(req.Code)),
		Title:         req.Title,
		Category:      req.Category,
		Description:   req.Description,
		Prerequisites: req.Prerequisites,
		StepsJSON:     req.StepsJSON,
		RollbackSteps: req.RollbackSteps,
		AuthorName:    authorName,
		IsActive:      true,
	}

	if err := s.repo.CreateRunbook(ctx, rb); err != nil {
		return nil, errors.Internal(ctx, "knowledge create runbook", err)
	}

	return rb, nil
}

func (s *knowledgeService) Search(ctx context.Context, keyword, category string, limit int) ([]model.KnowledgeSearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return []model.KnowledgeSearchResult{}, nil
	}
	return s.repo.Search(ctx, keyword, category, limit)
}

func (s *knowledgeService) GetStats(ctx context.Context) (*model.KnowledgeStats, error) {
	return s.repo.GetStats(ctx)
}

// Helper: Generate URL-friendly slug
func generateSlug(input string) string {
	slug := strings.ToLower(input)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}
