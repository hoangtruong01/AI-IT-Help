<script setup lang="ts">
import type { Ticket, ServiceCatalogItem, PaginatedResponse, CreateTicketPayload, TicketComment, TicketTimeline, AITicketAnalysis } from '~/types'
import { classifyApiError, dataViewState, type ApiViewState } from '~/utils/api-view-state'

definePageMeta({ layout: 'default' })

const api = useApi()
const authStore = useAuthStore()
const toast = useToast()

// State
const tickets = ref<Ticket[]>([])
const serviceItems = ref<ServiceCatalogItem[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const pageState = ref<ApiViewState>('loading')

// Filters
const searchQuery = ref('')
const selectedPriority = ref('All')
const selectedStatus = ref('All')

// Create Modal State
const isCreateModalOpen = ref(false)
const creating = ref(false)
const createError = ref<string | null>(null)
const newTicket = reactive<CreateTicketPayload>({
  title: '',
  description: '',
  service_item_id: '',
  category: 'Network & Access',
  priority: 'MEDIUM',
  requester_id: '',
  requester_name: '',
  requester_email: ''
})

// Details Modal State
const isDetailModalOpen = ref(false)
const selectedTicket = ref<Ticket | null>(null)
const comments = ref<TicketComment[]>([])
const timeline = ref<TicketTimeline[]>([])
const newCommentText = ref('')
const isInternalNote = ref(false)
const actionLoading = ref(false)
const actionError = ref<string | null>(null)

// AI Operations Copilot Widget State
const isAnalyzingAI = ref(false)
const aiSuggestion = ref<AITicketAnalysis | null>(null)

async function analyzeTicketWithAI() {
  if (!selectedTicket.value) return
  isAnalyzingAI.value = true
  try {
    const res = await api.post<AITicketAnalysis>('/api/v1/ai/analyze-ticket', {
      title: selectedTicket.value.title,
      description: selectedTicket.value.description
    }).catch(() => null)
    if (res) {
      aiSuggestion.value = res
    } else {
      aiSuggestion.value = null
      toast.add({ title: 'AI analysis is unavailable', color: 'error' })
    }
  } finally {
    isAnalyzingAI.value = false
  }
}

function applyAISuggestion() {
  if (aiSuggestion.value) {
    newCommentText.value = `[AI Operations Copilot Suggested Resolution]:\n${aiSuggestion.value.suggested_resolution}`
  }
}

async function fetchServiceItems() {
  try {
    const res = await api.get<ServiceCatalogItem[]>('/api/v1/services/items')
    serviceItems.value = res || []
  } catch (err: unknown) {
    console.error('Failed to load service items:', err)
  }
}

async function fetchTickets() {
  loading.value = true
  pageState.value = 'loading'
  error.value = null
  try {
    const params: Record<string, string | number | boolean | undefined> = {
      page: 1,
      page_size: 50
    }
    if (searchQuery.value) params.search = searchQuery.value
    if (selectedPriority.value && selectedPriority.value !== 'All') params.priority = selectedPriority.value
    if (selectedStatus.value && selectedStatus.value !== 'All') params.status = selectedStatus.value

    const res = await api.get<PaginatedResponse<Ticket>>('/api/v1/tickets', params)
    tickets.value = res?.data || []
    pageState.value = dataViewState(tickets.value)
  } catch (err: unknown) {
    tickets.value = []
    pageState.value = classifyApiError(err)
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    error.value = errObj?.data?.error?.message || errObj?.message || 'Failed to fetch tickets from Helpdesk Service.'
  } finally {
    loading.value = false
  }
}

function handleServiceItemSelect(itemId: string) {
  const item = serviceItems.value.find(i => i.id === itemId)
  if (item) {
    newTicket.service_item_id = item.id
    newTicket.category = item.category_name || 'General IT'
    newTicket.priority = item.default_priority
    if (!newTicket.title) {
      newTicket.title = item.name
    }
  }
}

async function handleCreateTicket() {
  if (!newTicket.title || !newTicket.description) {
    createError.value = 'Title and description are required.'
    return
  }

  creating.value = true
  createError.value = null
  try {
    const payload: CreateTicketPayload = {
      ...newTicket,
      requester_id: authStore.user?.id || 'emp-guest',
      requester_name: authStore.user?.full_name || 'Guest User',
      requester_email: authStore.user?.email || 'user@eomp.local'
    }

    await api.post<Ticket>('/api/v1/tickets', { ...payload })
    isCreateModalOpen.value = false
    // Reset
    Object.assign(newTicket, {
      title: '',
      description: '',
      service_item_id: '',
      category: 'Network & Access',
      priority: 'MEDIUM'
    })
    await fetchTickets()
  } catch (err: unknown) {
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    createError.value = errObj?.data?.error?.message || errObj?.message || 'Failed to create ticket.'
  } finally {
    creating.value = false
  }
}

async function openTicketDetail(ticket: Ticket) {
  selectedTicket.value = ticket
  isDetailModalOpen.value = true
  actionError.value = null
  comments.value = []
  timeline.value = []

  try {
    const [cRes, tRes] = await Promise.all([
      api.get<TicketComment[]>(`/api/v1/tickets/${ticket.id}/comments`),
      api.get<TicketTimeline[]>(`/api/v1/tickets/${ticket.id}/timeline`)
    ])
    comments.value = cRes || []
    timeline.value = tRes || []
  } catch (err) {
    console.error('Failed to load ticket details:', err)
  }
}

async function handleUpdateStatus(newStatus: string) {
  if (!selectedTicket.value) return
  actionLoading.value = true
  actionError.value = null
  try {
    const updated = await api.patch<Ticket>(`/api/v1/tickets/${selectedTicket.value.id}/status`, {
      status: newStatus,
      notes: `Status changed to ${newStatus}`,
      version: selectedTicket.value.version
    })
    selectedTicket.value = updated
    await fetchTickets()
    // Refresh timeline
    timeline.value = await api.get<TicketTimeline[]>(`/api/v1/tickets/${selectedTicket.value.id}/timeline`)
  } catch (err: unknown) {
    const errorObj = err as { statusCode?: number, status?: number, response?: { status?: number }, data?: { error?: { message?: string } }, message?: string }
    const status = errorObj.statusCode || errorObj.status || errorObj.response?.status
    actionError.value = status === 409
      ? 'This ticket was updated by another user. Refresh the ticket before retrying.'
      : errorObj.data?.error?.message || errorObj.message || 'Failed to update ticket status.'
  } finally {
    actionLoading.value = false
  }
}

async function handleAssignToMe() {
  if (!selectedTicket.value || !authStore.user) return
  actionLoading.value = true
  try {
    const updated = await api.patch<Ticket>(`/api/v1/tickets/${selectedTicket.value.id}/assign`, {
      assignee_id: authStore.user.id,
      assignee_name: `${authStore.user.full_name} (${authStore.user.role})`,
      version: selectedTicket.value.version
    })
    selectedTicket.value = updated
    await fetchTickets()
    timeline.value = await api.get<TicketTimeline[]>(`/api/v1/tickets/${selectedTicket.value.id}/timeline`)
  } catch (err: unknown) {
    console.error('Failed to assign ticket:', err)
  } finally {
    actionLoading.value = false
  }
}

async function handleAddComment() {
  if (!selectedTicket.value || !newCommentText.value.trim()) return
  actionLoading.value = true
  try {
    await api.post<TicketComment>(`/api/v1/tickets/${selectedTicket.value.id}/comments`, {
      content: newCommentText.value.trim(),
      is_internal: isInternalNote.value
    })
    newCommentText.value = ''
    // Refresh comments
    comments.value = await api.get<TicketComment[]>(`/api/v1/tickets/${selectedTicket.value.id}/comments`)
  } catch (err: unknown) {
    console.error('Failed to add comment:', err)
  } finally {
    actionLoading.value = false
  }
}

function formatSLACountdown(ticket: Ticket): { text: string, color: string } {
  if (ticket.status === 'RESOLVED' || ticket.status === 'CLOSED') {
    return { text: 'Resolved', color: 'text-emerald-400' }
  }

  const deadline = new Date(ticket.sla_resolution_deadline).getTime()
  const now = Date.now()
  const diffMinutes = Math.floor((deadline - now) / (1000 * 60))

  if (diffMinutes < 0) {
    const absMin = Math.abs(diffMinutes)
    const hours = Math.floor(absMin / 60)
    const mins = absMin % 60
    return {
      text: `Breached (+${hours > 0 ? `${hours}h ` : ''}${mins}m)`,
      color: 'text-rose-400 font-bold'
    }
  }

  const hours = Math.floor(diffMinutes / 60)
  const mins = diffMinutes % 60
  if (hours > 0) {
    return {
      text: `${hours}h ${mins}m left`,
      color: ticket.sla_status === 'WARNING' ? 'text-amber-400 font-bold' : 'text-slate-300'
    }
  }
  return {
    text: `${mins}m left`,
    color: 'text-amber-400 font-bold'
  }
}

// Watch filters
let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch([searchQuery, selectedPriority, selectedStatus], () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchTickets()
  }, 250)
})

onMounted(() => {
  fetchServiceItems()
  fetchTickets()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-ticket"
            class="w-6 h-6 text-indigo-400"
          />
          IT Help Desk Management
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          ITIL Incident Management, Dynamic SLA Engine & Automated Escalation (:8084)
        </p>
      </div>

      <button
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 transition-all hover:scale-105"
        @click="isCreateModalOpen = true"
      >
        <UIcon
          name="i-lucide-plus-circle"
          class="w-4 h-4"
        />
        <span>+ Create New Ticket</span>
      </button>
    </div>

    <ApiStatePanel
      v-if="!loading && (pageState === 'forbidden' || pageState === 'unavailable')"
      :state="pageState"
      resource="helpdesk tickets"
      @retry="fetchTickets"
    />

    <!-- Filter & Search Bar -->
    <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row items-center justify-between gap-3">
      <div class="relative w-full sm:w-80">
        <UIcon
          name="i-lucide-search"
          class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
        />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter by ticket number, title, requester..."
          class="w-full pl-9 pr-4 py-2 text-xs rounded-xl bg-slate-950/80 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-indigo-500"
        >
      </div>

      <div class="flex items-center gap-2.5 w-full sm:w-auto">
        <!-- Priority Filter -->
        <select
          v-model="selectedPriority"
          class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
        >
          <option value="All">
            All Priorities
          </option>
          <option value="URGENT">
            Urgent
          </option>
          <option value="HIGH">
            High
          </option>
          <option value="MEDIUM">
            Medium
          </option>
          <option value="LOW">
            Low
          </option>
        </select>

        <!-- Status Filter -->
        <select
          v-model="selectedStatus"
          class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
        >
          <option value="All">
            All Statuses
          </option>
          <option value="OPEN">
            Open
          </option>
          <option value="ASSIGNED">
            Assigned
          </option>
          <option value="IN_PROGRESS">
            In Progress
          </option>
          <option value="WAITING_USER">
            Waiting User
          </option>
          <option value="RESOLVED">
            Resolved
          </option>
          <option value="CLOSED">
            Closed
          </option>
        </select>

        <button
          class="p-2 rounded-xl bg-slate-800/60 hover:bg-slate-800 text-slate-300 transition-colors"
          title="Refresh"
          @click="fetchTickets"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
          />
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div
      v-if="error && pageState !== 'forbidden' && pageState !== 'unavailable'"
      class="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center justify-between"
    >
      <div class="flex items-center gap-2">
        <UIcon
          name="i-lucide-alert-triangle"
          class="w-4 h-4 text-rose-400"
        />
        <span>{{ error }}</span>
      </div>
      <button
        class="underline hover:text-white"
        @click="fetchTickets"
      >
        Retry
      </button>
    </div>

    <!-- Ticket Table -->
    <div class="overflow-hidden rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800/80">
            <tr>
              <th class="p-4">
                Ticket Number & Title
              </th>
              <th class="p-4">
                Category
              </th>
              <th class="p-4">
                Requester
              </th>
              <th class="p-4">
                Priority
              </th>
              <th class="p-4">
                Status
              </th>
              <th class="p-4">
                SLA Countdown
              </th>
              <th class="p-4 text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <!-- Loading Skeletons -->
            <template v-if="loading && tickets.length === 0">
              <tr
                v-for="i in 5"
                :key="i"
              >
                <td
                  colspan="7"
                  class="p-4 text-center text-slate-500 animate-pulse"
                >
                  Loading incidents from Helpdesk service...
                </td>
              </tr>
            </template>

            <!-- Empty State -->
            <tr v-else-if="pageState === 'empty'">
              <td
                colspan="7"
                class="p-8 text-center text-slate-500"
              >
                No tickets found matching your query.
              </td>
            </tr>

            <!-- Data Rows -->
            <tr
              v-for="t in tickets"
              :key="t.id"
              class="hover:bg-slate-800/40 transition-colors group cursor-pointer"
              @click="openTicketDetail(t)"
            >
              <td class="p-4">
                <div class="flex items-center gap-2.5">
                  <span class="font-mono text-[11px] font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                    {{ t.ticket_number }}
                  </span>
                  <span class="font-semibold text-white group-hover:text-indigo-300 transition-colors">
                    {{ t.title }}
                  </span>
                </div>
              </td>
              <td class="p-4 font-mono text-slate-400">
                {{ t.category }}
              </td>
              <td class="p-4 text-white">
                {{ t.requester_name }}
              </td>
              <td class="p-4">
                <span
                  class="px-2 py-0.5 rounded-full font-semibold text-[10px] border"
                  :class="t.priority === 'URGENT' ? 'bg-rose-500/15 text-rose-300 border-rose-500/30' : t.priority === 'HIGH' ? 'bg-amber-500/15 text-amber-300 border-amber-500/30' : 'bg-slate-800 text-slate-300 border-slate-700'"
                >
                  {{ t.priority }}
                </span>
              </td>
              <td class="p-4">
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-medium border"
                  :class="t.status === 'RESOLVED' ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : t.status === 'IN_PROGRESS' ? 'bg-blue-500/10 text-blue-300 border-blue-500/20' : 'bg-indigo-500/10 text-indigo-300 border-indigo-500/20'"
                >
                  {{ t.status }}
                </span>
              </td>
              <td class="p-4 font-mono">
                <span :class="formatSLACountdown(t).color">
                  {{ formatSLACountdown(t).text }}
                </span>
              </td>
              <td class="p-4 text-right">
                <button
                  class="px-2.5 py-1 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/40 text-indigo-300 border border-indigo-500/30 text-[11px] font-medium transition-colors"
                  @click.stop="openTicketDetail(t)"
                >
                  View &rarr;
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Ticket Modal -->
    <div
      v-if="isCreateModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-5">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-plus-circle"
              class="w-5 h-5 text-indigo-400"
            />
            Raise New Incident / Service Request
          </h3>
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

        <div
          v-if="createError"
          class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs"
        >
          {{ createError }}
        </div>

        <form
          class="space-y-4 text-xs"
          @submit.prevent="handleCreateTicket"
        >
          <!-- Service Catalog Item Picker -->
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Service Catalog Template</label>
            <select
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              @change="(e: any) => handleServiceItemSelect(e.target.value)"
            >
              <option value="">
                Select IT Service Template...
              </option>
              <option
                v-for="item in serviceItems"
                :key="item.id"
                :value="item.id"
              >
                [{{ item.category_name }}] {{ item.name }} (SLA: {{ Math.floor(item.sla_resolution_minutes / 60) }}h)
              </option>
            </select>
          </div>

          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Ticket Title *</label>
            <input
              v-model="newTicket.title"
              type="text"
              required
              placeholder="Brief summary of the issue or request..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
            >
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Category</label>
              <input
                v-model="newTicket.category"
                type="text"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Priority Level</label>
              <select
                v-model="newTicket.priority"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="URGENT">
                  Urgent (SLA: 2 Hours)
                </option>
                <option value="HIGH">
                  High (SLA: 4 Hours)
                </option>
                <option value="MEDIUM">
                  Medium (SLA: 8 Hours)
                </option>
                <option value="LOW">
                  Low (SLA: 24 Hours)
                </option>
              </select>
            </div>
          </div>

          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Detailed Description *</label>
            <textarea
              v-model="newTicket.description"
              rows="4"
              required
              placeholder="Provide exact symptoms, error codes, steps to reproduce, or requirements..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold"
              @click="isCreateModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="creating"
              class="px-5 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white font-semibold flex items-center gap-2 shadow-lg shadow-indigo-500/25 disabled:opacity-50"
            >
              <UIcon
                v-if="creating"
                name="i-lucide-loader-2"
                class="w-4 h-4 animate-spin"
              />
              <span>{{ creating ? 'Submitting...' : 'Submit Ticket' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Ticket Detail Modal -->
    <div
      v-if="isDetailModalOpen && selectedTicket"
      data-testid="ticket-detail-modal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-3xl max-h-[90vh] overflow-y-auto rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-6">
        <!-- Header -->
        <div class="flex items-start justify-between border-b border-slate-800 pb-4">
          <div>
            <div class="flex items-center gap-2 mb-1">
              <span class="font-mono text-xs font-bold text-indigo-400 px-2.5 py-0.5 rounded-md bg-indigo-500/10 border border-indigo-500/20">
                {{ selectedTicket.ticket_number }}
              </span>
              <span
                class="text-[10px] font-semibold px-2 py-0.5 rounded-full border"
                :class="selectedTicket.priority === 'URGENT' ? 'bg-rose-500/15 text-rose-300 border-rose-500/30' : 'bg-amber-500/15 text-amber-300 border-amber-500/30'"
              >
                {{ selectedTicket.priority }}
              </span>
              <span class="text-[10px] font-semibold px-2 py-0.5 rounded-full bg-blue-500/15 text-blue-300 border border-blue-500/30">
                {{ selectedTicket.status }}
              </span>
            </div>
            <h2 class="text-lg font-bold text-white">
              {{ selectedTicket.title }}
            </h2>
          </div>
          <button
            class="text-slate-400 hover:text-white"
            @click="isDetailModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <!-- Description Box -->
        <div class="p-4 rounded-2xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 leading-relaxed whitespace-pre-line">
          {{ selectedTicket.description }}
        </div>

        <!-- AI Operations Copilot Widget -->
        <div class="p-4 rounded-2xl bg-gradient-to-r from-indigo-950/40 to-slate-950/60 border border-indigo-500/30 space-y-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2 text-xs font-bold text-indigo-300">
              <UIcon
                name="i-lucide-bot"
                class="w-4 h-4 text-cyan-400 animate-pulse"
              />
              <span>AI Operations Copilot & RAG Assistant</span>
            </div>
            <button
              type="button"
              :disabled="isAnalyzingAI"
              class="px-3 py-1 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold flex items-center gap-1.5 shadow-md transition-all disabled:opacity-50"
              @click="analyzeTicketWithAI"
            >
              <UIcon
                v-if="isAnalyzingAI"
                name="i-lucide-loader-2"
                class="w-3.5 h-3.5 animate-spin"
              />
              <UIcon
                v-else
                name="i-lucide-sparkles"
                class="w-3.5 h-3.5 text-amber-300"
              />
              <span>{{ isAnalyzingAI ? 'Analyzing...' : '✨ Chẩn Đoán Với AI' }}</span>
            </button>
          </div>

          <div
            v-if="aiSuggestion"
            class="space-y-2 text-xs"
          >
            <div class="flex items-center justify-between text-[11px]">
              <span class="text-slate-400">Predicted Category: <strong class="text-cyan-300">{{ aiSuggestion.suggested_category }}</strong></span>
              <span class="text-emerald-400 font-mono font-bold">Confidence: {{ (aiSuggestion.confidence * 100).toFixed(0) }}%</span>
            </div>
            <div class="p-3 rounded-xl bg-slate-900/90 border border-slate-800 space-y-1">
              <p class="text-indigo-200 font-semibold">
                Dự Đoán Nguyên Nhân (Root Cause):
              </p>
              <p class="text-slate-300">
                {{ aiSuggestion.root_cause }}
              </p>
              <p class="text-indigo-200 font-semibold pt-1">
                Gợi Ý Xử Lý (Suggested SOP):
              </p>
              <p class="text-slate-300 whitespace-pre-line font-mono text-[11px]">
                {{ aiSuggestion.suggested_resolution }}
              </p>
            </div>
            <div class="flex items-center justify-end gap-2 pt-1">
              <button
                type="button"
                class="px-3 py-1 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold flex items-center gap-1 transition-all"
                @click="applyAISuggestion"
              >
                <UIcon
                  name="i-lucide-copy-check"
                  class="w-3.5 h-3.5"
                />
                <span>Paste vào bình luận xử lý</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Meta Grid -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div class="p-3 rounded-xl bg-slate-950/50 border border-slate-800/80">
            <p class="text-slate-500 text-[10px] font-semibold">
              Requester
            </p>
            <p class="font-medium text-white truncate mt-0.5">
              {{ selectedTicket.requester_name }}
            </p>
          </div>
          <div class="p-3 rounded-xl bg-slate-950/50 border border-slate-800/80">
            <p class="text-slate-500 text-[10px] font-semibold">
              Assignee
            </p>
            <p class="font-medium text-indigo-300 truncate mt-0.5">
              {{ selectedTicket.assignee_name || 'Unassigned' }}
            </p>
          </div>
          <div class="p-3 rounded-xl bg-slate-950/50 border border-slate-800/80">
            <p class="text-slate-500 text-[10px] font-semibold">
              SLA Status
            </p>
            <p
              class="font-medium truncate mt-0.5"
              :class="formatSLACountdown(selectedTicket).color"
            >
              {{ selectedTicket.sla_status }}
            </p>
          </div>
          <div class="p-3 rounded-xl bg-slate-950/50 border border-slate-800/80">
            <p class="text-slate-500 text-[10px] font-semibold">
              Category
            </p>
            <p class="font-medium text-slate-300 truncate mt-0.5">
              {{ selectedTicket.category }}
            </p>
          </div>
        </div>

        <!-- Workflow Lifecycle Actions -->
        <div class="p-4 rounded-2xl bg-indigo-950/20 border border-indigo-500/20 space-y-3">
          <div
            v-if="actionError"
            data-testid="ticket-action-error"
            role="alert"
            class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs"
          >
            {{ actionError }}
          </div>
          <p class="text-xs font-bold text-indigo-300 uppercase tracking-wider">
            Incident Transition Actions
          </p>
          <div class="flex flex-wrap items-center gap-2">
            <button
              v-if="!selectedTicket.assignee_id"
              :disabled="actionLoading"
              class="px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition-all"
              @click="handleAssignToMe"
            >
              🙋 Assign to Me
            </button>
            <button
              v-if="selectedTicket.status !== 'IN_PROGRESS' && selectedTicket.status !== 'RESOLVED' && selectedTicket.status !== 'CLOSED'"
              :disabled="actionLoading"
              class="px-3 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition-all"
              @click="handleUpdateStatus('IN_PROGRESS')"
            >
              🚀 Start Work
            </button>
            <button
              v-if="selectedTicket.status !== 'WAITING_USER' && selectedTicket.status !== 'RESOLVED' && selectedTicket.status !== 'CLOSED'"
              :disabled="actionLoading"
              class="px-3 py-1.5 rounded-xl bg-amber-600 hover:bg-amber-500 text-white text-xs font-semibold transition-all"
              @click="handleUpdateStatus('WAITING_USER')"
            >
              ⏳ Waiting User Info
            </button>
            <button
              v-if="selectedTicket.status !== 'RESOLVED' && selectedTicket.status !== 'CLOSED'"
              :disabled="actionLoading"
              class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition-all"
              @click="handleUpdateStatus('RESOLVED')"
            >
              ✅ Mark as Resolved
            </button>
            <button
              v-if="selectedTicket.status === 'RESOLVED'"
              :disabled="actionLoading"
              class="px-3 py-1.5 rounded-xl bg-slate-700 hover:bg-slate-600 text-white text-xs font-semibold transition-all"
              @click="handleUpdateStatus('CLOSED')"
            >
              🔒 Close Ticket
            </button>
          </div>
        </div>

        <!-- Comments & Activity Tabs -->
        <div class="space-y-4">
          <h4 class="text-xs font-bold text-slate-300 uppercase tracking-wider">
            Discussion & Activity Log ({{ comments.length }})
          </h4>

          <!-- Comment Stream -->
          <div class="space-y-3 max-h-48 overflow-y-auto">
            <div
              v-for="c in comments"
              :key="c.id"
              class="p-3 rounded-xl border text-xs space-y-1"
              :class="c.is_internal ? 'bg-amber-950/20 border-amber-500/30' : 'bg-slate-950 border-slate-800'"
            >
              <div class="flex items-center justify-between text-[11px]">
                <span class="font-bold text-white flex items-center gap-1.5">
                  {{ c.author_name }}
                  <span
                    v-if="c.is_internal"
                    class="text-[9px] px-1.5 py-0.2 rounded bg-amber-500/20 text-amber-300"
                  >INTERNAL NOTE</span>
                </span>
                <span class="text-slate-500">{{ new Date(c.created_at).toLocaleTimeString() }}</span>
              </div>
              <p class="text-slate-300">
                {{ c.content }}
              </p>
            </div>
          </div>

          <!-- Add Comment Input -->
          <div class="flex flex-col gap-2">
            <textarea
              v-model="newCommentText"
              rows="2"
              placeholder="Add response to user or internal technical notes..."
              class="w-full p-3 rounded-xl bg-slate-950 border border-slate-800 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500"
            />
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-2 text-xs text-slate-400 cursor-pointer">
                <input
                  v-model="isInternalNote"
                  type="checkbox"
                  class="rounded bg-slate-950 border-slate-800"
                >
                <span>Internal technician note only</span>
              </label>
              <button
                :disabled="actionLoading || !newCommentText.trim()"
                class="px-4 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold disabled:opacity-50"
                @click="handleAddComment"
              >
                Post Note
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
