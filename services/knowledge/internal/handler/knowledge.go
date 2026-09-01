package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/knowledge/internal/model"
	"eomp/services/knowledge/internal/service"
)

// KnowledgeHandler exposes REST endpoints for the Knowledge Service.
type KnowledgeHandler struct {
	svc service.KnowledgeService
}

func requireKnowledgeActor(w http.ResponseWriter, r *http.Request) (middleware.Actor, bool) {
	actor := middleware.GetActor(r.Context())
	if !actor.IsValid() {
		errors.WriteHTTP(w, errors.Unauthorized("valid user identity and role are required"))
		return middleware.Actor{}, false
	}
	return actor, true
}

// NewKnowledgeHandler constructs a new KnowledgeHandler.
func NewKnowledgeHandler(svc service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{svc: svc}
}

// GetStats returns aggregate KPI statistics.
func (h *KnowledgeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireKnowledgeActor(w, r); !ok {
		return
	}
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// ListCategories returns all knowledge categories.
func (h *KnowledgeHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireKnowledgeActor(w, r); !ok {
		return
	}
	categories, err := h.svc.ListCategories(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, categories)
}

// CreateCategory creates a new category.
func (h *KnowledgeHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() && !actor.IsManager() {
		errors.WriteHTTP(w, errors.Forbidden("category creation requires manager or admin role"))
		return
	}
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
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	query := model.ArticleListQuery{
		Category:      r.URL.Query().Get("category"),
		Search:        r.URL.Query().Get("search"),
		Page:          1,
		PageSize:      20,
		PublishedOnly: actor.IsEmployee(),
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
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	var art *model.KnowledgeArticle
	var err error

	// If ID looks like UUID, find by ID, else find by slug
	if strings.Contains(id, "-") && len(id) == 36 {
		art, err = h.svc.GetArticleByIDForVisibility(r.Context(), id, actor.IsEmployee())
	} else {
		art, err = h.svc.GetArticleBySlugForVisibility(r.Context(), id, actor.IsEmployee())
	}

	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, art)
}

// CreateArticle creates a new knowledge article.
func (h *KnowledgeHandler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	if actor.IsEmployee() {
		errors.WriteHTTP(w, errors.Forbidden("employees cannot create knowledge articles"))
		return
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		errors.WriteHTTP(w, errors.Forbidden("manager department scope is required"))
		return
	}
	var req model.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}

	// Author and department are trusted identity fields, never client input.
	req.AuthorID = &actor.ID
	authorName := actor.Name
	if authorName == "" {
		authorName = actor.Email
	}
	req.AuthorName = &authorName
	if actor.DepartmentID != "" {
		req.DepartmentID = &actor.DepartmentID
	} else {
		req.DepartmentID = nil
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
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	if actor.IsEmployee() {
		errors.WriteHTTP(w, errors.Forbidden("employees cannot update knowledge articles"))
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	var req model.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}
	current, err := h.svc.GetArticleByID(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	if actor.IsAgent() && current.AuthorID != actor.ID {
		errors.WriteHTTP(w, errors.NotFound("article not found"))
		return
	}
	if actor.IsManager() {
		if actor.DepartmentID == "" {
			errors.WriteHTTP(w, errors.Forbidden("manager department scope is required"))
			return
		}
		if current.DepartmentID != actor.DepartmentID {
			errors.WriteHTTP(w, errors.NotFound("article not found"))
			return
		}
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
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	if !actor.IsAdmin() {
		errors.WriteHTTP(w, errors.Forbidden("article deletion requires admin role"))
		return
	}
	id := r.PathValue("id")
	if id == "" {
		id = strings.TrimPrefix(r.URL.Path, "/api/v1/knowledge/articles/")
	}

	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil || version <= 0 {
		errors.WriteHTTP(w, errors.BadRequest("version query parameter is required and must be a positive integer"))
		return
	}

	if err := h.svc.DeleteArticle(r.Context(), id, version); err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Article deleted successfully"})
}

// ListRunbooks lists SOP Runbooks.
func (h *KnowledgeHandler) ListRunbooks(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireKnowledgeActor(w, r); !ok {
		return
	}
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
	if _, ok := requireKnowledgeActor(w, r); !ok {
		return
	}
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
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	if actor.IsEmployee() {
		errors.WriteHTTP(w, errors.Forbidden("employees cannot create runbooks"))
		return
	}
	var req model.CreateRunbookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("Invalid JSON body"))
		return
	}
	authorName := actor.Name
	if authorName == "" {
		authorName = actor.Email
	}
	req.AuthorName = &authorName

	rb, err := h.svc.CreateRunbook(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, rb)
}

// Search performs semantic and fulltext search across articles and runbooks.
func (h *KnowledgeHandler) Search(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireKnowledgeActor(w, r)
	if !ok {
		return
	}
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("query")
	}
	category := r.URL.Query().Get("category")
	limit := 10
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		limit = l
	}

	results, err := h.svc.Search(r.Context(), q, category, limit, actor.IsEmployee())
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
