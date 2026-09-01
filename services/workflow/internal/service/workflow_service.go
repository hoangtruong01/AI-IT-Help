package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/repository"
)

// WorkflowService handles workflow lifecycle execution and approval matrix processing
type WorkflowService interface {
	ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error)
	GetDefinition(ctx context.Context, id string) (*model.WorkflowDefinition, error)
	ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error)
	GetInstance(ctx context.Context, id string) (*model.WorkflowInstance, error)
	GetInstanceForActor(ctx context.Context, id string, actor middleware.Actor) (*model.WorkflowInstance, error)
	StartWorkflow(ctx context.Context, req *model.CreateInstanceRequest) (*model.WorkflowInstance, error)

	ListApprovals(ctx context.Context, actor middleware.Actor, requestedApproverID, status string, page, pageSize int) (*model.ApprovalListResponse, error)
	ProcessApprovalDecision(ctx context.Context, approvalID string, req *model.ApprovalDecisionRequest, actor middleware.Actor, actorName string) error

	ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error)
	GetStats(ctx context.Context) (*model.WorkflowStats, error)
}

type workflowService struct {
	repo repository.Repository
	bus  eventbus.EventBus
}

// NewWorkflowService constructs a new WorkflowService
func NewWorkflowService(repo repository.Repository, bus eventbus.EventBus) WorkflowService {
	return &workflowService{repo: repo, bus: bus}
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
		return nil, errors.Internal(ctx, "workflow get definition", err)
	}
	if def == nil {
		return nil, errors.NotFound("workflow definition not found")
	}
	return def, nil
}

func (s *workflowService) ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error) {
	actor := middleware.Actor{ID: query.ActorID, Role: query.ActorRole, DepartmentID: query.ActorDepartmentID}
	if !actor.IsValid() {
		return nil, errors.Unauthorized("valid user identity and role are required")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return nil, errors.Forbidden("manager department scope is required")
	}
	return s.repo.ListInstances(ctx, query)
}

func (s *workflowService) GetInstance(ctx context.Context, id string) (*model.WorkflowInstance, error) {
	if id == "" {
		return nil, errors.BadRequest("instance id is required")
	}
	inst, err := s.repo.FindInstanceByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "workflow list definitions", err)
	}
	if inst == nil {
		return nil, errors.NotFound("workflow instance not found")
	}
	return inst, nil
}

func (s *workflowService) GetInstanceForActor(ctx context.Context, id string, actor middleware.Actor) (*model.WorkflowInstance, error) {
	if id == "" {
		return nil, errors.BadRequest("instance id is required")
	}
	if !actor.IsValid() {
		return nil, errors.Unauthorized("valid user identity and role are required")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return nil, errors.Forbidden("manager department scope is required")
	}
	inst, err := s.repo.FindInstanceByIDForActor(ctx, id, actor)
	if err != nil {
		return nil, errors.Internal(ctx, "workflow get scoped instance", err)
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

	approvalStep, err := firstApprovalStep(def.StepsConfig)
	if err != nil {
		return nil, errors.Internal(ctx, "workflow parse approval step", err)
	}

	instNumber, err := s.repo.NextInstanceNumber(ctx)
	if err != nil {
		return nil, errors.Internal(ctx, "workflow generate instance number", err)
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
		DepartmentID:    req.DepartmentID,
		CurrentStepName: approvalStep.Name,
		Status:          model.InstanceStatusWaitingApproval,
		ContextData:     req.ContextData,
		StartedAt:       time.Now(),
	}

	// Create initial Approval Request
	approval := &model.ApprovalRequest{
		Title:         fmt.Sprintf("Approve: %s", inst.Title),
		ApproverID:    approvalStep.Role,
		ApproverName:  approvalStep.Name,
		ApproverRole:  approvalStep.Role,
		ApprovalLevel: 1,
		Status:        model.ApprovalStatusPending,
		SLADeadline:   time.Now().Add(24 * time.Hour),
	}

	// Log start
	workflowLog := &model.WorkflowLog{
		ActorID:   req.RequesterID,
		ActorName: req.RequesterName,
		Action:    "WORKFLOW_STARTED",
		Message:   fmt.Sprintf("Started workflow instance %s for '%s'. Dispatched approval request to %s.", inst.InstanceNumber, inst.Title, approvalStep.Name),
	}
	if err := s.repo.CreateInstanceWithApprovalAndLog(ctx, inst, approval, workflowLog); err != nil {
		return nil, errors.Internal(ctx, "workflow start instance", err)
	}

	// Publish approval.requested event via EventBus
	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "workflow",
			Type:   eventbus.TopicApprovalRequested,
			Data: map[string]any{
				"instance_id":     inst.ID,
				"instance_number": inst.InstanceNumber,
				"title":           inst.Title,
				"requester_id":    inst.RequesterID,
				"requester_email": inst.RequesterEmail,
				"approval_id":     approval.ID,
				"approver_id":     approval.ApproverID,
			},
		})
	}

	return s.repo.FindInstanceByID(ctx, inst.ID)
}

func (s *workflowService) ListApprovals(ctx context.Context, actor middleware.Actor, requestedApproverID, status string, page, pageSize int) (*model.ApprovalListResponse, error) {
	if !actor.IsValid() {
		return nil, errors.Unauthorized("valid user identity and role are required")
	}
	if actor.IsEmployee() || actor.IsAgent() {
		return nil, errors.Forbidden("approval access requires manager or admin role")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return nil, errors.Forbidden("manager department scope is required")
	}

	approverID := actor.ID
	approverRole := actor.Role
	departmentID := ""
	if actor.IsAdmin() {
		approverID = requestedApproverID
		approverRole = ""
	} else if actor.IsManager() {
		departmentID = actor.DepartmentID
	}
	return s.repo.ListApprovals(ctx, approverID, approverRole, departmentID, status, page, pageSize)
}

func (s *workflowService) ProcessApprovalDecision(ctx context.Context, approvalID string, req *model.ApprovalDecisionRequest, actor middleware.Actor, actorName string) error {
	if !actor.IsValid() {
		return errors.Unauthorized("valid user identity and role are required")
	}
	if actor.IsEmployee() || actor.IsAgent() {
		return errors.Forbidden("approval access requires manager or admin role")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return errors.Forbidden("manager department scope is required")
	}
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

	instance, err := s.repo.FindInstanceByIDForActor(ctx, approval.InstanceID, actor)
	if err != nil {
		return errors.Internal(ctx, "workflow find scoped approval instance", err)
	}
	if instance == nil {
		return errors.NotFound("approval request not found")
	}

	if !actor.IsAdmin() && actor.ID != approval.ApproverID && !(approval.ApproverID == approval.ApproverRole && actor.Role == approval.ApproverRole) {
		return errors.Forbidden("approval request is not assigned to this user")
	}

	now := time.Now()
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

	if err := s.repo.ApplyApprovalDecision(
		ctx, approvalID, req.Decision, req.Notes, &now,
		approval.InstanceID, instance.Version, newInstanceStatus, nextStep, completedAt,
	); err != nil {
		if err == repository.ErrApprovalConflict {
			return errors.Conflict("approval was already decided or workflow changed; reload and retry")
		}
		return errors.Internal(ctx, "workflow apply approval decision", err)
	}

	// Log decision
	_ = s.repo.AddLog(ctx, &model.WorkflowLog{
		InstanceID: approval.InstanceID,
		ActorID:    actor.ID,
		ActorName:  actorName,
		Action:     req.Decision,
		Message:    fmt.Sprintf("Approver %s decided '%s' with notes: '%s'", actorName, req.Decision, req.Notes),
	})

	// Publish approval.decided event via EventBus
	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "workflow",
			Type:   eventbus.TopicApprovalDecided,
			Data: map[string]any{
				"approval_id": approvalID,
				"instance_id": approval.InstanceID,
				"decision":    req.Decision,
				"notes":       req.Notes,
				"actor_id":    actor.ID,
				"actor_name":  actorName,
				"status":      newInstanceStatus,
			},
		})
	}

	return nil
}

type configuredWorkflowStep struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Role string `json:"role"`
}

func firstApprovalStep(raw string) (configuredWorkflowStep, error) {
	var steps []configuredWorkflowStep
	if err := json.Unmarshal([]byte(raw), &steps); err != nil {
		return configuredWorkflowStep{}, err
	}
	for _, step := range steps {
		if step.Type == model.StepTypeApproval && step.Name != "" && step.Role != "" {
			return step, nil
		}
	}
	return configuredWorkflowStep{}, fmt.Errorf("approval step is missing")
}

func (s *workflowService) ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error) {
	return s.repo.ListLogs(ctx, instanceID)
}

func (s *workflowService) GetStats(ctx context.Context) (*model.WorkflowStats, error) {
	return s.repo.GetStats(ctx)
}
