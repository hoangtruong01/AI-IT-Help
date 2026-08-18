package service

import (
	"context"
	"fmt"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/repository"
)

// WorkflowService handles workflow lifecycle execution and approval matrix processing
type WorkflowService interface {
	ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error)
	GetDefinition(ctx context.Context, id string) (*model.WorkflowDefinition, error)
	ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error)
	GetInstance(ctx context.Context, id string) (*model.WorkflowInstance, error)
	StartWorkflow(ctx context.Context, req *model.CreateInstanceRequest) (*model.WorkflowInstance, error)

	ListApprovals(ctx context.Context, approverID, status string, page, pageSize int) (*model.ApprovalListResponse, error)
	ProcessApprovalDecision(ctx context.Context, approvalID string, req *model.ApprovalDecisionRequest, actorID, actorName string) error

	ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error)
	GetStats(ctx context.Context) (*model.WorkflowStats, error)
}

type workflowService struct {
	repo repository.Repository
}

// NewWorkflowService constructs a new WorkflowService
func NewWorkflowService(repo repository.Repository) WorkflowService {
	return &workflowService{repo: repo}
}

func (s *workflowService) ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error) {
	return s.repo.ListDefinitions(ctx)
}

func (s *workflowService) GetDefinition(ctx context.Context, id string) (*model.WorkflowDefinition, error) {
	if id == "" {
		return nil, errors.BadRequest("definition id is required")
	}
	def, err := s.repo.FindDefinitionByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to get definition: %v", err))
	}
	if def == nil {
		return nil, errors.NotFound("workflow definition not found")
	}
	return def, nil
}

func (s *workflowService) ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error) {
	return s.repo.ListInstances(ctx, query)
}

func (s *workflowService) GetInstance(ctx context.Context, id string) (*model.WorkflowInstance, error) {
	if id == "" {
		return nil, errors.BadRequest("instance id is required")
	}
	inst, err := s.repo.FindInstanceByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to get instance: %v", err))
	}
	if inst == nil {
		return nil, errors.NotFound("workflow instance not found")
	}
	return inst, nil
}

func (s *workflowService) StartWorkflow(ctx context.Context, req *model.CreateInstanceRequest) (*model.WorkflowInstance, error) {
	if req.DefinitionID == "" || req.Title == "" || req.RequesterEmail == "" {
		return nil, errors.BadRequest("definition_id, title, and requester_email are required")
	}

	def, err := s.repo.FindDefinitionByID(ctx, req.DefinitionID)
	if err != nil || def == nil {
		return nil, errors.NotFound("workflow definition not found")
	}

	instNumber, err := s.repo.NextInstanceNumber(ctx)
	if err != nil {
		return nil, errors.InternalServerError("failed to generate instance number")
	}

	inst := &model.WorkflowInstance{
		InstanceNumber:  instNumber,
		DefinitionID:    def.ID,
		DefinitionName:  def.Name,
		EntityType:      req.EntityType,
		EntityID:        req.EntityID,
		Title:           req.Title,
		RequesterID:     req.RequesterID,
		RequesterName:   req.RequesterName,
		RequesterEmail:  req.RequesterEmail,
		CurrentStepName: "Manager Approval",
		Status:          model.InstanceStatusWaitingApproval,
		ContextData:     req.ContextData,
		StartedAt:       time.Now(),
	}

	if err := s.repo.CreateInstance(ctx, inst); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to start workflow: %v", err))
	}

	// Create initial Approval Request
	approval := &model.ApprovalRequest{
		InstanceID:    inst.ID,
		Title:         fmt.Sprintf("Approve: %s", inst.Title),
		ApproverID:    "e0000000-0000-0000-0000-000000000002",
		ApproverName:  "David Tran (IT Manager)",
		ApproverRole:  "ROLE_MANAGER",
		ApprovalLevel: 1,
		Status:        model.ApprovalStatusPending,
		SLADeadline:   time.Now().Add(24 * time.Hour),
	}
	_ = s.repo.CreateApproval(ctx, approval)

	// Log start
	_ = s.repo.AddLog(ctx, &model.WorkflowLog{
		InstanceID: inst.ID,
		ActorID:    req.RequesterID,
		ActorName:  req.RequesterName,
		Action:     "WORKFLOW_STARTED",
		Message:    fmt.Sprintf("Started workflow instance %s for '%s'. Dispatched approval request to Manager.", inst.InstanceNumber, inst.Title),
	})

	return s.repo.FindInstanceByID(ctx, inst.ID)
}

func (s *workflowService) ListApprovals(ctx context.Context, approverID, status string, page, pageSize int) (*model.ApprovalListResponse, error) {
	return s.repo.ListApprovals(ctx, approverID, status, page, pageSize)
}

func (s *workflowService) ProcessApprovalDecision(ctx context.Context, approvalID string, req *model.ApprovalDecisionRequest, actorID, actorName string) error {
	if req.Decision != model.ApprovalStatusApproved && req.Decision != model.ApprovalStatusRejected {
		return errors.BadRequest("decision must be APPROVED or REJECTED")
	}

	approval, err := s.repo.FindApprovalByID(ctx, approvalID)
	if err != nil || approval == nil {
		return errors.NotFound("approval request not found")
	}

	if approval.Status != model.ApprovalStatusPending {
		return errors.Conflict(fmt.Sprintf("approval request is already %s", approval.Status))
	}

	now := time.Now()
	if err := s.repo.UpdateApprovalDecision(ctx, approvalID, req.Decision, req.Notes, &now); err != nil {
		return errors.InternalServerError("failed to update approval decision")
	}

	// Update Workflow Instance state
	var newInstanceStatus string
	var nextStep string
	var completedAt *time.Time

	if req.Decision == model.ApprovalStatusApproved {
		newInstanceStatus = model.InstanceStatusCompleted
		nextStep = "Completed (Approved)"
		completedAt = &now
	} else {
		newInstanceStatus = model.InstanceStatusRejected
		nextStep = "Terminated (Rejected)"
		completedAt = &now
	}

	_ = s.repo.UpdateInstanceStatus(ctx, approval.InstanceID, newInstanceStatus, nextStep, completedAt)

	// Log decision
	_ = s.repo.AddLog(ctx, &model.WorkflowLog{
		InstanceID: approval.InstanceID,
		ActorID:    actorID,
		ActorName:  actorName,
		Action:     req.Decision,
		Message:    fmt.Sprintf("Approver %s decided '%s' with notes: '%s'", actorName, req.Decision, req.Notes),
	})

	return nil
}

func (s *workflowService) ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error) {
	return s.repo.ListLogs(ctx, instanceID)
}

func (s *workflowService) GetStats(ctx context.Context) (*model.WorkflowStats, error) {
	return s.repo.GetStats(ctx)
}
