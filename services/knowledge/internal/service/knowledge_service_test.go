package service

import (
	"context"
	"testing"
	"time"

	"eomp/services/knowledge/internal/model"
	"eomp/services/knowledge/internal/repository"
)

// Mock repository for unit testing
type mockRepository struct {
	categories []model.KnowledgeCategory
	articles   []model.KnowledgeArticle
	runbooks   []model.KnowledgeRunbook
}

func newMockRepo() *mockRepository {
	return &mockRepository{
		categories: []model.KnowledgeCategory{
			{ID: "c1", Name: "IT Security", Code: "sec", Icon: "i-lucide-shield-check"},
			{ID: "c2", Name: "Network", Code: "net", Icon: "i-lucide-network"},
		},
		articles: []model.KnowledgeArticle{
			{
				ID:          "a1",
				CategoryID:  "c1",
				Title:       "How to Reset User MFA Tokens",
				Slug:        "how-to-reset-user-mfa-tokens",
				Summary:     "SOP for MFA token resets in Okta.",
				Content:     "Step 1: Verify identity. Step 2: Open Okta.",
				Tags:        []string{"MFA", "Security"},
				AuthorName:  "Security Team",
				ViewCount:   100,
				IsPublished: true,
				Version:     1,
				CreatedAt:   time.Now(),
			},
		},
		runbooks: []model.KnowledgeRunbook{
			{
				ID:          "r1",
				Code:        "RB-SEC-02",
				Title:       "User MFA Token Reset SOP",
				Category:    "IT Security",
				Description: "SOP for resetting MFA tokens",
				StepsJSON: []model.RunbookStep{
					{Step: 1, Action: "Verify Identity"},
				},
				IsActive: true,
			},
		},
	}
}

func (m *mockRepository) ListCategories(ctx context.Context) ([]model.KnowledgeCategory, error) {
	return m.categories, nil
}
func (m *mockRepository) CreateCategory(ctx context.Context, cat *model.KnowledgeCategory) error {
	cat.ID = "c_new"
	m.categories = append(m.categories, *cat)
	return nil
}
func (m *mockRepository) ListArticles(ctx context.Context, query model.ArticleListQuery) (*model.ArticleListResponse, error) {
	return &model.ArticleListResponse{
		Data:       m.articles,
		Total:      len(m.articles),
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil
}
func (m *mockRepository) FindArticleByID(ctx context.Context, id string) (*model.KnowledgeArticle, error) {
	return m.FindArticleByIDForVisibility(ctx, id, false)
}
func (m *mockRepository) FindArticleByIDForVisibility(ctx context.Context, id string, publishedOnly bool) (*model.KnowledgeArticle, error) {
	for _, a := range m.articles {
		if a.ID == id && (!publishedOnly || a.IsPublished) {
			return &a, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) FindArticleBySlug(ctx context.Context, slug string) (*model.KnowledgeArticle, error) {
	return m.FindArticleBySlugForVisibility(ctx, slug, false)
}
func (m *mockRepository) FindArticleBySlugForVisibility(ctx context.Context, slug string, publishedOnly bool) (*model.KnowledgeArticle, error) {
	for _, a := range m.articles {
		if a.Slug == slug && (!publishedOnly || a.IsPublished) {
			return &a, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) CreateArticle(ctx context.Context, art *model.KnowledgeArticle) error {
	art.ID = "a_new"
	m.articles = append(m.articles, *art)
	return nil
}
func (m *mockRepository) UpdateArticle(ctx context.Context, art *model.KnowledgeArticle) error {
	for i, a := range m.articles {
		if a.ID == art.ID {
			if a.Version != art.Version {
				return repository.ErrVersionConflict
			}
			art.Version++
			m.articles[i] = *art
			return nil
		}
	}
	return nil
}
func (m *mockRepository) DeleteArticle(ctx context.Context, id string, expectedVersion int) error {
	var remaining []model.KnowledgeArticle
	for _, a := range m.articles {
		if a.ID == id && a.Version != expectedVersion {
			return repository.ErrVersionConflict
		}
		if a.ID != id {
			remaining = append(remaining, a)
		}
	}
	m.articles = remaining
	return nil
}
func (m *mockRepository) IncrementArticleViews(ctx context.Context, id string) error {
	return nil
}
func (m *mockRepository) ListRunbooks(ctx context.Context, category, search string, page, pageSize int) (*model.RunbookListResponse, error) {
	return &model.RunbookListResponse{
		Data:       m.runbooks,
		Total:      len(m.runbooks),
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil
}
func (m *mockRepository) FindRunbookByID(ctx context.Context, id string) (*model.KnowledgeRunbook, error) {
	for _, r := range m.runbooks {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) FindRunbookByCode(ctx context.Context, code string) (*model.KnowledgeRunbook, error) {
	for _, r := range m.runbooks {
		if r.Code == code {
			return &r, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) CreateRunbook(ctx context.Context, rb *model.KnowledgeRunbook) error {
	rb.ID = "r_new"
	m.runbooks = append(m.runbooks, *rb)
	return nil
}
func (m *mockRepository) Search(ctx context.Context, keyword, category string, limit int, publishedOnly bool) ([]model.KnowledgeSearchResult, error) {
	return []model.KnowledgeSearchResult{
		{
			ID:         "a1",
			Type:       "article",
			Title:      "How to Reset User MFA Tokens",
			Snippet:    "SOP for MFA token resets in Okta.",
			Category:   "IT Security",
			Score:      0.95,
			SlugOrCode: "how-to-reset-user-mfa-tokens",
		},
	}, nil
}
func (m *mockRepository) GetStats(ctx context.Context) (*model.KnowledgeStats, error) {
	return &model.KnowledgeStats{
		TotalArticles:   len(m.articles),
		TotalCategories: len(m.categories),
		TotalRunbooks:   len(m.runbooks),
		TotalViews:      100,
	}, nil
}

func TestKnowledgeService_ListArticles(t *testing.T) {
	repo := newMockRepo()
	svc := NewKnowledgeService(repo)

	res, err := svc.ListArticles(context.Background(), model.ArticleListQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("expected 1 article, got %d", res.Total)
	}
}

func TestKnowledgeService_CreateArticle(t *testing.T) {
	repo := newMockRepo()
	svc := NewKnowledgeService(repo)

	req := &model.CreateArticleRequest{
		CategoryID: "c1",
		Title:      "New VPN Setup Guide",
		Content:    "Complete guide for setting up VPN.",
	}
	art, err := svc.CreateArticle(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if art.Slug != "new-vpn-setup-guide" {
		t.Errorf("expected slug 'new-vpn-setup-guide', got '%s'", art.Slug)
	}
}

func TestKnowledgeService_GetArticleBySlug(t *testing.T) {
	repo := newMockRepo()
	svc := NewKnowledgeService(repo)

	art, err := svc.GetArticleBySlug(context.Background(), "how-to-reset-user-mfa-tokens")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if art == nil || art.Title != "How to Reset User MFA Tokens" {
		t.Errorf("expected MFA article, got %v", art)
	}
}

func TestKnowledgeService_RequiresVersionForArticleWrites(t *testing.T) {
	repo := newMockRepo()
	svc := NewKnowledgeService(repo)
	ctx := context.Background()

	if _, err := svc.UpdateArticle(ctx, "a1", &model.UpdateArticleRequest{}); err == nil {
		t.Fatal("expected update without version to fail")
	}
	if err := svc.DeleteArticle(ctx, "a1", 0); err == nil {
		t.Fatal("expected delete without version to fail")
	}
}

func TestKnowledgeService_Search(t *testing.T) {
	repo := newMockRepo()
	svc := NewKnowledgeService(repo)

	results, err := svc.Search(context.Background(), "MFA", "", 10, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results for 'MFA'")
	}
	if results[0].Score < 0.9 {
		t.Errorf("expected high score, got %f", results[0].Score)
	}
}
