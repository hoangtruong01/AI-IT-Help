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
