package service

import (
	"context"
	"fmt"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/asset/internal/model"
	"eomp/services/asset/internal/repository"
)

// AssetService defines business logic for hardware, software, and equipment lifecycle
type AssetService interface {
	ListAssets(ctx context.Context, query model.AssetListQuery) (*model.AssetListResponse, error)
	GetAsset(ctx context.Context, id string) (*model.Asset, error)
	CreateAsset(ctx context.Context, req *model.CreateAssetRequest) (*model.Asset, error)
	AssignAsset(ctx context.Context, id string, req *model.AssignAssetRequest) error
	ReturnAsset(ctx context.Context, id string, condition string, notes *string) error
	UpdateStatus(ctx context.Context, id, status, location string, notes *string) error
	GetStats(ctx context.Context) (*model.AssetStatsResponse, error)
	ListAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error)
}

type assetService struct {
	repo repository.Repository
}

// NewAssetService constructs a new AssetService
func NewAssetService(repo repository.Repository) AssetService {
	return &assetService{repo: repo}
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

	currentVal := req.PurchaseCost
	if currentVal <= 0 {
		currentVal = 0.00
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
		CurrentValue:   currentVal,
		Status:         model.StatusInStock,
		Location:       req.Location,
		Notes:          req.Notes,
	}

	if err := s.repo.CreateAsset(ctx, asset); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create asset: %v", err))
	}

	return s.repo.FindAssetByID(ctx, asset.ID)
}

func (s *assetService) AssignAsset(ctx context.Context, id string, req *model.AssignAssetRequest) error {
	if req.UserID == "" || req.UserName == "" {
		return errors.BadRequest("user_id and user_name are required")
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

	return s.repo.AssignAsset(ctx, id, req.UserID, req.UserName, req.DepartmentID, cond, req.Notes)
}

func (s *assetService) ReturnAsset(ctx context.Context, id string, condition string, notes *string) error {
	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil || asset == nil {
		return errors.NotFound("asset not found")
	}

	if cond := condition; cond == "" {
		condition = "GOOD"
	}

	return s.repo.ReturnAsset(ctx, id, condition, notes)
}

func (s *assetService) UpdateStatus(ctx context.Context, id, status, location string, notes *string) error {
	asset, err := s.repo.FindAssetByID(ctx, id)
	if err != nil || asset == nil {
		return errors.NotFound("asset not found")
	}
	return s.repo.UpdateAssetStatus(ctx, id, status, location, notes)
}

func (s *assetService) GetStats(ctx context.Context) (*model.AssetStatsResponse, error) {
	return s.repo.GetAssetStats(ctx)
}

func (s *assetService) ListAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error) {
	return s.repo.ListAssetAssignments(ctx, assetID)
}
