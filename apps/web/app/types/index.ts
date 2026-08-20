/**
 * Common API response and entity types for EOMP.
 */

/** Standard API response wrapper */
export interface ApiResponse<T> {
  data: T
  message?: string
}

/** Paginated response */
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

/** Health check response */
export interface HealthResponse {
  status: 'ok' | 'error'
  service: string
  version: string
}

/** User entity */
export interface User {
  id: string
  email: string
  full_name: string
  role: 'ROLE_ADMIN' | 'ROLE_MANAGER' | 'ROLE_AGENT' | 'ROLE_EMPLOYEE' | string
  department_id?: string | null
  is_active: boolean
  created_at: string
}

/** Auth login/refresh response */
export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

/** Department entity */
export interface Department {
  id: string
  name: string
  code: string
  manager_id?: string | null
  parent_id?: string | null
  created_at: string
  updated_at: string
}

/** Employee entity */
export interface Employee {
  id: string
  user_id?: string | null
  first_name: string
  last_name: string
  full_name: string
  email: string
  phone?: string | null
  job_title: string
  department_id?: string | null
  department_name?: string | null
  department_code?: string | null
  manager_id?: string | null
  manager_name?: string | null
  status: 'ACTIVE' | 'ON_LEAVE' | 'PROBATION' | 'TERMINATED' | string
  location: string
  joined_at: string
  created_at: string
  updated_at: string
}

/** Create Employee Payload */
export interface CreateEmployeePayload {
  first_name: string
  last_name: string
  email: string
  phone?: string
  job_title: string
  department_id?: string
  manager_id?: string
  status?: string
  location?: string
  joined_at?: string
}

/** Service Category */
export interface ServiceCategory {
  id: string
  name: string
  icon: string
  description?: string | null
  items?: ServiceCatalogItem[]
  created_at: string
}

/** Service Catalog Item */
export interface ServiceCatalogItem {
  id: string
  category_id: string
  category_name?: string | null
  name: string
  code: string
  description?: string | null
  default_priority: 'URGENT' | 'HIGH' | 'MEDIUM' | 'LOW' | string
  sla_response_minutes: number
  sla_resolution_minutes: number
  requires_approval: boolean
  is_active: boolean
  created_at: string
}

/** Ticket entity */
export interface Ticket {
  id: string
  ticket_number: string
  title: string
  description: string
  service_item_id?: string | null
  category: string
  priority: 'URGENT' | 'HIGH' | 'MEDIUM' | 'LOW' | string
  status: 'OPEN' | 'ASSIGNED' | 'IN_PROGRESS' | 'WAITING_USER' | 'RESOLVED' | 'CLOSED' | string
  requester_id: string
  requester_name: string
  requester_email: string
  assignee_id?: string | null
  assignee_name?: string | null
  department_id?: string | null
  affected_ci_id?: string | null
  sla_response_deadline: string
  sla_resolution_deadline: string
  responded_at?: string | null
  resolved_at?: string | null
  closed_at?: string | null
  sla_status: 'WITHIN_SLA' | 'WARNING' | 'BREACHED' | string
  created_at: string
  updated_at: string
}

/** Create Ticket Payload */
export interface CreateTicketPayload {
  title: string
  description: string
  service_item_id?: string
  category: string
  priority: string
  requester_id?: string
  requester_name?: string
  requester_email?: string
  department_id?: string
  affected_ci_id?: string
}

/** Ticket Comment */
export interface TicketComment {
  id: string
  ticket_id: string
  author_id: string
  author_name: string
  author_role: string
  content: string
  is_internal: boolean
  created_at: string
}

/** Ticket Timeline record */
export interface TicketTimeline {
  id: string
  ticket_id: string
  actor_id: string
  actor_name: string
  action: string
  old_value?: string | null
  new_value?: string | null
  notes?: string | null
  created_at: string
}

/** Asset entity */
export interface Asset {
  id: string
  asset_tag: string
  name: string
  category: 'LAPTOP' | 'DESKTOP' | 'SERVER' | 'MONITOR' | 'MOBILE' | 'NETWORK' | 'LICENSE' | 'OTHER' | string
  model?: string | null
  serial_number?: string | null
  purchase_date?: string | null
  purchase_cost: number
  warranty_expiry?: string | null
  current_value: number
  status: 'PROCUREMENT' | 'RECEIVED' | 'IN_STOCK' | 'ASSIGNED' | 'IN_USE' | 'MAINTENANCE' | 'RETIRED' | 'DISPOSED' | string
  location: string
  assigned_to_user_id?: string | null
  assigned_to_user_name?: string | null
  assigned_at?: string | null
  notes?: string | null
  created_at: string
  updated_at: string
}

/** Create Asset Payload */
export interface CreateAssetPayload {
  asset_tag: string
  name: string
  category: string
  model?: string
  serial_number?: string
  purchase_date?: string
  purchase_cost: number
  warranty_expiry?: string
  location: string
  notes?: string
}

/** Asset Assignment */
export interface AssetAssignment {
  id: string
  asset_id: string
  user_id: string
  user_name: string
  department_id?: string | null
  assigned_at: string
  returned_at?: string | null
  condition_on_assign: string
  condition_on_return?: string | null
  notes?: string | null
}

/** Asset Stats */
export interface AssetStats {
  total_assets: number
  in_use: number
  in_stock: number
  in_maintenance: number
  total_value: number
}

/** CMDB Configuration Item */
export interface ConfigurationItem {
  id: string
  ci_code: string
  name: string
  ci_type: 'APPLICATION' | 'API_SERVICE' | 'SERVER' | 'DATABASE' | 'NETWORK_DEVICE' | 'CLOUD_RESOURCE' | string
  environment: 'PRODUCTION' | 'STAGING' | 'DEVELOPMENT' | 'DR' | string
  owner_id?: string | null
  owner_name?: string | null
  status: 'OPERATIONAL' | 'DEGRADED' | 'OFFLINE' | 'MAINTENANCE' | string
  ip_address?: string | null
  asset_id?: string | null
  description?: string | null
  created_at: string
  updated_at: string
}

/** Create CI Payload */
export interface CreateCIPayload {
  ci_code: string
  name: string
  ci_type: string
  environment: string
  owner_id?: string
  owner_name?: string
  ip_address?: string
  asset_id?: string
  description?: string
}

/** CMDB CI Relationship */
export interface CIRelationship {
  id: string
  parent_ci_id: string
  parent_ci_name?: string | null
  parent_ci_type?: string | null
  child_ci_id: string
  child_ci_name?: string | null
  child_ci_type?: string | null
  relationship_type: 'DEPENDS_ON' | 'RUNS_ON' | 'CONNECTS_TO' | 'BACKED_UP_BY' | string
  impact_weight: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | string
  created_at: string
}

/** CMDB Topology Graph */
export interface CMDBTopologyGraph {
  nodes: ConfigurationItem[]
  edges: CIRelationship[]
}

/** Workflow Definition */
export interface WorkflowDefinition {
  id: string
  code: string
  name: string
  description?: string | null
  category: string
  trigger_type: string
  is_active: boolean
  steps_config: string
  created_at: string
  updated_at: string
}

/** Workflow Instance */
export interface WorkflowInstance {
  id: string
  instance_number: string
  definition_id: string
  definition_name: string
  entity_type: string
  entity_id: string
  title: string
  requester_id: string
  requester_name: string
  requester_email: string
  current_step_name: string
  status: 'PENDING' | 'RUNNING' | 'WAITING_APPROVAL' | 'APPROVED' | 'REJECTED' | 'COMPLETED' | 'FAILED' | 'CANCELLED' | string
  context_data?: string | null
  started_at: string
  completed_at?: string | null
  created_at: string
  updated_at: string
}

/** Approval Request */
export interface ApprovalRequest {
  id: string
  instance_id: string
  step_id?: string | null
  title: string
  approver_id: string
  approver_name: string
  approver_role: string
  approval_level: number
  status: 'PENDING' | 'APPROVED' | 'REJECTED' | string
  decision_notes?: string | null
  decided_at?: string | null
  sla_deadline: string
  created_at: string
}

/** Workflow Log */
export interface WorkflowLog {
  id: string
  instance_id: string
  actor_id: string
  actor_name: string
  action: string
  message: string
  created_at: string
}

/** Create Workflow Instance Payload */
export interface CreateWorkflowInstancePayload {
  definition_id: string
  entity_type: string
  entity_id: string
  title: string
  requester_id?: string
  requester_name?: string
  requester_email?: string
  context_data?: string
}

/** Approval Decision Payload */
export interface ApprovalDecisionPayload {
  decision: 'APPROVED' | 'REJECTED' | string
  notes: string
}

/** Workflow Stats */
export interface WorkflowStats {
  total_definitions: number
  active_instances: number
  pending_approvals: number
  completed_today: number
}

/** Notification entity */
export interface Notification {
  id: string
  recipient_id: string
  recipient_email: string
  title: string
  message: string
  category: 'INCIDENT' | 'APPROVAL' | 'ASSET' | 'SLA' | 'SECURITY' | 'SYSTEM' | string
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT' | string
  is_read: boolean
  read_at?: string | null
  channel: string
  metadata?: string | null
  created_at: string
}

/** Create Notification Payload */
export interface CreateNotificationPayload {
  recipient_id?: string
  recipient_email: string
  title: string
  message: string
  category?: string
  priority?: string
  channel?: string
  metadata?: string
}

/** Notification Stats */
export interface NotificationStats {
  total: number
  unread: number
  incidents: number
  approvals: number
}

/** Navigation menu item */
export interface MenuItem {
  label: string
  icon?: string
  to?: string
  children?: MenuItem[]
  badge?: string | number
}

/** Knowledge Category */
export interface KnowledgeCategory {
  id: string
  name: string
  code: string
  icon: string
  description?: string | null
  created_at?: string
  updated_at?: string
}

/** Knowledge Article */
export interface KnowledgeArticle {
  id: string
  category_id: string
  category_name?: string | null
  category_code?: string | null
  title: string
  slug: string
  summary: string
  content: string
  tags: string[]
  author_id: string
  author_name: string
  view_count: number
  helpful_count: number
  is_published: boolean
  created_at: string
  updated_at: string
}

/** Create Article Payload */
export interface CreateArticlePayload {
  category_id: string
  title: string
  slug?: string
  summary: string
  content: string
  tags?: string[]
  author_id?: string
  author_name?: string
  is_published?: boolean
}

/** Runbook Step */
export interface RunbookStep {
  step: number
  action: string
  command?: string
  expected?: string
}

/** Knowledge Runbook */
export interface KnowledgeRunbook {
  id: string
  code: string
  title: string
  category: string
  description: string
  prerequisites: string
  steps_json: RunbookStep[]
  rollback_steps: string
  author_name: string
  is_active: boolean
  created_at: string
  updated_at: string
}

/** Create Runbook Payload */
export interface CreateRunbookPayload {
  code: string
  title: string
  category: string
  description: string
  prerequisites: string
  steps_json: RunbookStep[]
  rollback_steps: string
  author_name?: string
}

/** Knowledge Search Result */
export interface KnowledgeSearchResult {
  id: string
  type: 'article' | 'runbook' | string
  title: string
  snippet: string
  category: string
  score: number
  tags?: string[]
  slug_or_code: string
  view_count?: number
  updated_time: string
}

/** Knowledge Stats */
export interface KnowledgeStats {
  total_articles: number
  total_categories: number
  total_runbooks: number
  total_views: number
}

/** AI Message */
export interface AIChatMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  citations?: AICitation[]
  confidence?: number
  fallback_mode?: boolean
  timestamp?: string
}

/** AI Citation */
export interface AICitation {
  article_id: string
  title: string
  score: number
  category?: string
  type?: 'article' | 'runbook' | string
}

/** AI Chat Request */
export interface AIChatRequest {
  session_id?: string
  messages: Array<{
    role: string
    content: string
  }>
}

/** AI Chat Response */
export interface AIChatResponse {
  answer: string
  citations?: AICitation[]
  confidence: number
  tokens_used: number
  fallback_mode?: boolean
}

/** AI Ticket Analysis */
export interface AITicketAnalysis {
  ticket_id: string
  suggested_category: string
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'URGENT' | string
  summary: string
  root_cause: string
  suggested_resolution: string
  confidence: number
  citations?: AICitation[]
  requires_human_review: boolean
  created_at: string
}

/** ITIL Problem Record */
export interface Problem {
  id: string
  problem_number: string
  title: string
  description: string
  category: string
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  status: 'OPEN' | 'UNDER_INVESTIGATION' | 'WORKAROUND_FOUND' | 'KNOWN_ERROR' | 'RESOLVED' | 'CLOSED' | string
  impact: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  urgency: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  assignee_id?: string | null
  assignee_name?: string | null
  root_cause?: string | null
  workaround?: string | null
  resolution?: string | null
  is_known_error: boolean
  linked_count: number
  created_at: string
  updated_at: string
  resolved_at?: string | null
  closed_at?: string | null
}

/** Problem Incident Link */
export interface ProblemIncidentLink {
  id: string
  problem_id: string
  ticket_id: string
  ticket_number: string
  ticket_title: string
  linked_by: string
  linked_at: string
}

/** Link Incident Payload */
export interface LinkIncidentPayload {
  ticket_id: string
  linked_by?: string
}

/** Create Problem Payload */
export interface CreateProblemPayload {
  title: string
  description: string
  category: string
  priority: string
  impact: string
  urgency: string
  assignee_id?: string
  assignee_name?: string
  root_cause?: string
  workaround?: string
  is_known_error?: boolean
  ticket_ids?: string[]
}

/** Update Problem Status Payload */
export interface UpdateProblemStatusPayload {
  status: string
  resolution?: string
  notes?: string
}

/** Update Problem RCA Payload */
export interface UpdateProblemRCAPayload {
  root_cause?: string
  workaround?: string
  is_known_error?: boolean
}

/** Problem Stats */
export interface ProblemStats {
  total_problems: number
  under_investigation: number
  known_errors: number
  resolved_problems: number
  total_linked_tickets: number
}

/** ITIL Change Request (RFC) */
export interface ChangeRequest {
  id: string
  change_number: string
  title: string
  description: string
  change_type: 'STANDARD' | 'NORMAL' | 'EMERGENCY' | 'MAJOR' | string
  category: string
  priority: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  risk_level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  impact_level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  probability_level: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL' | string
  status: 'DRAFT' | 'SUBMITTED' | 'CAB_REVIEW' | 'APPROVED' | 'REJECTED' | 'SCHEDULED' | 'IMPLEMENTING' | 'COMPLETED' | 'FAILED' | 'CANCELLED' | string
  requester_id: string
  requester_name: string
  requester_email: string
  assigned_to_id?: string | null
  assigned_to_name?: string | null
  reason_for_change: string
  implementation_plan: string
  rollback_plan: string
  test_plan: string
  scheduled_start_time?: string | null
  scheduled_end_time?: string | null
  actual_start_time?: string | null
  actual_end_time?: string | null
  downtime_required: boolean
  downtime_minutes: number
  cab_required_count: number
  cab_approved_count: number
  created_at: string
  updated_at: string
}

/** CAB Review Vote */
export interface CABReview {
  id: string
  change_id: string
  reviewer_id: string
  reviewer_name: string
  reviewer_role: string
  vote: 'APPROVED' | 'REJECTED' | 'ABSTAIN' | string
  comments?: string | null
  reviewed_at: string
}

/** Create Change Payload */
export interface CreateChangePayload {
  title: string
  description: string
  change_type: string
  category: string
  priority: string
  impact_level: string
  probability_level: string
  requester_id: string
  requester_name: string
  requester_email: string
  assigned_to_id?: string
  assigned_to_name?: string
  reason_for_change: string
  implementation_plan: string
  rollback_plan: string
  test_plan: string
  scheduled_start_time?: string
  scheduled_end_time?: string
  downtime_required: boolean
  downtime_minutes: number
}

/** Update Change Status Payload */
export interface UpdateChangeStatusPayload {
  status: string
  notes?: string
}

/** Submit CAB Vote Payload */
export interface SubmitCABVotePayload {
  reviewer_id: string
  reviewer_name: string
  reviewer_role: string
  vote: string
  comments?: string
}

/** Change Calendar Item */
export interface ChangeCalendarItem {
  id: string
  change_number: string
  title: string
  change_type: string
  category: string
  risk_level: string
  status: string
  scheduled_start?: string | null
  scheduled_end?: string | null
  downtime_required: boolean
  downtime_minutes: number
}

/** Change Stats */
export interface ChangeStats {
  active_changes: number
  pending_cab_review: number
  emergency_changes: number
  success_rate_percent: number
  total_this_month: number
}

/** Microservice Health Status for SRE Monitoring */
export interface ServiceHealthStatus {
  id: string
  name: string
  category: string
  port: number
  status: 'ONLINE' | 'DEGRADED' | 'OFFLINE' | string
  uptime_pct: number
  latency_ms: number
  cpu_pct: number
  memory_mb: number
  version: string
  error_rate_pct: number
  total_requests: number
  last_probe_time: string
}

/** Cluster Overview Metrics */
export interface ClusterOverview {
  total_services: number
  online_services: number
  degraded_services: number
  offline_services: number
  cluster_health_pct: number
  total_requests_per_min: number
  avg_latency_p95_ms: number
  error_rate_pct: number
}

/** SRE Live Structured Log Entry */
export interface LogEntry {
  id: string
  timestamp: string
  service: string
  level: 'INFO' | 'WARN' | 'ERROR' | 'FATAL' | string
  message: string
  request_id?: string
  caller?: string
}

/** Executive BI & SLA Overview */
export interface ExecutiveOverview {
  avg_mttr_minutes: number
  avg_mttd_minutes: number
  sla_compliance_pct: number
  fcr_rate_pct: number
  csat_rating: number
  total_incidents: number
  total_resolved: number
  total_breached: number
  mttr_improvement_pct: number
  period_label: string
}

/** Incident Daily Trend */
export interface IncidentTrend {
  date: string
  opened_count: number
  resolved_count: number
  sla_compliance_pct: number
}

/** Category Metrics Breakdown */
export interface CategoryBreakdown {
  category_name: string
  category_code: string
  icon: string
  total_count: number
  resolved_count: number
  avg_resolution_minutes: number
  share_pct: number
}

/** Department SLA Metric */
export interface DepartmentSLAMetric {
  department_name: string
  department_code: string
  total_tickets: number
  within_sla_count: number
  breached_sla_count: number
  sla_compliance_pct: number
  avg_mttr_minutes: number
}

/** Agent Performance Scorecard */
export interface AgentScorecard {
  agent_id: string
  agent_name: string
  agent_avatar: string
  job_title: string
  department: string
  tickets_assigned: number
  tickets_resolved: number
  avg_mttr_minutes: number
  csat_rating: number
  sla_compliance_pct: number
}

/** Export Report Response */
export interface ExportReportResponse {
  filename: string
  mime_type: string
  content_base64: string
  total_records: number
  generated_at: string
  generation_time_ms: number
}

