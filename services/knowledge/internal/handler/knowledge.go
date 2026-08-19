package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/knowledge/internal/model"
	"eomp/services/knowledge/internal/service"
)

// KnowledgeHandler exposes REST endpoints for the Knowledge Service.
type KnowledgeHandler struct {
	svc service.KnowledgeService
}

// NewKnowledgeHandler constructs a new KnowledgeHandler.
func NewKnowledgeHandler(svc service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

// GetStats returns aggregate KPI statistics.
func (h *KnowledgeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// ListCategories returns all knowledge categories.
func (h *KnowledgeHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, categories)
}

// CreateCategory creates a new category.
func (h *KnowledgeHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}

	cat, err := h.svc.CreateCategory(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, cat)
}

// ListArticles lists articles with category filter, search, and pagination.
func (h *KnowledgeHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	query := model.ArticleListQuery{
		Category: r.URL.Query().Get("category"),
		Search:   r.URL.Query().Get("search"),
		Page:     1,
		PageSize: 20,
	}

	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		query.Page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 {
		query.PageSize = ps
	}

	res, err := h.svc.ListArticles(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}

// GetArticle returns an article by ID or slug.
func (h *KnowledgeHandler) GetArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	var art *model.KnowledgeArticle
	var err error

	// If ID looks like UUID, find by ID, else find by slug
	if strings.Contains(id, "-") && len(id) == 36 {
		art, err = h.svc.GetArticleByID(r.Context(), id)
	} else {
		art, err = h.svc.GetArticleBySlug(r.Context(), id)
	}

	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, art)
}

// CreateArticle creates a new knowledge article.
func (h *KnowledgeHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req model.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}

	// Extract author info from Gateway header if available
	if req.AuthorID == nil || *req.AuthorID == "" {
		userID := r.Header.Get("X-User-ID")
		if userID != "" {
			req.AuthorID = &userID
		}
	}
	if req.AuthorName == nil || *req.AuthorName == "" {
		userName := r.Header.Get("X-User-Email")
		if userName != "" {
			req.AuthorName = &userName
		}
	}

	art, err := h.svc.CreateArticle(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, art)
}

// UpdateArticle updates an existing article.
func (h *KnowledgeHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	var req model.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}

	art, err := h.svc.UpdateArticle(r.Context(), id, &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, art)
}

// DeleteArticle deletes an article.
func (h *KnowledgeHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	if err := h.svc.DeleteArticle(r.Context(), id); err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Article deleted successfully"})
}

// ListRunbooks lists SOP Runbooks.
func (h *KnowledgeHandler) ListRunbooks(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	page := 1
	pageSize := 20

	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 {
		pageSize = ps
	}

	res, err := h.svc.ListRunbooks(r.Context(), category, search, page, pageSize)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}

// GetRunbook returns an SOP Runbook by ID or code.
func (h *KnowledgeHandler) GetRunbook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/runbooks/")
	}

	var rb *model.KnowledgeRunbook
	var err error

	if strings.HasPrefix(strings.ToUpper(id), "RB-") {
		rb, err = h.svc.GetRunbookByCode(r.Context(), id)
	} else {
		rb, err = h.svc.GetRunbookByID(r.Context(), id)
	}

	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rb)
}

// CreateRunbook creates a new SOP Runbook.
func (h *KnowledgeHandler) CreateRunbook(w http.ResponseWriter, r *http.Request) {
	var req model.CreateRunbookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}

	rb, err := h.svc.CreateRunbook(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, rb)
}

// Search performs semantic and fulltext search across articles and runbooks.
func (h *KnowledgeHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("query")
	}
	category := r.URL.Query().Get("category")
	limit := 10
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	results, err := h.svc.Search(r.Context(), q, category, limit)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"query":   q,
		"total":   len(results),
		"results": results,
	})
}
