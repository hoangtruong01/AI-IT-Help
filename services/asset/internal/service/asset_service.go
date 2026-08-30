package service

import (
	"context"
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/asset/internal/model"
	"eomp/services/asset/internal/repository"
)

// AssetService defines business logic for hardware, software, and equipment lifecycle
type AssetService interface {
	ListAssets(ctx context.Context, query model.AssetListQuery) (*model.AssetListResponse, error)
	GetAsset(ctx context.Context, id string) (*model.Asset, error)
	CreateAsset(ctx context.Context, req *model.CreateAssetRequest) (*model.Asset, error)
	AssignAsset(ctx context.Context, id string, req *model.AssignAssetRequest) error
	ReturnAsset(ctx context.Context, id string, condition string, notes *string, expectedVersion int) error
	UpdateStatus(ctx context.Context, id, status, location string, notes *string, expectedVersion int) error
	GetStats(ctx context.Context) (*model.AssetStatsResponse, error)
	ListAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error)
	GetEmployeeAssetHistory(ctx context.Context, userID string) ([]model.EmployeeAssetHistoryItem, error)
	GetAssetIncidents(ctx context.Context, assetID string) ([]model.AssetIncidentHistoryItem, error)
}

type assetService struct {
	repo               repository.Repository
	helpdeskServiceURL string
	httpClient         *http.Client
	bus                eventbus.EventBus
}

// NewAssetService constructs a new AssetService
func NewAssetService(repo repository.Repository, helpdeskServiceURL string, bus eventbus.EventBus) AssetService {
	if helpdeskServiceURL == "" {
		helpdeskServiceURL = "http://localhost:8084"
	}
	return &assetService{
		repo:               repo,
		helpdeskServiceURL: helpdeskServiceURL,
		httpClient:         &http.Client{Timeout: 5 * time.Second},
		bus:                bus,
	}
}

func (s *assetService) ListAssets(ctx context.Context, query model.AssetListQuery) (*model.AssetListResponse, error) {
	return s.repo.ListAssets(ctx, query)
}

func (s *assetService) GetAsset(ctx context.Context, id string) (*model.Asset, error) {
	if id == "" {
		return nil, errors.BadRequest("asset id is required")
	}
	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to query asset: %v", err))
	}
	if asset == nil {
		return nil, errors.NotFound("asset not found")
	}
	return asset, nil
}

func (s *assetService) CreateAsset(ctx context.Context, req *model.CreateAssetRequest) (*model.Asset, error) {
	if req.AssetTag == "" || req.Name == "" || req.Category == "" {
		return nil, errors.BadRequest("asset_tag, name, and category are required")
	}

	existing, _ := s.repo.FindAssetByTag(ctx, req.AssetTag)
	if existing != nil {
		return nil, errors.Conflict(fmt.Sprintf("asset with tag '%s' already exists", req.AssetTag))
	}

	if req.Location == "" {
		req.Location = "Headquarters Warehouse"
	}

	asset := &model.Asset{
		AssetTag:       req.AssetTag,
		Name:           req.Name,
		Category:       req.Category,
		Model:          req.Model,
		SerialNumber:   req.SerialNumber,
		PurchaseDate:   req.PurchaseDate,
		PurchaseCost:   req.PurchaseCost,
		WarrantyExpiry: req.WarrantyExpiry,
		CurrentValue:   req.PurchaseCost,
		Status:         model.StatusInStock,
		Location:       req.Location,
		Notes:          req.Notes,
	}

	if err := s.repo.CreateAsset(ctx, asset); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create asset: %v", err))
	}

	return asset, nil
}

func (s *assetService) AssignAsset(ctx context.Context, id string, req *model.AssignAssetRequest) error {
	if req.ExpectedVersion <= 0 {
		return errors.BadRequest("version is required for optimistic concurrency control")
	}
	if req.UserID == "" || req.UserName == "" {
		return errors.BadRequest("user_id and user_name are required for assignment")
	}

	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil || asset == nil {
		return errors.NotFound("asset not found")
	}

	if asset.Status == model.StatusInUse {
		return errors.Conflict("asset is already assigned and in use")
	}

	cond := req.ConditionOnAssign
	if cond == "" {
		cond = "GOOD"
	}

	if err := s.repo.AssignAsset(ctx, id, req.UserID, req.UserName, req.DepartmentID, cond, req.Notes, req.ExpectedVersion); err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return errors.Conflict("asset was modified by another request; reload and retry")
		}
		return err
	}

	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "asset",
			Type:   eventbus.TopicAssetAssigned,
			Data: map[string]any{
				"asset_id":      id,
				"asset_tag":     asset.AssetTag,
				"asset_name":    asset.Name,
				"user_id":       req.UserID,
				"user_name":     req.UserName,
				"department_id": req.DepartmentID,
				"condition":     cond,
			},
		})
	}

	return nil
}

func (s *assetService) ReturnAsset(ctx context.Context, id string, condition string, notes *string, expectedVersion int) error {
	if expectedVersion <= 0 {
		return errors.BadRequest("version is required for optimistic concurrency control")
	}
	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil || asset == nil {
		return errors.NotFound("asset not found")
	}

	if cond := condition; cond == "" {
		condition = "GOOD"
	}

	if err := s.repo.ReturnAsset(ctx, id, condition, notes, expectedVersion); err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			return errors.Conflict("asset was modified by another request; reload and retry")
		}
		return err
	}

	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "asset",
			Type:   eventbus.TopicAssetReturned,
			Data: map[string]any{
				"asset_id":   id,
				"asset_tag":  asset.AssetTag,
				"asset_name": asset.Name,
				"condition":  condition,
			},
		})
	}

	return nil
}

func (s *assetService) UpdateStatus(ctx context.Context, id, status, location string, notes *string, expectedVersion int) error {
	if expectedVersion <= 0 {
		return errors.BadRequest("version is required for optimistic concurrency control")
	}
	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil || asset == nil {
		return errors.NotFound("asset not found")
	}
	err = s.repo.UpdateAssetStatus(ctx, id, status, location, notes, expectedVersion)
	if stdErrors.Is(err, repository.ErrVersionConflict) {
		return errors.Conflict("asset was modified by another request; reload and retry")
	}
	return err
}

func (s *assetService) GetStats(ctx context.Context) (*model.AssetStatsResponse, error) {
	return s.repo.GetAssetStats(ctx)
}

func (s *assetService) ListAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error) {
	return s.repo.ListAssetAssignments(ctx, assetID)
}

func (s *assetService) GetEmployeeAssetHistory(ctx context.Context, userID string) ([]model.EmployeeAssetHistoryItem, error) {
	if userID == "" {
		return nil, errors.BadRequest("user_id is required")
	}
	return s.repo.ListAssignmentsByEmployee(ctx, userID)
}

func (s *assetService) GetAssetIncidents(ctx context.Context, assetID string) ([]model.AssetIncidentHistoryItem, error) {
	if assetID == "" {
		return nil, errors.BadRequest("asset_id is required")
	}

	asset, err := s.repo.FindAssetByID(ctx, assetID)
	if err != nil || asset == nil {
		return nil, errors.NotFound("asset not found")
	}

	// Query helpdesk service for tickets associated with this asset ID / CI ID / Asset Tag
	url := fmt.Sprintf("%s/api/v1/tickets/asset/%s", s.helpdeskServiceURL, assetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []model.AssetIncidentHistoryItem{}, nil
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return []model.AssetIncidentHistoryItem{}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []model.AssetIncidentHistoryItem{}, nil
	}

	var responseEnvelope struct {
		Success bool                             `json:"success"`
		Data    []model.AssetIncidentHistoryItem `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&responseEnvelope); err != nil {
		return []model.AssetIncidentHistoryItem{}, nil
	}

	return responseEnvelope.Data, nil
}
