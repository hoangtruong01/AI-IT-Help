<script setup lang="ts">
import type { WorkflowDefinition, WorkflowInstance, ApprovalRequest, WorkflowLog, WorkflowStats, PaginatedResponse, CreateWorkflowInstancePayload, ApprovalDecisionPayload } from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()
const authStore = useAuthStore()

// State
const activeTab = ref<'approvals' | 'instances' | 'definitions'>('approvals')
const stats = ref<WorkflowStats>({
  total_definitions: 0,
  active_instances: 0,
  pending_approvals: 0,
  completed_today: 0
})
const loading = ref(false)

// Approvals State
const approvals = ref<ApprovalRequest[]>([])
const approvalStatusFilter = ref('PENDING')
const selectedApproval = ref<ApprovalRequest | null>(null)
const isDecisionModalOpen = ref(false)
const decisionType = ref<'APPROVED' | 'REJECTED'>('APPROVED')
const decisionNotes = ref('')
const processingDecision = ref(false)

// Instances State
const instances = ref<WorkflowInstance[]>([])
const instanceSearch = ref('')
const selectedInstance = ref<WorkflowInstance | null>(null)
const isInstanceDetailOpen = ref(false)
const instanceLogs = ref<WorkflowLog[]>([])

// Definitions State
const definitions = ref<WorkflowDefinition[]>([])
const isLaunchModalOpen = ref(false)
const launching = ref(false)
const launchError = ref<string | null>(null)
const newInstance = reactive<CreateWorkflowInstancePayload>({
  definition_id: '',
  entity_type: 'SERVICE_REQUEST',
  entity_id: 'REQ-1001',
  title: '',
  requester_id: '',
  requester_name: '',
  requester_email: ''
})

async function fetchStats() {
  try {
    const res = await api.get<WorkflowStats>('/api/v1/workflows/stats')
    if (res) stats.value = res
  } catch (err) {
    console.error('Failed to load workflow stats:', err)
  }
}

async function fetchApprovals() {
  loading.value = true
  try {
    const res = await api.get<PaginatedResponse<ApprovalRequest>>('/api/v1/approvals', {
      status: approvalStatusFilter.value !== 'All' ? approvalStatusFilter.value : undefined,
      page: 1,
      page_size: 50
    })
    approvals.value = res?.data || []
  } catch (err) {
    console.error('Failed to load approvals:', err)
  } finally {
    loading.value = false
  }
}

async function fetchInstances() {
  loading.value = true
  try {
    const res = await api.get<PaginatedResponse<WorkflowInstance>>('/api/v1/workflows/instances', {
      search: instanceSearch.value || undefined,
      page: 1,
      page_size: 50
    })
    instances.value = res?.data || []
  } catch (err) {
    console.error('Failed to load workflow instances:', err)
  } finally {
    loading.value = false
  }
}

async function fetchDefinitions() {
  try {
    const res = await api.get<WorkflowDefinition[]>('/api/v1/workflows/definitions')
    definitions.value = res || []
  } catch (err) {
    console.error('Failed to load definitions:', err)
  }
}

function openDecisionModal(app: ApprovalRequest, type: 'APPROVED' | 'REJECTED') {
  selectedApproval.value = app
  decisionType.value = type
  decisionNotes.value = type === 'APPROVED' ? 'Approved based on standard policy criteria.' : 'Rejected due to budget constraints or policy mismatch.'
  isDecisionModalOpen.value = true
}

async function handleSubmitDecision() {
  if (!selectedApproval.value) return
  processingDecision.value = true
  try {
    const payload: ApprovalDecisionPayload = {
      decision: decisionType.value,
      notes: decisionNotes.value
    }
    await api.post(`/api/v1/approvals/${selectedApproval.value.id}/decision`, payload)
    isDecisionModalOpen.value = false
    await Promise.all([fetchApprovals(), fetchStats(), fetchInstances()])
  } catch (err) {
    console.error('Failed to record approval decision:', err)
  } finally {
    processingDecision.value = false
  }
}

async function openInstanceDetail(inst: WorkflowInstance) {
  selectedInstance.value = inst
  isInstanceDetailOpen.value = true
  instanceLogs.value = []
  try {
    const res = await api.get<WorkflowLog[]>(`/api/v1/workflows/instances/${inst.id}/logs`)
    instanceLogs.value = res || []
  } catch (err) {
    console.error('Failed to load instance logs:', err)
  }
}

async function handleLaunchWorkflow() {
  if (!newInstance.definition_id || !newInstance.title) {
    launchError.value = 'Blueprint selection and title are required.'
    return
  }

  launching.value = true
  launchError.value = null
  try {
    if (!authStore.user?.id || !authStore.user.email) {
      launchError.value = 'Your authenticated user profile is required to start a workflow.'
      return
    }
    const payload: CreateWorkflowInstancePayload = {
      ...newInstance,
      requester_id: authStore.user.id,
      requester_name: authStore.user.full_name || authStore.user.email,
      requester_email: authStore.user.email
    }

    await api.post('/api/v1/workflows/instances', payload)
    isLaunchModalOpen.value = false
    Object.assign(newInstance, {
      definition_id: '',
      title: ''
    })
    await Promise.all([fetchInstances(), fetchApprovals(), fetchStats()])
  } catch (err: unknown) {
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    launchError.value = errObj?.data?.error?.message || errObj?.message || 'Failed to start workflow instance.'
  } finally {
    launching.value = false
  }
}

watch(approvalStatusFilter, () => {
  fetchApprovals()
})

let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch(instanceSearch, () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchInstances()
  }, 250)
})

onMounted(() => {
  fetchStats()
  fetchApprovals()
  fetchInstances()
  fetchDefinitions()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-git-branch"
            class="w-6 h-6 text-purple-400"
          />
          Workflow Engine & Approval Matrix
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          ITIL Multi-level Approval Chains, Automated Orchestration & Execution Logs (:8085)
        </p>
      </div>

      <button
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white text-xs font-semibold shadow-lg shadow-purple-500/20 hover:scale-105 transition-all"
        @click="isLaunchModalOpen = true"
      >
        <UIcon
          name="i-lucide-play"
          class="w-4 h-4"
        />
        <span>+ Launch Workflow</span>
      </button>
    </div>

    <!-- Live KPI Metrics Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Pending Approvals
        </p>
        <p class="text-2xl font-extrabold text-amber-400 mt-1">
          {{ stats.pending_approvals }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Active Executions
        </p>
        <p class="text-2xl font-extrabold text-indigo-400 mt-1">
          {{ stats.active_instances }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Completed Today
        </p>
        <p class="text-2xl font-extrabold text-emerald-400 mt-1">
          {{ stats.completed_today }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Workflow Blueprints
        </p>
        <p class="text-2xl font-extrabold text-purple-400 mt-1">
          {{ stats.total_definitions }}
        </p>
      </div>
    </div>

    <!-- Tabs Navigation -->
    <div class="flex items-center gap-2 border-b border-slate-800 pb-2">
      <button
        class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
        :class="activeTab === 'approvals' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30' : 'text-slate-400 hover:text-white'"
        @click="activeTab = 'approvals'"
      >
        <UIcon
          name="i-lucide-check-circle"
          class="w-4 h-4"
        />
        <span>My Approvals Queue</span>
        <span
          v-if="stats.pending_approvals > 0"
          class="px-1.5 py-0.2 rounded-full bg-amber-500 text-slate-950 font-bold text-[10px]"
        >
          {{ stats.pending_approvals }}
        </span>
      </button>

      <button
        class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
        :class="activeTab === 'instances' ? 'bg-purple-600/20 text-purple-300 border border-purple-500/30' : 'text-slate-400 hover:text-white'"
        @click="activeTab = 'instances'"
      >
        <UIcon
          name="i-lucide-activity"
          class="w-4 h-4"
        />
        <span>Live Workflow Executions</span>
      </button>

      <button
        class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
        :class="activeTab === 'definitions' ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30' : 'text-slate-400 hover:text-white'"
        @click="activeTab = 'definitions'"
      >
        <UIcon
          name="i-lucide-layers"
          class="w-4 h-4"
        />
        <span>Workflow Blueprints (DAG)</span>
      </button>
    </div>

    <!-- TAB 1: APPROVALS QUEUE -->
    <div
      v-if="activeTab === 'approvals'"
      class="space-y-4"
    >
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <span class="text-xs text-slate-400 font-semibold">Filter:</span>
          <select
            v-model="approvalStatusFilter"
            class="px-3 py-1.5 rounded-xl bg-slate-950 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-purple-500"
          >
            <option value="PENDING">
              Pending My Review
            </option>
            <option value="APPROVED">
              Approved
            </option>
            <option value="REJECTED">
              Rejected
            </option>
            <option value="All">
              All History
            </option>
          </select>
        </div>

        <button
          class="p-2 rounded-xl bg-slate-800/60 hover:bg-slate-800 text-slate-300 transition-colors"
          @click="fetchApprovals"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
          />
        </button>
      </div>

      <div
        v-if="approvals.length === 0"
        class="p-12 text-center text-slate-500 rounded-2xl bg-slate-900/60 border border-slate-800"
      >
        No pending approval requests in your queue.
      </div>

      <div
        v-else
        class="space-y-3"
      >
        <div
          v-for="app in approvals"
          :key="app.id"
          class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:border-purple-500/30 transition-colors"
        >
          <div class="space-y-1.5">
            <div class="flex items-center gap-2">
              <span class="font-mono text-[11px] font-bold text-amber-400 px-2 py-0.5 rounded bg-amber-500/10 border border-amber-500/20">
                Level {{ app.approval_level }}
              </span>
              <h3 class="text-sm font-bold text-white">
                {{ app.title }}
              </h3>
            </div>
            <p class="text-xs text-slate-400">
              Assigned Approver: <span class="text-purple-300 font-medium">{{ app.approver_name }}</span> ({{ app.approver_role }}) · SLA Deadline: <span class="font-mono text-slate-300">{{ new Date(app.sla_deadline).toLocaleString() }}</span>
            </p>
            <p
              v-if="app.decision_notes"
              class="text-xs text-slate-400 italic"
            >
              "{{ app.decision_notes }}"
            </p>
          </div>

          <div class="flex items-center gap-2">
            <template v-if="app.status === 'PENDING'">
              <button
                class="px-3.5 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-md shadow-emerald-500/20 transition-all flex items-center gap-1.5"
                @click="openDecisionModal(app, 'APPROVED')"
              >
                <UIcon
                  name="i-lucide-check"
                  class="w-3.5 h-3.5"
                />
                <span>Approve</span>
              </button>
              <button
                class="px-3.5 py-1.5 rounded-xl bg-rose-600/20 hover:bg-rose-600/40 text-rose-300 border border-rose-500/30 text-xs font-semibold transition-all flex items-center gap-1.5"
                @click="openDecisionModal(app, 'REJECTED')"
              >
                <UIcon
                  name="i-lucide-x"
                  class="w-3.5 h-3.5"
                />
                <span>Reject</span>
              </button>
            </template>

            <span
              v-else
              class="px-3 py-1 rounded-full text-xs font-semibold border"
              :class="app.status === 'APPROVED' ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30' : 'bg-rose-500/15 text-rose-300 border-rose-500/30'"
            >
              {{ app.status }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: LIVE WORKFLOW EXECUTIONS -->
    <div
      v-else-if="activeTab === 'instances'"
      class="space-y-4"
    >
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex items-center justify-between gap-3">
        <div class="relative w-full sm:w-80">
          <UIcon
            name="i-lucide-search"
            class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
          />
          <input
            v-model="instanceSearch"
            type="text"
            placeholder="Search execution instances..."
            class="w-full pl-9 pr-4 py-2 text-xs rounded-xl bg-slate-950 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-purple-500"
          >
        </div>
        <button
          class="p-2 rounded-xl bg-slate-800/60 hover:bg-slate-800 text-slate-300"
          @click="fetchInstances"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
          />
        </button>
      </div>

      <div class="space-y-3">
        <div
          v-for="inst in instances"
          :key="inst.id"
          class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4 hover:border-purple-500/30 transition-colors cursor-pointer"
          @click="openInstanceDetail(inst)"
        >
          <div class="space-y-1.5">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs font-bold text-purple-400 px-2 py-0.5 rounded bg-purple-500/10 border border-purple-500/20">
                {{ inst.instance_number }}
              </span>
              <h3 class="text-sm font-bold text-white">
                {{ inst.title }}
              </h3>
            </div>
            <p class="text-xs text-slate-400">
              Blueprint: <span class="text-indigo-300">{{ inst.definition_name }}</span> · Requester: <span class="text-white">{{ inst.requester_name }}</span> · Current Stage: <span class="text-amber-300 font-semibold">{{ inst.current_step_name }}</span>
            </p>
          </div>

          <div class="flex items-center gap-3">
            <span
              class="px-2.5 py-1 rounded-full text-xs font-semibold border"
              :class="inst.status === 'COMPLETED' ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30' : inst.status === 'WAITING_APPROVAL' ? 'bg-amber-500/15 text-amber-300 border-amber-500/30' : inst.status === 'REJECTED' ? 'bg-rose-500/15 text-rose-300 border-rose-500/30' : 'bg-blue-500/15 text-blue-300 border-blue-500/30'"
            >
              {{ inst.status }}
            </span>
            <button class="px-3 py-1.5 rounded-xl bg-purple-600/20 hover:bg-purple-600/40 text-purple-300 border border-purple-500/30 text-xs font-semibold transition-all">
              View Audit &rarr;
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 3: BLUEPRINTS -->
    <div
      v-else-if="activeTab === 'definitions'"
      class="grid grid-cols-1 sm:grid-cols-3 gap-4"
    >
      <div
        v-for="def in definitions"
        :key="def.id"
        class="p-5 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3 hover:border-purple-500/40 transition-colors flex flex-col justify-between"
      >
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <span class="font-mono text-xs font-bold text-purple-400 px-2 py-0.5 rounded bg-purple-500/10 border border-purple-500/20">
              {{ def.code }}
            </span>
            <span class="text-[10px] px-2 py-0.5 rounded-full font-bold bg-indigo-500/20 text-indigo-300">
              {{ def.category }}
            </span>
          </div>
          <h3 class="text-sm font-bold text-white">
            {{ def.name }}
          </h3>
          <p class="text-xs text-slate-400 leading-relaxed">
            {{ def.description }}
          </p>
        </div>

        <button
          class="w-full py-2 rounded-xl bg-purple-600/20 hover:bg-purple-600 text-purple-200 hover:text-white text-xs font-semibold border border-purple-500/30 transition-all flex items-center justify-center gap-1.5"
          @click="newInstance.definition_id = def.id; newInstance.title = `Request for ${def.name}`; isLaunchModalOpen = true"
        >
          <UIcon
            name="i-lucide-play"
            class="w-3.5 h-3.5"
          />
          <span>Launch Instance</span>
        </button>
      </div>
    </div>

    <!-- Decision Modal -->
    <div
      v-if="isDecisionModalOpen && selectedApproval"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-md rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <h3 class="text-base font-bold text-white flex items-center gap-2">
          <UIcon
            :name="decisionType === 'APPROVED' ? 'i-lucide-check-circle' : 'i-lucide-x-circle'"
            :class="decisionType === 'APPROVED' ? 'text-emerald-400' : 'text-rose-400'"
            class="w-5 h-5"
          />
          Confirm Decision: {{ decisionType }}
        </h3>
        <p class="text-slate-300 font-semibold">
          {{ selectedApproval.title }}
        </p>

        <div class="space-y-1">
          <label class="font-semibold text-slate-300">Decision Notes & Rationale *</label>
          <textarea
            v-model="decisionNotes"
            rows="3"
            required
            class="w-full p-3 rounded-xl bg-slate-950 border border-slate-800 text-white placeholder-slate-500 focus:outline-none focus:border-purple-500"
          />
        </div>

        <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-800">
          <button
            type="button"
            class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300"
            @click="isDecisionModalOpen = false"
          >
            Cancel
          </button>
          <button
            type="button"
            :disabled="processingDecision"
            class="px-5 py-2 rounded-xl text-white font-semibold flex items-center gap-1.5"
            :class="decisionType === 'APPROVED' ? 'bg-emerald-600 hover:bg-emerald-500' : 'bg-rose-600 hover:bg-rose-500'"
            @click="handleSubmitDecision"
          >
            <UIcon
              v-if="processingDecision"
              name="i-lucide-loader-2"
              class="w-4 h-4 animate-spin"
            />
            <span>{{ processingDecision ? 'Recording...' : `Confirm ${decisionType}` }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Launch Instance Modal -->
    <div
      v-if="isLaunchModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-play"
              class="w-5 h-5 text-purple-400"
            />
            Launch Workflow Process Instance
          </h3>
          <button
            class="text-slate-400 hover:text-white"
            @click="isLaunchModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div
          v-if="launchError"
          class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300"
        >
          {{ launchError }}
        </div>

        <form
          class="space-y-3"
          @submit.prevent="handleLaunchWorkflow"
        >
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Workflow Blueprint *</label>
            <select
              v-model="newInstance.definition_id"
              required
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-purple-500"
            >
              <option value="">
                Select Process Blueprint...
              </option>
              <option
                v-for="d in definitions"
                :key="d.id"
                :value="d.id"
              >
                [{{ d.category }}] {{ d.name }}
              </option>
            </select>
          </div>

          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Instance Title / Purpose *</label>
            <input
              v-model="newInstance.title"
              type="text"
              required
              placeholder="e.g. Provisioning workstation for Emily Davis"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-purple-500"
            >
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Entity Type</label>
              <select
                v-model="newInstance.entity_type"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white"
              >
                <option value="SERVICE_REQUEST">
                  Service Request
                </option>
                <option value="TICKET">
                  Incident Ticket
                </option>
                <option value="ASSET">
                  Asset Order
                </option>
                <option value="CHANGE">
                  Change Request (CAB)
                </option>
              </select>
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Reference ID</label>
              <input
                v-model="newInstance.entity_id"
                type="text"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono"
              >
            </div>
          </div>

          <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300"
              @click="isLaunchModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="launching"
              class="px-5 py-2 rounded-xl bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 text-white font-semibold"
            >
              {{ launching ? 'Launching...' : 'Start Execution' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Instance Detail & Audit Log Modal -->
    <div
      v-if="isInstanceDetailOpen && selectedInstance"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-2xl rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <div>
            <div class="flex items-center gap-2">
              <span class="font-mono font-bold text-purple-400">{{ selectedInstance.instance_number }}</span>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-purple-500/20 text-purple-300">{{ selectedInstance.status }}</span>
            </div>
            <h3 class="text-sm font-bold text-white mt-1">
              {{ selectedInstance.title }}
            </h3>
          </div>
          <button
            class="text-slate-400 hover:text-white"
            @click="isInstanceDetailOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="space-y-3">
          <h4 class="font-bold text-slate-300 uppercase tracking-wider text-[11px]">
            Execution Timeline Audit Trail
          </h4>
          <div class="space-y-2 max-h-64 overflow-y-auto">
            <div
              v-for="log in instanceLogs"
              :key="log.id"
              class="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-1"
            >
              <div class="flex items-center justify-between font-semibold">
                <span class="text-purple-300">{{ log.action }}</span>
                <span class="text-slate-500 text-[10px]">{{ new Date(log.created_at).toLocaleTimeString() }}</span>
              </div>
              <p class="text-slate-300 text-[11px]">
                {{ log.message }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
