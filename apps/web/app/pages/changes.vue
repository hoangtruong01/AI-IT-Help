<script setup lang="ts">
import type {
  ChangeRequest,
  CABReview,
  ChangeStats,
  ChangeCalendarItem,
  CreateChangePayload,
  UpdateChangeStatusPayload,
  SubmitCABVotePayload,
  PaginatedResponse
} from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()
const authStore = useAuthStore()
const toast = useToast()

// Active View
const activeView = ref<'table' | 'risk_matrix' | 'calendar'>('table')

// State
const changes = ref<ChangeRequest[]>([])
const calendarItems = ref<ChangeCalendarItem[]>([])
const stats = ref<ChangeStats>({
  active_changes: 0,
  pending_cab_review: 0,
  emergency_changes: 0,
  success_rate_percent: 0,
  total_this_month: 0
})

const loading = ref(false)
const searchQuery = ref('')
const selectedType = ref('All')
const selectedStatus = ref('All')
const selectedRisk = ref('All')

// Detail / Review Drawer State
const isDetailOpen = ref(false)
const selectedChange = ref<ChangeRequest | null>(null)
const cabReviews = ref<CABReview[]>([])
const isUpdatingStatus = ref(false)

// CAB Vote Modal State
const isVoteModalOpen = ref(false)
const voteDecision = ref<'APPROVED' | 'REJECTED' | 'ABSTAIN'>('APPROVED')
const voteComments = ref('')
const isSubmittingVote = ref(false)

// Create RFC Modal State
const isCreateModalOpen = ref(false)
const isCreating = ref(false)
const newChange = reactive<CreateChangePayload>({
  title: '',
  description: '',
  change_type: 'NORMAL',
  category: 'DevOps & Infrastructure',
  priority: 'MEDIUM',
  impact_level: 'MEDIUM',
  probability_level: 'MEDIUM',
  requester_id: '',
  requester_name: '',
  requester_email: '',
  reason_for_change: '',
  implementation_plan: '',
  rollback_plan: '',
  test_plan: '',
  downtime_required: false,
  downtime_minutes: 0
})

const categories = [
  'DevOps & Infrastructure',
  'Network & Access',
  'Database & Storage',
  'IT Security & Access',
  'Core ERP & Helpdesk'
]

// Fetch Changes & Calendar
async function fetchData() {
  loading.value = true
  try {
    const [cRes, sRes, calRes] = await Promise.all([
      api.get<PaginatedResponse<ChangeRequest>>('/api/v1/changes', {
        params: {
          type: selectedType.value === 'All' ? undefined : selectedType.value,
          status: selectedStatus.value === 'All' ? undefined : selectedStatus.value,
          risk: selectedRisk.value === 'All' ? undefined : selectedRisk.value,
          page_size: 50
        }
      }).catch(() => null),
      api.get<ChangeStats>('/api/v1/changes/stats').catch(() => null),
      api.get<ChangeCalendarItem[]>('/api/v1/changes/calendar').catch(() => null)
    ])

    changes.value = cRes?.data ?? []

    stats.value = sRes ?? {
      active_changes: 0,
      pending_cab_review: 0,
      emergency_changes: 0,
      success_rate_percent: 0,
      total_this_month: 0
    }
    calendarItems.value = calRes ?? []

    if (!cRes || !sRes || !calRes) {
      toast.add({ title: 'Some change-management data could not be loaded', color: 'warning' })
    }
  } finally {
    loading.value = false
  }
}

// Open Change Drawer
async function openChangeDetail(c: ChangeRequest) {
  selectedChange.value = c
  isDetailOpen.value = true

  try {
    const res = await api.get<{ change: ChangeRequest, cab_reviews: CABReview[] }>(`/api/v1/changes/${c.id}`)
    cabReviews.value = res?.cab_reviews ?? []
  } catch {
    cabReviews.value = []
  }
}

// Submit CAB Vote
async function handleSubmitCABVote() {
  if (!selectedChange.value) return
  if (!authStore.user) {
    toast.add({ title: 'An authenticated user profile is required', color: 'error' })
    return
  }
  isSubmittingVote.value = true
  try {
    const payload: SubmitCABVotePayload = {
      reviewer_id: authStore.user.id,
      reviewer_name: authStore.user.full_name,
      reviewer_role: authStore.user.role,
      vote: voteDecision.value,
      comments: voteComments.value || undefined
    }

    const res = await api.post<{ review: CABReview, change: ChangeRequest }>(`/api/v1/changes/${selectedChange.value.id}/cab-vote`, payload)
    if (res && res.change) {
      selectedChange.value = res.change
    }
    toast.add({
      title: 'CAB Vote Recorded',
      description: `Your ${voteDecision.value} vote has been officially logged in the CAB audit register.`,
      color: 'success'
    })
    isVoteModalOpen.value = false
    voteComments.value = ''
    openChangeDetail(selectedChange.value)
    fetchData()
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Vote Failed', description: errObj?.message || 'Failed to record CAB vote', color: 'error' })
  } finally {
    isSubmittingVote.value = false
  }
}

// Update Change Lifecycle Status (Test Case 7.2 Quorum Check)
async function handleUpdateStatus(targetStatus: string) {
  if (!selectedChange.value) return
  isUpdatingStatus.value = true
  try {
    const payload: UpdateChangeStatusPayload = {
      status: targetStatus,
      notes: `Status transitioned to ${targetStatus}`,
      version: selectedChange.value.version ?? 1
    }

    const updated = await api.patch<ChangeRequest>(`/api/v1/changes/${selectedChange.value.id}/status`, payload)
    selectedChange.value = updated
    toast.add({
      title: 'Status Updated',
      description: `Change Request ${updated.change_number} transitioned to ${targetStatus}.`,
      color: 'success'
    })
    fetchData()
  } catch (err: unknown) {
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    const errMsg = errObj?.data?.error?.message || errObj?.message || 'Status transition rejected.'
    toast.add({
      title: 'Action Blocked (CAB Enforcement)',
      description: errMsg,
      color: 'error'
    })
  } finally {
    isUpdatingStatus.value = false
  }
}

// Create RFC
async function handleCreateChange() {
  if (!newChange.title.trim() || !newChange.reason_for_change.trim()) {
    toast.add({ title: 'Validation Error', description: 'Title and Reason for Change are mandatory', color: 'error' })
    return
  }

  isCreating.value = true
  try {
    newChange.requester_id = authStore.user?.id || 'u-req'
    newChange.requester_name = authStore.user?.full_name || 'System Operator'
    newChange.requester_email = authStore.user?.email || 'operator@eomp.local'

    await api.post('/api/v1/changes', newChange)
    toast.add({
      title: 'RFC Created',
      description: 'Request for Change submitted to Change Advisory Board queue.',
      color: 'success'
    })
    isCreateModalOpen.value = false
    fetchData()
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Error', description: errObj?.message || 'Failed to create RFC', color: 'error' })
  } finally {
    isCreating.value = false
  }
}

// Filtered Changes
const filteredChanges = computed(() => {
  return changes.value.filter((c) => {
    const q = searchQuery.value.toLowerCase()
    const matchesQuery = !q || c.change_number.toLowerCase().includes(q) || c.title.toLowerCase().includes(q) || c.description.toLowerCase().includes(q)
    const matchesType = selectedType.value === 'All' || c.change_type === selectedType.value
    const matchesStatus = selectedStatus.value === 'All' || c.status === selectedStatus.value
    const matchesRisk = selectedRisk.value === 'All' || c.risk_level === selectedRisk.value
    return matchesQuery && matchesType && matchesStatus && matchesRisk
  })
})

// Risk Matrix Grid Computations
function getRiskGridChanges(probability: string, impact: string): ChangeRequest[] {
  return changes.value.filter(c => c.probability_level === probability && c.impact_level === impact)
}

onMounted(() => {
  fetchData()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-git-pull-request"
            class="w-7 h-7 text-indigo-400"
          />
          Change Advisory Board (CAB) & RFC Lifecycle
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          ITIL v4 Change Management, 3x3 Risk Assessment Matrix, Maintenance Window Scheduling & CAB Multi-Reviewer Quorum
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <button
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 text-xs font-semibold shadow-md transition-all"
          @click="fetchData"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
          />
          <span>Refresh</span>
        </button>

        <button
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-indigo-600 to-cyan-600 hover:from-indigo-500 hover:to-cyan-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 hover:scale-105 transition-all"
          @click="isCreateModalOpen = true"
        >
          <UIcon
            name="i-lucide-plus-circle"
            class="w-4 h-4"
          />
          <span>+ New RFC Change</span>
        </button>
      </div>
    </div>

    <!-- 4 KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-indigo-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Active RFCs</span>
          <UIcon
            name="i-lucide-git-commit"
            class="w-5 h-5 text-indigo-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.active_changes }}
        </p>
        <span class="text-[10px] text-indigo-400 mt-1 block">Scheduled / Implementing</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-amber-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Pending CAB Review</span>
          <UIcon
            name="i-lucide-users"
            class="w-5 h-5 text-amber-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.pending_cab_review }}
        </p>
        <span class="text-[10px] text-amber-400 mt-1 block">Awaiting Quorum Votes</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-rose-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Emergency RFCs</span>
          <UIcon
            name="i-lucide-flame"
            class="w-5 h-5 text-rose-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.emergency_changes }}
        </p>
        <span class="text-[10px] text-rose-400 mt-1 block">2-Signature Authorization</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-emerald-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Change Success Rate</span>
          <UIcon
            name="i-lucide-shield-check"
            class="w-5 h-5 text-emerald-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.success_rate_percent }}%
        </p>
        <span class="text-[10px] text-emerald-400 mt-1 block">Zero Unplanned Outages</span>
      </div>
    </div>

    <!-- View Switcher Tabs -->
    <div class="flex items-center justify-between gap-4 border-b border-slate-800 pb-3">
      <div class="flex items-center gap-1 p-1 bg-slate-900/80 border border-slate-800 rounded-xl w-fit">
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeView === 'table' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeView = 'table'"
        >
          <UIcon
            name="i-lucide-list"
            class="w-4 h-4"
          />
          <span>Danh Sách RFCs (Table)</span>
        </button>
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeView === 'risk_matrix' ? 'bg-rose-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeView = 'risk_matrix'"
        >
          <UIcon
            name="i-lucide-grid"
            class="w-4 h-4"
          />
          <span>Ma Trận Rủi Ro (3x3 Risk Matrix)</span>
        </button>
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeView === 'calendar' ? 'bg-cyan-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeView = 'calendar'"
        >
          <UIcon
            name="i-lucide-calendar-range"
            class="w-4 h-4"
          />
          <span>Lịch Bảo Trì (Maintenance Calendar)</span>
        </button>
      </div>
    </div>

    <!-- View 1: RFC Table -->
    <div
      v-if="activeView === 'table'"
      class="space-y-4"
    >
      <!-- Filters -->
      <div class="p-4 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div class="relative flex-1">
          <UIcon
            name="i-lucide-search"
            class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"
          />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search RFCs by CHG ID, title, requester, keywords..."
            class="w-full pl-10 pr-4 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white placeholder-slate-400 text-xs focus:outline-none focus:border-indigo-500"
          >
        </div>

        <div class="flex items-center gap-2">
          <select
            v-model="selectedType"
            class="px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-slate-300 text-xs focus:outline-none focus:border-indigo-500"
          >
            <option value="All">
              Type: All
            </option>
            <option value="EMERGENCY">
              EMERGENCY
            </option>
            <option value="MAJOR">
              MAJOR
            </option>
            <option value="NORMAL">
              NORMAL
            </option>
            <option value="STANDARD">
              STANDARD
            </option>
          </select>

          <select
            v-model="selectedRisk"
            class="px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-slate-300 text-xs focus:outline-none focus:border-indigo-500"
          >
            <option value="All">
              Risk: All
            </option>
            <option value="CRITICAL">
              CRITICAL
            </option>
            <option value="HIGH">
              HIGH
            </option>
            <option value="MEDIUM">
              MEDIUM
            </option>
            <option value="LOW">
              LOW
            </option>
          </select>
        </div>
      </div>

      <!-- Table -->
      <div class="rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl overflow-hidden shadow-2xl">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs text-slate-300">
            <thead class="bg-slate-950/80 text-slate-400 uppercase font-semibold border-b border-slate-800 text-[10px] tracking-wider">
              <tr>
                <th class="p-4">
                  RFC Number & Title
                </th>
                <th class="p-4">
                  Type / Category
                </th>
                <th class="p-4">
                  Risk Level
                </th>
                <th class="p-4">
                  Status
                </th>
                <th class="p-4">
                  CAB Quorum Progress
                </th>
                <th class="p-4">
                  Maintenance Window
                </th>
                <th class="p-4 text-right">
                  Action
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60">
              <tr
                v-for="c in filteredChanges"
                :key="c.id"
                class="hover:bg-slate-800/40 transition-colors group cursor-pointer"
                @click="openChangeDetail(c)"
              >
                <td class="p-4">
                  <div class="space-y-1">
                    <div class="flex items-center gap-2">
                      <span class="font-mono text-xs font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                        {{ c.change_number }}
                      </span>
                      <span
                        class="text-[9px] font-mono px-1.5 py-0.5 rounded font-bold"
                        :class="{
                          'bg-rose-500/20 text-rose-300 border border-rose-500/30': c.change_type === 'EMERGENCY',
                          'bg-amber-500/20 text-amber-300 border border-amber-500/30': c.change_type === 'MAJOR',
                          'bg-blue-500/20 text-blue-300 border border-blue-500/30': c.change_type === 'NORMAL',
                          'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30': c.change_type === 'STANDARD'
                        }"
                      >
                        {{ c.change_type }}
                      </span>
                    </div>
                    <p class="font-bold text-white group-hover:text-indigo-300 transition-colors max-w-md">
                      {{ c.title }}
                    </p>
                  </div>
                </td>

                <td class="p-4 font-mono text-slate-400">
                  {{ c.category }}
                </td>

                <td class="p-4">
                  <span
                    class="text-[10px] font-bold px-2 py-0.5 rounded border"
                    :class="{
                      'bg-rose-500/20 text-rose-300 border-rose-500/30': c.risk_level === 'CRITICAL',
                      'bg-amber-500/20 text-amber-300 border-amber-500/30': c.risk_level === 'HIGH',
                      'bg-yellow-500/20 text-yellow-300 border-yellow-500/30': c.risk_level === 'MEDIUM',
                      'bg-emerald-500/20 text-emerald-300 border-emerald-500/30': c.risk_level === 'LOW'
                    }"
                  >
                    {{ c.risk_level }}
                  </span>
                </td>

                <td class="p-4">
                  <span
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold border"
                    :class="{
                      'bg-amber-500/10 text-amber-400 border-amber-500/20': c.status === 'CAB_REVIEW',
                      'bg-emerald-500/10 text-emerald-400 border-emerald-500/20': c.status === 'APPROVED' || c.status === 'COMPLETED',
                      'bg-blue-500/10 text-blue-400 border-blue-500/20': c.status === 'SCHEDULED' || c.status === 'IMPLEMENTING',
                      'bg-rose-500/10 text-rose-400 border-rose-500/20': c.status === 'REJECTED' || c.status === 'FAILED'
                    }"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-current" />
                    <span>{{ c.status }}</span>
                  </span>
                </td>

                <td class="p-4">
                  <div class="space-y-1">
                    <div class="flex items-center justify-between text-[10px]">
                      <span class="text-slate-400 font-mono">Quorum: {{ c.cab_approved_count }}/{{ c.cab_required_count }} Votes</span>
                      <span
                        v-if="c.cab_approved_count >= c.cab_required_count"
                        class="text-emerald-400 font-bold"
                      >Ready</span>
                    </div>
                    <div class="w-24 h-1.5 bg-slate-800 rounded-full overflow-hidden">
                      <div
                        class="h-full bg-emerald-500 transition-all"
                        :style="{ width: `${Math.min(100, (c.cab_approved_count / Math.max(1, c.cab_required_count)) * 100)}%` }"
                      />
                    </div>
                  </div>
                </td>

                <td class="p-4 text-[11px] text-slate-300">
                  <div v-if="c.scheduled_start_time">
                    <p class="font-mono text-slate-400">
                      {{ new Date(c.scheduled_start_time).toLocaleDateString() }}
                    </p>
                    <span
                      v-if="c.downtime_required"
                      class="text-[9px] text-rose-400 font-semibold"
                    >Downtime: {{ c.downtime_minutes }}m</span>
                    <span
                      v-else
                      class="text-[9px] text-emerald-400"
                    >Zero Downtime</span>
                  </div>
                  <span
                    v-else
                    class="text-slate-500"
                  >Not Scheduled</span>
                </td>

                <td class="p-4 text-right">
                  <button
                    class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold flex items-center gap-1 ml-auto shadow"
                    @click.stop="openChangeDetail(c)"
                  >
                    <span>CAB Review</span>
                    <UIcon
                      name="i-lucide-arrow-right"
                      class="w-3.5 h-3.5"
                    />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- View 2: 3x3 Risk Matrix Visualizer -->
    <div
      v-if="activeView === 'risk_matrix'"
      class="p-6 rounded-3xl bg-slate-900/70 border border-slate-800 backdrop-blur-xl space-y-6 shadow-2xl"
    >
      <div class="flex items-center justify-between border-b border-slate-800 pb-4">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-grid"
              class="w-5 h-5 text-rose-400"
            />
            ITIL 3x3 Risk Assessment Matrix (Probability vs Impact)
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Evaluates failure probability against business service disruption severity.
          </p>
        </div>
      </div>

      <!-- 3x3 Grid -->
      <div class="grid grid-cols-4 gap-3 text-xs">
        <!-- Header row -->
        <div class="p-3 font-bold text-slate-400 uppercase text-center flex items-center justify-center">
          Probability \ Impact
        </div>
        <div class="p-3 font-bold text-emerald-400 uppercase text-center bg-emerald-950/20 rounded-xl border border-emerald-500/20">
          LOW IMPACT
        </div>
        <div class="p-3 font-bold text-amber-400 uppercase text-center bg-amber-950/20 rounded-xl border border-amber-500/20">
          MEDIUM IMPACT
        </div>
        <div class="p-3 font-bold text-rose-400 uppercase text-center bg-rose-950/20 rounded-xl border border-rose-500/20">
          HIGH / CRITICAL IMPACT
        </div>

        <!-- Row 1: High Probability -->
        <div class="p-4 font-bold text-rose-400 flex items-center justify-center bg-rose-950/10 rounded-xl border border-rose-500/20">
          HIGH PROBABILITY
        </div>
        <div class="p-4 rounded-xl bg-amber-950/30 border border-amber-500/30 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-amber-400 font-bold block">MEDIUM RISK</span>
          <div
            v-for="c in getRiskGridChanges('HIGH', 'LOW')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-amber-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-amber-950/40 border border-amber-500/40 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-amber-400 font-bold block">HIGH RISK</span>
          <div
            v-for="c in getRiskGridChanges('HIGH', 'MEDIUM')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-amber-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-rose-950/50 border border-rose-500/50 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-rose-400 font-bold block">CRITICAL RISK</span>
          <div
            v-for="c in getRiskGridChanges('HIGH', 'HIGH')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-rose-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-rose-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>

        <!-- Row 2: Medium Probability -->
        <div class="p-4 font-bold text-amber-400 flex items-center justify-center bg-amber-950/10 rounded-xl border border-amber-500/20">
          MEDIUM PROBABILITY
        </div>
        <div class="p-4 rounded-xl bg-emerald-950/20 border border-emerald-500/20 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-emerald-400 font-bold block">LOW RISK</span>
          <div
            v-for="c in getRiskGridChanges('MEDIUM', 'LOW')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-emerald-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-amber-950/30 border border-amber-500/30 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-amber-400 font-bold block">MEDIUM RISK</span>
          <div
            v-for="c in getRiskGridChanges('MEDIUM', 'MEDIUM')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-amber-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-amber-950/40 border border-amber-500/40 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-amber-400 font-bold block">HIGH RISK</span>
          <div
            v-for="c in getRiskGridChanges('MEDIUM', 'HIGH')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-amber-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>

        <!-- Row 3: Low Probability -->
        <div class="p-4 font-bold text-emerald-400 flex items-center justify-center bg-emerald-950/10 rounded-xl border border-emerald-500/20">
          LOW PROBABILITY
        </div>
        <div class="p-4 rounded-xl bg-emerald-950/20 border border-emerald-500/20 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-emerald-400 font-bold block">LOW RISK</span>
          <div
            v-for="c in getRiskGridChanges('LOW', 'LOW')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-emerald-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-amber-950/20 border border-amber-500/20 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-amber-400 font-bold block">MEDIUM RISK</span>
          <div
            v-for="c in getRiskGridChanges('LOW', 'MEDIUM')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-amber-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-indigo-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
        <div class="p-4 rounded-xl bg-rose-950/40 border border-rose-500/40 min-h-[110px] space-y-2">
          <span class="text-[10px] font-mono text-rose-400 font-bold block">CRITICAL RISK</span>
          <div
            v-for="c in getRiskGridChanges('LOW', 'CRITICAL')"
            :key="c.id"
            class="p-2 rounded-lg bg-slate-900 border border-slate-800 text-[11px] cursor-pointer hover:border-rose-500/50"
            @click="openChangeDetail(c)"
          >
            <span class="font-mono text-rose-300 font-bold">{{ c.change_number }}</span>
            <p class="truncate text-slate-200">
              {{ c.title }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- View 3: Maintenance Calendar & Timeline -->
    <div
      v-if="activeView === 'calendar'"
      class="p-6 rounded-3xl bg-slate-900/70 border border-slate-800 backdrop-blur-xl space-y-5 shadow-2xl"
    >
      <div class="flex items-center justify-between border-b border-slate-800 pb-3">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-calendar-range"
              class="w-5 h-5 text-cyan-400"
            />
            Maintenance Windows & Downtime Schedule
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Avoid concurrent downtime conflicts across interdependent microservices and network edges.
          </p>
        </div>
      </div>

      <div class="space-y-3">
        <div
          v-for="cal in calendarItems"
          :key="cal.id"
          class="p-4 rounded-2xl bg-slate-950 border border-slate-800 flex items-center justify-between gap-4 hover:border-cyan-500/30 transition-all cursor-pointer"
        >
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                {{ cal.change_number }}
              </span>
              <span
                class="text-[10px] font-bold px-2 py-0.5 rounded border"
                :class="cal.risk_level === 'CRITICAL' || cal.risk_level === 'HIGH' ? 'bg-rose-500/15 text-rose-300 border-rose-500/30' : 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'"
              >
                {{ cal.risk_level }} RISK
              </span>
              <span class="text-[10px] text-slate-400 font-mono">{{ cal.category }}</span>
            </div>
            <h4 class="text-sm font-bold text-white">
              {{ cal.title }}
            </h4>
          </div>

          <div class="text-right shrink-0 space-y-1">
            <span class="text-xs font-mono text-cyan-300 block">
              {{ cal.scheduled_start ? new Date(cal.scheduled_start).toLocaleString() : 'Unscheduled' }}
            </span>
            <span
              v-if="cal.downtime_required"
              class="text-[10px] font-bold text-rose-400 px-2 py-0.5 rounded bg-rose-500/10 border border-rose-500/20"
            >
              Planned Outage: {{ cal.downtime_minutes }}m
            </span>
            <span
              v-else
              class="text-[10px] font-bold text-emerald-400 px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20"
            >
              Zero Downtime
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- CAB Review & Detail Drawer Modal -->
    <div
      v-if="isDetailOpen && selectedChange"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-4xl max-h-[90vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-5 text-white shadow-2xl">
        <!-- Drawer Header -->
        <div class="flex items-start justify-between gap-4 border-b border-slate-800 pb-4">
          <div class="space-y-1.5">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs font-bold text-indigo-400 px-2.5 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                {{ selectedChange.change_number }}
              </span>
              <span
                class="text-[10px] font-bold px-2 py-0.5 rounded border"
                :class="selectedChange.change_type === 'EMERGENCY' ? 'bg-rose-500/20 text-rose-300 border-rose-500/30' : 'bg-blue-500/20 text-blue-300 border-blue-500/30'"
              >
                {{ selectedChange.change_type }}
              </span>
              <span class="text-[10px] font-bold px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                {{ selectedChange.category }}
              </span>
            </div>
            <h2 class="text-lg font-bold text-white">
              {{ selectedChange.title }}
            </h2>
          </div>

          <button
            class="p-2 text-slate-400 hover:text-white rounded-lg bg-slate-800"
            @click="isDetailOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <!-- CAB Quorum Card & Action Buttons -->
        <div class="p-5 rounded-2xl bg-indigo-950/30 border border-indigo-500/30 space-y-4">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <span class="text-xs font-bold text-indigo-300 uppercase tracking-wider flex items-center gap-2">
                <UIcon
                  name="i-lucide-users"
                  class="w-4 h-4 text-cyan-400"
                />
                Change Advisory Board (CAB) Quorum Status
              </span>
              <p class="text-[11px] text-slate-300 mt-0.5">
                Requires <strong>{{ selectedChange.cab_required_count }}</strong> verified CAB approvals. Current approvals: <strong>{{ selectedChange.cab_approved_count }}/{{ selectedChange.cab_required_count }}</strong>.
              </p>
            </div>

            <div class="flex items-center gap-2">
              <button
                class="px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-semibold shadow flex items-center gap-1.5 transition-all"
                @click="isVoteModalOpen = true"
              >
                <UIcon
                  name="i-lucide-check-circle"
                  class="w-4 h-4"
                />
                <span>Submit CAB Vote</span>
              </button>
            </div>
          </div>

          <!-- Existing CAB Reviews -->
          <div
            v-if="cabReviews.length > 0"
            class="space-y-2 pt-2 border-t border-indigo-500/20 text-xs"
          >
            <span class="text-[11px] font-semibold text-slate-300">CAB Member Voting Records:</span>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <div
                v-for="r in cabReviews"
                :key="r.id"
                class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-1"
              >
                <div class="flex items-center justify-between">
                  <span class="font-bold text-white">{{ r.reviewer_name }}</span>
                  <span
                    class="text-[9px] font-bold px-2 py-0.5 rounded"
                    :class="r.vote === 'APPROVED' ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-rose-500/20 text-rose-400 border border-rose-500/30'"
                  >
                    {{ r.vote }}
                  </span>
                </div>
                <p
                  v-if="r.comments"
                  class="text-slate-400 italic text-[11px]"
                >
                  "{{ r.comments }}"
                </p>
                <span class="text-[9px] text-slate-500 block font-mono">{{ new Date(r.reviewed_at).toLocaleString() }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Lifecycle Transition Actions (Enforces Test Case 7.2) -->
        <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 flex flex-wrap items-center justify-between gap-3 text-xs">
          <div>
            <span class="font-semibold text-slate-300">Lifecycle State: </span>
            <span class="font-bold text-indigo-400 ml-1">{{ selectedChange.status }}</span>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              v-if="selectedChange.status === 'CAB_REVIEW' && selectedChange.cab_approved_count >= selectedChange.cab_required_count"
              :disabled="isUpdatingStatus"
              class="px-3 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-semibold transition-all"
              @click="handleUpdateStatus('SCHEDULED')"
            >
              📅 Schedule Deployment
            </button>
            <button
              v-if="selectedChange.status !== 'IMPLEMENTING' && selectedChange.status !== 'COMPLETED'"
              :disabled="isUpdatingStatus"
              class="px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-all shadow"
              @click="handleUpdateStatus('IMPLEMENTING')"
            >
              🚀 Start Implementation
            </button>
            <button
              v-if="selectedChange.status === 'IMPLEMENTING'"
              :disabled="isUpdatingStatus"
              class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold transition-all"
              @click="handleUpdateStatus('COMPLETED')"
            >
              ✅ Mark Completed
            </button>
            <button
              v-if="selectedChange.status === 'IMPLEMENTING'"
              :disabled="isUpdatingStatus"
              class="px-3 py-1.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white font-semibold transition-all"
              @click="handleUpdateStatus('FAILED')"
            >
              ❌ Mark Failed / Rollback
            </button>
          </div>
        </div>

        <!-- Plan Specifications Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-xs">
          <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-2">
            <span class="font-bold text-emerald-400 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-play"
                class="w-4 h-4"
              />
              Implementation Plan:
            </span>
            <div class="p-3 rounded-xl bg-slate-900 border border-slate-800/80 font-mono text-[11px] text-slate-300 whitespace-pre-line leading-relaxed">
              {{ selectedChange.implementation_plan }}
            </div>
          </div>

          <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-2">
            <span class="font-bold text-rose-400 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-undo"
                class="w-4 h-4"
              />
              Rollback Plan:
            </span>
            <div class="p-3 rounded-xl bg-slate-900 border border-slate-800/80 font-mono text-[11px] text-slate-300 whitespace-pre-line leading-relaxed">
              {{ selectedChange.rollback_plan }}
            </div>
          </div>

          <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-2">
            <span class="font-bold text-cyan-400 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-check-circle-2"
                class="w-4 h-4"
              />
              Test & Verification Plan:
            </span>
            <div class="p-3 rounded-xl bg-slate-900 border border-slate-800/80 font-mono text-[11px] text-slate-300 whitespace-pre-line leading-relaxed">
              {{ selectedChange.test_plan }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- CAB Vote Modal Overlay -->
    <div
      v-if="isVoteModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-md p-5 bg-slate-900 border border-slate-800 rounded-3xl space-y-4 text-white shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="font-bold text-sm flex items-center gap-2">
            <UIcon
              name="i-lucide-vote"
              class="w-4 h-4 text-emerald-400"
            />
            Submit Official CAB Vote
          </h3>
          <button
            class="text-slate-400 hover:text-white"
            @click="isVoteModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-4 h-4"
            />
          </button>
        </div>

        <div class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-semibold mb-1">Your Decision *</label>
            <div class="grid grid-cols-2 gap-2">
              <button
                type="button"
                class="p-2.5 rounded-xl border text-xs font-bold transition-all"
                :class="voteDecision === 'APPROVED' ? 'bg-emerald-600 text-white border-emerald-500' : 'bg-slate-950 text-slate-400 border-slate-800'"
                @click="voteDecision = 'APPROVED'"
              >
                👍 Approve RFC
              </button>
              <button
                type="button"
                class="p-2.5 rounded-xl border text-xs font-bold transition-all"
                :class="voteDecision === 'REJECTED' ? 'bg-rose-600 text-white border-rose-500' : 'bg-slate-950 text-slate-400 border-slate-800'"
                @click="voteDecision = 'REJECTED'"
              >
                👎 Reject RFC
              </button>
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Reviewer Justification / Feedback</label>
            <textarea
              v-model="voteComments"
              rows="3"
              placeholder="e.g. Risk mitigation verified. Zero downtime window approved..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-2 border-t border-slate-800">
          <button
            class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-300 text-xs font-semibold"
            @click="isVoteModalOpen = false"
          >
            Cancel
          </button>
          <button
            :disabled="isSubmittingVote"
            class="px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold disabled:opacity-50"
            @click="handleSubmitCABVote"
          >
            {{ isSubmittingVote ? 'Submitting...' : 'Confirm Vote' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create RFC Modal Overlay -->
    <div
      v-if="isCreateModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-2xl max-h-[85vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-4 text-white shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h2 class="text-base font-bold flex items-center gap-2 text-white">
            <UIcon
              name="i-lucide-git-pull-request"
              class="w-5 h-5 text-indigo-400"
            />
            Submit Request for Change (RFC)
          </h2>
          <button
            class="text-slate-400 hover:text-white"
            @click="isCreateModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-semibold mb-1">Change Title *</label>
            <input
              v-model="newChange.title"
              type="text"
              placeholder="e.g. Kubernetes Ingress Gateway Upgrade to Traefik v3.1"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
            >
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Change Type *</label>
              <select
                v-model="newChange.change_type"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="NORMAL">
                  NORMAL (1 CAB Signature)
                </option>
                <option value="MAJOR">
                  MAJOR (2 CAB Signatures)
                </option>
                <option value="EMERGENCY">
                  EMERGENCY (2 CAB Signatures)
                </option>
                <option value="STANDARD">
                  STANDARD (Pre-Approved)
                </option>
              </select>
            </div>

            <div>
              <label class="block text-slate-300 font-semibold mb-1">Category *</label>
              <select
                v-model="newChange.category"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
                <option
                  v-for="c in categories"
                  :key="c"
                  :value="c"
                >
                  {{ c }}
                </option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Service Impact Level *</label>
              <select
                v-model="newChange.impact_level"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="LOW">
                  LOW (Single microservice non-critical)
                </option>
                <option value="MEDIUM">
                  MEDIUM (Internal developer tools)
                </option>
                <option value="HIGH">
                  HIGH (Customer-facing services)
                </option>
                <option value="CRITICAL">
                  CRITICAL (Core Gateway / Auth / DB)
                </option>
              </select>
            </div>

            <div>
              <label class="block text-slate-300 font-semibold mb-1">Failure Probability *</label>
              <select
                v-model="newChange.probability_level"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="LOW">
                  LOW (Tested on Staging)
                </option>
                <option value="MEDIUM">
                  MEDIUM (Config changes)
                </option>
                <option value="HIGH">
                  HIGH (Complex schema migration)
                </option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Reason for Change & Business Justification *</label>
            <textarea
              v-model="newChange.reason_for_change"
              rows="2"
              placeholder="Why is this change necessary?..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Implementation Step-by-Step Plan *</label>
            <textarea
              v-model="newChange.implementation_plan"
              rows="3"
              placeholder="1. Step 1... 2. Step 2..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Rollback Plan *</label>
              <textarea
                v-model="newChange.rollback_plan"
                rows="2"
                placeholder="Steps to revert..."
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Test & Verification Plan *</label>
              <textarea
                v-model="newChange.test_plan"
                rows="2"
                placeholder="Health check endpoints..."
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-3 border-t border-slate-800">
          <button
            class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 text-xs font-semibold"
            @click="isCreateModalOpen = false"
          >
            Cancel
          </button>
          <button
            :disabled="isCreating"
            class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 disabled:opacity-50"
            @click="handleCreateChange"
          >
            {{ isCreating ? 'Submitting RFC...' : 'Submit to CAB Queue' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
