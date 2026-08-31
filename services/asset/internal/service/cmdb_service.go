package service

import (
	"context"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/asset/internal/model"
	"eomp/services/asset/internal/repository"
)

// CMDBService defines business logic for Configuration Items and Dependency Topologies
type CMDBService interface {
	ListCIs(ctx context.Context, env, ciType, status string) ([]model.ConfigurationItem, error)
	GetCI(ctx context.Context, id string) (*model.ConfigurationItem, error)
	CreateCI(ctx context.Context, req *model.CreateCIRequest) (*model.ConfigurationItem, error)
	UpdateCIStatus(ctx context.Context, id, status string, expectedVersion int) error
	ListRelationships(ctx context.Context) ([]model.CIRelationship, error)
	CreateRelationship(ctx context.Context, req *model.CreateCIRelationshipRequest) error
	GetTopology(ctx context.Context) (*model.CMDBTopologyGraph, error)
}

type cmdbService struct {
	repo repository.Repository
}

// NewCMDBService constructs a new CMDBService
func NewCMDBService(repo repository.Repository) CMDBService {
	return &cmdbService{repo: repo}
}

func (s *cmdbService) ListCIs(ctx context.Context, env, ciType, status string) ([]model.ConfigurationItem, error) {
	return s.repo.ListCIs(ctx, env, ciType, status)
}

func (s *cmdbService) GetCI(ctx context.Context, id string) (*model.ConfigurationItem, error) {
	if id == "" {
		return nil, errors.BadRequest("ci id is required")
	}
	ci, err := s.repo.FindCIByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "cmdb get configuration item", err)
	}
	if ci == nil {
		return nil, errors.NotFound("configuration item not found")
	}
	return ci, nil
}

func (s *cmdbService) CreateCI(ctx context.Context, req *model.CreateCIRequest) (*model.ConfigurationItem, error) {
	if req.CICode == "" || req.Name == "" || req.CIType == "" {
		return nil, errors.BadRequest("ci_code, name, and ci_type are required")
	}

	if req.Environment == "" {
		req.Environment = model.EnvProduction
	}

	ci := &model.ConfigurationItem{
		CICode:      req.CICode,
		Name:        req.Name,
		CIType:      req.CIType,
		Environment: req.Environment,
		OwnerID:     req.OwnerID,
		OwnerName:   req.OwnerName,
		Status:      model.CIStatusOperational,
		IPAddress:   req.IPAddress,
		AssetID:     req.AssetID,
		Description: req.Description,
	}

	if err := s.repo.CreateCI(ctx, ci); err != nil {
		return nil, errors.Internal(ctx, "cmdb create configuration item", err)
	}

	return s.repo.FindCIByID(ctx, ci.ID)
}

func (s *cmdbService) UpdateCIStatus(ctx context.Context, id, status string, expectedVersion int) error {
	if expectedVersion <= 0 {
		return errors.BadRequest("version is required for CI status updates")
	}
	validStatus := map[string]bool{
		model.CIStatusOperational: true,
		model.CIStatusDegraded:    true,
		model.CIStatusOffline:     true,
		model.CIStatusMaintenance: true,
	}
	if !validStatus[status] {
		return errors.BadRequest("invalid CI status")
	}
	ci, err := s.repo.FindCIByID(ctx, id)
	if err != nil || ci == nil {
		return errors.NotFound("configuration item not found")
	}
	if err := s.repo.UpdateCIStatus(ctx, id, status, expectedVersion); err != nil {
		if err == repository.ErrVersionConflict {
			return errors.Conflict("configuration item was modified by another request")
		}
		return errors.Internal(ctx, "cmdb update configuration item", err)
	}
	return nil
}

func (s *cmdbService) ListRelationships(ctx context.Context) ([]model.CIRelationship, error) {
	return s.repo.ListRelationships(ctx)
}

func (s *cmdbService) CreateRelationship(ctx context.Context, req *model.CreateCIRelationshipRequest) error {
	if req.ParentCIID == "" || req.ChildCIID == "" || req.RelationshipType == "" {
		return errors.BadRequest("parent_ci_id, child_ci_id, and relationship_type are required")
	}

	if req.ImpactWeight == "" {
		req.ImpactWeight = "HIGH"
	}

	rel := &model.CIRelationship{
		ParentCIID:       req.ParentCIID,
		ChildCIID:        req.ChildCIID,
		RelationshipType: req.RelationshipType,
		ImpactWeight:     req.ImpactWeight,
	}

	return s.repo.CreateRelationship(ctx, rel)
}

func (s *cmdbService) GetTopology(ctx context.Context) (*model.CMDBTopologyGraph, error) {
	return s.repo.GetTopology(ctx)
}
