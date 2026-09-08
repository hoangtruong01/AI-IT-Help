package service

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/repository"
)

type mockWorkflowRepo struct {
	mu          sync.Mutex
	definitions map[string]*model.WorkflowDefinition
	instances   map[string]*model.WorkflowInstance
	approvals   map[string]*model.ApprovalRequest
	logs        []model.WorkflowLog
	nextInstNum int
	nextAppNum  int
}

func newMockWorkflowRepo() *mockWorkflowRepo {
	return &mockWorkflowRepo{
		definitions: make(map[string]*model.WorkflowDefinition),
		instances:   make(map[string]*model.WorkflowInstance),
		approvals:   make(map[string]*model.ApprovalRequest),
		logs:        make([]model.WorkflowLog, 0),
	}
}

func (m *mockWorkflowRepo) ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]model.WorkflowDefinition, 0, len(m.definitions))
	for _, d := range m.definitions {
		res = append(res, *d)
	}
	return res, nil
}

func (m *mockWorkflowRepo) FindDefinitionByID(ctx context.Context, id string) (*model.WorkflowDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.definitions[id]; ok {
		cp := *d
		return &cp, nil
	}
	return nil, nil
}

func (m *mockWorkflowRepo) FindDefinitionByCode(ctx context.Context, code string) (*model.WorkflowDefinition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.definitions {
		if d.Code == code {
			cp := *d
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *mockWorkflowRepo) ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error) {
	return &model.WorkflowListResponse{}, nil
}

func (m *mockWorkflowRepo) FindInstanceByID(ctx context.Context, id string) (*model.WorkflowInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[id]; ok {
		cp := *inst
		return &cp, nil
	}
	return nil, nil
}

func (m *mockWorkflowRepo) FindInstanceByIDForActor(ctx context.Context, id string, actor middleware.Actor) (*model.WorkflowInstance, error) {
	return m.FindInstanceByID(ctx, id)
}

func (m *mockWorkflowRepo) CreateInstance(ctx context.Context, inst *model.WorkflowInstance) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextInstNum++
	inst.ID = fmt.Sprintf("wfi-%d", m.nextInstNum)
	cp := *inst
	m.instances[inst.ID] = &cp
	return nil
}

func (m *mockWorkflowRepo) CreateInstanceWithApprovalAndLog(ctx context.Context, inst *model.WorkflowInstance, approval *model.ApprovalRequest, log *model.WorkflowLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextInstNum++
	inst.ID = fmt.Sprintf("wfi-%d", m.nextInstNum)
	inst.Version = 1
	cpInst := *inst
	m.instances[inst.ID] = &cpInst

	m.nextAppNum++
	approval.ID = fmt.Sprintf("app-%d", m.nextAppNum)
	approval.InstanceID = inst.ID
	cpApp := *approval
	m.approvals[approval.ID] = &cpApp

	if log != nil {
		log.ID = fmt.Sprintf("log-%d", len(m.logs)+1)
		log.InstanceID = inst.ID
		m.logs = append(m.logs, *log)
	}
	return nil
}

func (m *mockWorkflowRepo) UpdateInstanceStatus(ctx context.Context, id, status, currentStep string, completedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if inst, ok := m.instances[id]; ok {
		inst.Status = status
		inst.CurrentStepName = currentStep
		inst.CompletedAt = completedAt
		inst.Version++
	}
	return nil
}

func (m *mockWorkflowRepo) NextInstanceNumber(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextInstNum++
	return fmt.Sprintf("WFI-%04d", m.nextInstNum), nil
}

func (m *mockWorkflowRepo) ListApprovals(ctx context.Context, approverID, approverRole, departmentID, status string, page, pageSize int) (*model.ApprovalListResponse, error) {
	return &model.ApprovalListResponse{}, nil
}

func (m *mockWorkflowRepo) FindApprovalByID(ctx context.Context, id string) (*model.ApprovalRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if app, ok := m.approvals[id]; ok {
		cp := *app
		return &cp, nil
	}
	return nil, nil
}

func (m *mockWorkflowRepo) CreateApproval(ctx context.Context, app *model.ApprovalRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextAppNum++
	app.ID = fmt.Sprintf("app-%d", m.nextAppNum)
	cp := *app
	m.approvals[app.ID] = &cp
	return nil
}

func (m *mockWorkflowRepo) UpdateApprovalDecision(ctx context.Context, id, status, notes string, decidedAt *time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if app, ok := m.approvals[id]; ok {
		app.Status = status
		app.DecisionNotes = &notes
		app.DecidedAt = decidedAt
	}
	return nil
}

func (m *mockWorkflowRepo) ApplyApprovalDecision(ctx context.Context, approvalID, decision, notes string, decidedAt *time.Time, instanceID string, expectedInstanceVersion int, instanceStatus, currentStep string, completedAt *time.Time, nextApproval *model.ApprovalRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, ok := m.approvals[approvalID]
	if !ok || app.Status != model.ApprovalStatusPending {
		return repository.ErrApprovalConflict
	}
	app.Status = decision
	app.DecisionNotes = &notes
	app.DecidedAt = decidedAt

	inst, ok := m.instances[instanceID]
	if !ok || inst.Version != expectedInstanceVersion {
		return repository.ErrApprovalConflict
	}
	inst.Status = instanceStatus
	inst.CurrentStepName = currentStep
	inst.CompletedAt = completedAt
	inst.Version++

	if nextApproval != nil {
		m.nextAppNum++
		nextApproval.ID = fmt.Sprintf("app-%d", m.nextAppNum)
		nextApproval.CreatedAt = *decidedAt
		cpNext := *nextApproval
		m.approvals[nextApproval.ID] = &cpNext
	}

	return nil
}

func (m *mockWorkflowRepo) AddLog(ctx context.Context, log *model.WorkflowLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	log.ID = fmt.Sprintf("log-%d", len(m.logs)+1)
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockWorkflowRepo) ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]model.WorkflowLog, 0)
	for _, l := range m.logs {
		if l.InstanceID == instanceID {
			res = append(res, l)
		}
	}
	return res, nil
}

func (m *mockWorkflowRepo) GetStats(ctx context.Context) (*model.WorkflowStats, error) {
	return &model.WorkflowStats{}, nil
}

func TestFirstApprovalStep(t *testing.T) {
	step, err := firstApprovalStep(`[{"name":"Prepare","type":"AUTOMATED_ACTION"},{"name":"Manager review","type":"APPROVAL","role":"ROLE_MANAGER"}]`)
	if err != nil {
		t.Fatalf("expected a valid approval step: %v", err)
	}
	if step.Name != "Manager review" || step.Role != "ROLE_MANAGER" {
		t.Fatalf("unexpected approval step: %#v", step)
	}
}

func TestFirstApprovalStepRejectsInvalidConfiguration(t *testing.T) {
	cases := []string{
		`not-json`,
		`[{"name":"Prepare","type":"AUTOMATED_ACTION"}]`,
		`[{"name":"Manager review","type":"APPROVAL"}]`,
	}
	for _, raw := range cases {
		if _, err := firstApprovalStep(raw); err == nil {
			t.Fatalf("expected invalid configuration to fail: %s", raw)
		}
	}
}

func TestMultiStepWorkflow_TwoTiers_AdvanceAndComplete(t *testing.T) {
	ctx := context.Background()
	repo := newMockWorkflowRepo()
	bus := eventbus.NewMemoryEventBus()
	svc := NewWorkflowService(repo, bus)

	// Subscribe to events
	var publishedEvents []eventbus.Event
	var evMu sync.Mutex
	recordEvent := func(ctx context.Context, ev eventbus.Event) error {
		evMu.Lock()
		defer evMu.Unlock()
		publishedEvents = append(publishedEvents, ev)
		return nil
	}
	_ = bus.Subscribe(eventbus.TopicApprovalRequested, recordEvent)
	_ = bus.Subscribe(eventbus.TopicApprovalDecided, recordEvent)
	_ = bus.Subscribe(eventbus.TopicWorkflowCompleted, recordEvent)

	// Definition with 2 approval tiers
	stepsConfig := `[
		{"order": 1, "name": "Department Manager Review", "type": "APPROVAL", "role": "ROLE_MANAGER"},
		{"order": 2, "name": "Pre-flight Verification", "type": "AUTOMATED_ACTION"},
		{"order": 3, "name": "IT Director Final Approval", "type": "APPROVAL", "role": "ROLE_ADMIN"}
	]`
	def := &model.WorkflowDefinition{
		ID:          "def-hardware-multistep",
		Code:        "WF-HW-MULTI",
		Name:        "Hardware Multi-tier Request",
		Category:    "HARDWARE",
		TriggerType: "SERVICE_REQUEST",
		IsActive:    true,
		StepsConfig: stepsConfig,
	}
	repo.definitions[def.ID] = def

	// 1. Start Workflow
	inst, err := svc.StartWorkflow(ctx, &model.CreateInstanceRequest{
		DefinitionID:   def.ID,
		EntityType:     "ticket",
		EntityID:       "tk-test-100",
		Title:          "MacBook Pro M3 Max",
		RequesterID:    "emp-001",
		RequesterName:  "John Developer",
		RequesterEmail: "john@enterprise.local",
	})
	if err != nil {
		t.Fatalf("StartWorkflow failed: %v", err)
	}

	if inst.Status != model.InstanceStatusWaitingApproval {
		t.Fatalf("expected WAITING_APPROVAL, got %s", inst.Status)
	}
	if inst.CurrentStepName != "Department Manager Review" {
		t.Fatalf("expected step Department Manager Review, got %s", inst.CurrentStepName)
	}

	// Verify Tier 1 approval was created
	var tier1App *model.ApprovalRequest
	for _, a := range repo.approvals {
		if a.ApprovalLevel == 1 {
			tier1App = a
			break
		}
	}
	if tier1App == nil {
		t.Fatal("expected Tier 1 approval request in repo")
	}

	// 2. Manager Approves Tier 1
	mgrActor := middleware.Actor{ID: "mgr-001", Role: "ROLE_MANAGER", DepartmentID: "dept-it"}
	err = svc.ProcessApprovalDecision(ctx, tier1App.ID, &model.ApprovalDecisionRequest{
		Decision: model.ApprovalStatusApproved,
		Notes:    "Budget approved by Manager.",
	}, mgrActor, "Alice Manager")
	if err != nil {
		t.Fatalf("ProcessApprovalDecision Tier 1 failed: %v", err)
	}

	// Check instance state: Must NOT be completed yet! Must be advanced to Tier 2 step
	instAfterTier1, _ := repo.FindInstanceByID(ctx, inst.ID)
	if instAfterTier1.Status != model.InstanceStatusWaitingApproval {
		t.Fatalf("expected instance to remain WAITING_APPROVAL after tier 1, got %s", instAfterTier1.Status)
	}
	if instAfterTier1.CurrentStepName != "IT Director Final Approval" {
		t.Fatalf("expected current step to be IT Director Final Approval, got %s", instAfterTier1.CurrentStepName)
	}
	if instAfterTier1.CompletedAt != nil {
		t.Fatalf("expected CompletedAt to be nil after intermediate tier")
	}

	// Verify Tier 2 approval request was automatically created
	var tier2App *model.ApprovalRequest
	for _, a := range repo.approvals {
		if a.ApprovalLevel == 2 {
			tier2App = a
			break
		}
	}
	if tier2App == nil {
		t.Fatal("expected Tier 2 approval request to be created")
	}
	if tier2App.ApproverRole != "ROLE_ADMIN" || tier2App.ApproverName != "IT Director Final Approval" {
		t.Fatalf("unexpected Tier 2 config: %+v", tier2App)
	}

	// 3. Admin Approves Tier 2 (Final Tier)
	adminActor := middleware.Actor{ID: "admin-001", Role: "ROLE_ADMIN"}
	err = svc.ProcessApprovalDecision(ctx, tier2App.ID, &model.ApprovalDecisionRequest{
		Decision: model.ApprovalStatusApproved,
		Notes:    "Procurement signed off by IT Director.",
	}, adminActor, "Bob Director")
	if err != nil {
		t.Fatalf("ProcessApprovalDecision Tier 2 failed: %v", err)
	}

	// Check instance state: Now it MUST be COMPLETED
	instFinal, _ := repo.FindInstanceByID(ctx, inst.ID)
	if instFinal.Status != model.InstanceStatusCompleted {
		t.Fatalf("expected final instance status COMPLETED, got %s", instFinal.Status)
	}
	if instFinal.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set on final completion")
	}

	// Verify that terminal events were published
	time.Sleep(30 * time.Millisecond)
	evMu.Lock()
	defer evMu.Unlock()
	var completedEvCount, decidedEvCount int
	for _, ev := range publishedEvents {
		if ev.Type == eventbus.TopicWorkflowCompleted {
			completedEvCount++
		}
		if ev.Type == eventbus.TopicApprovalDecided {
			decidedEvCount++
			data := ev.Data.(map[string]any)
			if data["decision"] != model.ApprovalStatusApproved {
				t.Fatalf("expected decision APPROVED, got %v", data["decision"])
			}
			if data["entity_id"] != "tk-test-100" {
				t.Fatalf("expected entity_id tk-test-100, got %v", data["entity_id"])
			}
		}
	}

	if completedEvCount != 1 {
		t.Fatalf("expected 1 TopicWorkflowCompleted event, got %d", completedEvCount)
	}
	if decidedEvCount != 1 {
		t.Fatalf("expected 1 TopicApprovalDecided event on final approval, got %d", decidedEvCount)
	}
}

func TestMultiStepWorkflow_RejectedAtTier1(t *testing.T) {
	ctx := context.Background()
	repo := newMockWorkflowRepo()
	bus := eventbus.NewMemoryEventBus()
	svc := NewWorkflowService(repo, bus)

	var publishedEvents []eventbus.Event
	var evMu sync.Mutex
	_ = bus.Subscribe(eventbus.TopicApprovalDecided, func(ctx context.Context, ev eventbus.Event) error {
		evMu.Lock()
		defer evMu.Unlock()
		publishedEvents = append(publishedEvents, ev)
		return nil
	})

	stepsConfig := `[
		{"order": 1, "name": "Department Manager Review", "type": "APPROVAL", "role": "ROLE_MANAGER"},
		{"order": 2, "name": "IT Director Final Approval", "type": "APPROVAL", "role": "ROLE_ADMIN"}
	]`
	def := &model.WorkflowDefinition{
		ID:          "def-hardware-reject",
		Code:        "WF-HW-REJ",
		Name:        "Hardware Request Reject",
		Category:    "HARDWARE",
		TriggerType: "SERVICE_REQUEST",
		IsActive:    true,
		StepsConfig: stepsConfig,
	}
	repo.definitions[def.ID] = def

	inst, err := svc.StartWorkflow(ctx, &model.CreateInstanceRequest{
		DefinitionID:   def.ID,
		EntityType:     "ticket",
		EntityID:       "tk-test-reject",
		Title:          "Expensive Hardware",
		RequesterID:    "emp-002",
		RequesterName:  "Bob Jr",
		RequesterEmail: "bob@enterprise.local",
	})
	if err != nil {
		t.Fatalf("StartWorkflow failed: %v", err)
	}

	var tier1App *model.ApprovalRequest
	for _, a := range repo.approvals {
		if a.ApprovalLevel == 1 {
			tier1App = a
			break
		}
	}

	// Reject at Tier 1
	mgrActor := middleware.Actor{ID: "mgr-001", Role: "ROLE_MANAGER", DepartmentID: "dept-it"}
	err = svc.ProcessApprovalDecision(ctx, tier1App.ID, &model.ApprovalDecisionRequest{
		Decision: model.ApprovalStatusRejected,
		Notes:    "Over budget. Request rejected.",
	}, mgrActor, "Alice Manager")
	if err != nil {
		t.Fatalf("ProcessApprovalDecision rejection failed: %v", err)
	}

	// Instance must be REJECTED immediately, and no Tier 2 created
	instAfter, _ := repo.FindInstanceByID(ctx, inst.ID)
	if instAfter.Status != model.InstanceStatusRejected {
		t.Fatalf("expected status REJECTED, got %s", instAfter.Status)
	}

	for _, a := range repo.approvals {
		if a.ApprovalLevel == 2 {
			t.Fatal("tier 2 approval should NOT have been created after tier 1 rejection")
		}
	}

	time.Sleep(30 * time.Millisecond)
	evMu.Lock()
	defer evMu.Unlock()
	if len(publishedEvents) != 1 {
		t.Fatalf("expected 1 TopicApprovalDecided event, got %d", len(publishedEvents))
	}
	data := publishedEvents[0].Data.(map[string]any)
	if data["decision"] != model.ApprovalStatusRejected {
		t.Fatalf("expected decision REJECTED, got %v", data["decision"])
	}
	if data["entity_id"] != "tk-test-reject" {
		t.Fatalf("expected entity_id tk-test-reject, got %v", data["entity_id"])
	}
}
