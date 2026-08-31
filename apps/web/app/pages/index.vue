<script setup lang="ts">
import type {
  AssetStats,
  AuditLog,
  PaginatedResponse,
  ServiceHealthStatus,
  Ticket,
  WorkflowStats
} from '~/types'

definePageMeta({
  layout: 'default'
})

const api = useApi()
const authStore = useAuthStore()
const toast = useToast()
const isRefreshing = ref(false)

const ticketPage = ref<PaginatedResponse<Ticket> | null>(null)
const assetStats = ref<AssetStats | null>(null)
const employeeTotal = ref<number | null>(null)
const workflowStats = ref<WorkflowStats | null>(null)
const infraServices = ref<ServiceHealthStatus[]>([])
const auditLogs = ref<AuditLog[]>([])
const dashboardErrors = reactive<Record<string, 'forbidden' | 'unavailable' | null>>({
  tickets: null, assets: null, employees: null, workflows: null, services: null, audits: null
})

const activeTab = ref('all')

const filteredTickets = computed(() => {
  const tickets = ticketPage.value?.data ?? []
  if (activeTab.value === 'urgent') {
    return tickets.filter(t => t.priority === 'URGENT' || t.priority === 'HIGH')
  }
  if (activeTab.value === 'hardware') {
    return tickets.filter(t => t.category.toUpperCase().includes('HARDWARE'))
  }
  return tickets.slice(0, 4)
})

const urgentTicketCount = computed(() => (ticketPage.value?.data ?? []).filter(t => t.priority === 'URGENT' || t.priority === 'HIGH').length)
const hardwareTicketCount = computed(() => (ticketPage.value?.data ?? []).filter(t => t.category.toUpperCase().includes('HARDWARE')).length)
const onlineServiceCount = computed(() => infraServices.value.filter(service => service.status === 'ONLINE').length)
const allSystemsNominal = computed(() => infraServices.value.length > 0 && onlineServiceCount.value === infraServices.value.length)

const stats = computed(() => [
  {
    label: 'Total IT Tickets', value: ticketPage.value?.total ?? '—', subtext: `${urgentTicketCount.value} urgent/high loaded`, change: 'Live API data', trend: 'neutral',
    icon: 'i-lucide-ticket', gradient: 'from-amber-500 to-orange-600', borderGlow: 'hover:border-amber-500/50', badgeColor: 'bg-amber-500/15 text-amber-300 border-amber-500/30'
  },
  {
    label: 'Total IT Assets', value: assetStats.value?.total_assets ?? '—', subtext: `${assetStats.value?.in_use ?? 0} in use`, change: `${assetStats.value?.in_maintenance ?? 0} in maintenance`, trend: 'neutral',
    icon: 'i-lucide-laptop', gradient: 'from-emerald-500 to-teal-600', borderGlow: 'hover:border-emerald-500/50', badgeColor: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30'
  },
  {
    label: 'Total Employees', value: employeeTotal.value ?? '—', subtext: 'Employee directory', change: 'Live API data', trend: 'neutral',
    icon: 'i-lucide-users', gradient: 'from-blue-500 to-indigo-600', borderGlow: 'hover:border-blue-500/50', badgeColor: 'bg-blue-500/15 text-blue-300 border-blue-500/30'
  },
  {
    label: 'Pending Approvals', value: workflowStats.value?.pending_approvals ?? '—', subtext: `${workflowStats.value?.active_instances ?? 0} active workflows`, change: `${workflowStats.value?.completed_today ?? 0} completed today`, trend: 'warning',
    icon: 'i-lucide-clock', gradient: 'from-purple-500 to-pink-600', borderGlow: 'hover:border-purple-500/50', badgeColor: 'bg-purple-500/15 text-purple-300 border-purple-500/30'
  }
])

const recentAuditEvents = computed(() => auditLogs.value.map(log => ({
  action: formatLabel(log.event_type),
  detail: `${log.actor_name || log.actor_email || 'Unknown actor'} · ${log.service_name} · ${formatLabel(log.resource_type)} ${log.resource_id || ''}`.trim(),
  time: formatRelativeTime(log.created_at),
  icon: log.status === 'SUCCESS' ? 'i-lucide-shield-check' : 'i-lucide-shield-alert',
  color: log.status === 'SUCCESS'
    ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
    : 'text-amber-400 bg-amber-500/10 border-amber-500/20'
})))

function formatLabel(value: string) {
  return value.toLowerCase().replaceAll('_', ' ').replace(/\b\w/g, char => char.toUpperCase())
}

function formatRelativeTime(value: string) {
  const timestamp = new Date(value).getTime()
  if (!Number.isFinite(timestamp)) return 'Unknown time'
  const seconds = Math.max(0, Math.floor((Date.now() - timestamp) / 1000))
  if (seconds < 60) return `${seconds}s ago`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
  return `${Math.floor(seconds / 86400)}d ago`
}

function formatSLA(ticket: Ticket) {
  if (ticket.status === 'RESOLVED' || ticket.status === 'CLOSED') return 'Completed'
  const deadline = new Date(ticket.sla_resolution_deadline).getTime()
  if (!Number.isFinite(deadline)) return 'Unavailable'
  const minutes = Math.floor((deadline - Date.now()) / 60000)
  if (minutes < 0) return `Breached ${Math.abs(minutes)}m`
  if (minutes < 60) return `${minutes}m left`
  return `${Math.floor(minutes / 60)}h ${minutes % 60}m left`
}

function serviceDotClass(status: string) {
  if (status === 'ONLINE') return 'bg-emerald-400 shadow-emerald-400/50'
  if (status === 'DEGRADED') return 'bg-amber-400 shadow-amber-400/50'
  if (status === 'OFFLINE') return 'bg-rose-400 shadow-rose-400/50'
  return 'bg-slate-500 shadow-slate-500/30'
}

async function loadDashboard() {
  if (!authStore.user && authStore.token) {
    await authStore.fetchCurrentUser()
  }
  if (!authStore.user) return

  const role = authStore.role
  const canOperate = role === 'ROLE_ADMIN' || role === 'ROLE_MANAGER' || role === 'ROLE_AGENT'
  const canManagePeople = role === 'ROLE_ADMIN' || role === 'ROLE_MANAGER'

  const load = async <T>(key: string, request: () => Promise<T>): Promise<T | null> => {
    dashboardErrors[key] = null
    try {
      return await request()
    } catch (error: unknown) {
      const status = (error as { statusCode?: number }).statusCode
      dashboardErrors[key] = status === 403 ? 'forbidden' : 'unavailable'
      return null
    }
  }

  const [tickets, assets, employees, workflows, services, audits] = await Promise.all([
    load('tickets', () => api.get<PaginatedResponse<Ticket>>('/api/v1/tickets', { page: 1, page_size: 100 })),
    canOperate ? load('assets', () => api.get<AssetStats>('/api/v1/assets/stats')) : Promise.resolve(null),
    canManagePeople ? load('employees', () => api.get<PaginatedResponse<unknown>>('/api/v1/employees', { page: 1, page_size: 1 })) : Promise.resolve(null),
    canOperate ? load('workflows', () => api.get<WorkflowStats>('/api/v1/workflows/stats')) : Promise.resolve(null),
    canManagePeople ? load('services', () => api.get<ServiceHealthStatus[]>('/api/v1/monitoring/services')) : Promise.resolve(null),
    canManagePeople ? load('audits', () => api.get<PaginatedResponse<AuditLog>>('/api/v1/audit/logs', { page: 1, page_size: 3 })) : Promise.resolve(null)
  ])

  ticketPage.value = tickets
  assetStats.value = assets
  employeeTotal.value = employees?.total ?? null
  workflowStats.value = workflows
  infraServices.value = services ?? []
  auditLogs.value = audits?.data ?? []
  const attemptedResults: unknown[] = [tickets]
  if (canOperate) attemptedResults.push(assets, workflows)
  if (canManagePeople) attemptedResults.push(employees)
  if (canManagePeople) attemptedResults.push(services, audits)
  if (attemptedResults.some(result => result === null)) {
    toast.add({ title: 'Some dashboard data is unavailable', color: 'warning' })
  }
}

async function refreshInfrastructure() {
  if (authStore.role !== 'ROLE_ADMIN' && authStore.role !== 'ROLE_MANAGER') return
  isRefreshing.value = true
  try {
    infraServices.value = await api.get<ServiceHealthStatus[]>('/api/v1/monitoring/services')
  } catch {
    infraServices.value = []
    toast.add({ title: 'Service health is unavailable', color: 'error' })
  } finally {
    isRefreshing.value = false
  }
}

onMounted(loadDashboard)
</script>

<template>
  <div class="space-y-8 max-w-7xl mx-auto">
    <!-- Hero / Welcome Banner -->
    <div class="relative overflow-hidden rounded-3xl bg-gradient-to-r from-indigo-900/60 via-slate-900/80 to-slate-900/90 border border-indigo-500/20 p-6 md:p-8 shadow-2xl backdrop-blur-xl">
      <!-- Glow effect -->
      <div class="absolute -right-20 -bottom-20 w-80 h-80 bg-indigo-500/20 rounded-full blur-3xl pointer-events-none" />
      <div class="absolute top-0 right-1/4 w-60 h-60 bg-cyan-500/15 rounded-full blur-3xl pointer-events-none" />

      <div class="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
        <div class="space-y-2 max-w-2xl">
          <div class="flex items-center gap-2">
            <span class="px-2.5 py-1 rounded-full text-xs font-semibold bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-shield"
                class="w-3.5 h-3.5 text-indigo-400"
              />
              EOMP Command Center v0.1
            </span>
            <span
              class="px-2.5 py-1 rounded-full text-xs font-semibold border flex items-center gap-1.5"
              :class="allSystemsNominal ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30' : 'bg-slate-500/15 text-slate-300 border-slate-500/30'"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="allSystemsNominal ? 'bg-emerald-400' : 'bg-slate-400'"
              />
              {{ infraServices.length ? `${onlineServiceCount}/${infraServices.length} Services Online` : 'Health Unavailable' }}
            </span>
          </div>

          <h1 class="text-2xl md:text-3xl font-extrabold tracking-tight text-white">
            Enterprise Operations Management Platform
          </h1>
          <p class="text-slate-300 text-sm leading-relaxed">
            Unified control center for IT Service Management, Hardware Asset Tracking, Org Hierarchy, and Automated AI Operations.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-3 shrink-0">
          <NuxtLink
            to="/ai"
            class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700/80 text-slate-200 text-sm font-medium transition-all hover:scale-105"
          >
            <UIcon
              name="i-lucide-sparkles"
              class="w-4 h-4 text-indigo-400"
            />
            <span>AI Ops Assistant</span>
          </NuxtLink>

          <NuxtLink
            to="/helpdesk"
            class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-indigo-600 via-indigo-500 to-cyan-500 hover:from-indigo-500 hover:to-cyan-400 text-white text-sm font-semibold shadow-lg shadow-indigo-500/25 transition-all hover:scale-105"
          >
            <UIcon
              name="i-lucide-plus-circle"
              class="w-4 h-4"
            />
            <span>+ Create Ticket</span>
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- 4 Key Metrics Overview Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
      <div
        v-for="stat in stats"
        :key="stat.label"
        class="group relative overflow-hidden rounded-2xl bg-slate-900/60 backdrop-blur-xl border border-slate-800/80 p-5 transition-all duration-300 hover:bg-slate-900/90 hover:-translate-y-1 hover:shadow-xl hover:shadow-indigo-500/5"
        :class="stat.borderGlow"
      >
        <!-- Top Row: Icon & Badge -->
        <div class="flex items-center justify-between mb-4">
          <div
            class="w-12 h-12 rounded-xl flex items-center justify-center shadow-lg bg-gradient-to-br transition-transform duration-300 group-hover:scale-110"
            :class="stat.gradient"
          >
            <UIcon
              :name="stat.icon"
              class="w-6 h-6 text-white"
            />
          </div>
          <span
            class="text-[11px] font-semibold px-2.5 py-1 rounded-full border"
            :class="stat.badgeColor"
          >
            {{ stat.subtext }}
          </span>
        </div>

        <!-- Value & Label -->
        <div>
          <p class="text-3xl font-extrabold tracking-tight text-white mb-1">
            {{ stat.value }}
          </p>
          <p class="text-xs font-medium text-slate-400">
            {{ stat.label }}
          </p>
        </div>

        <!-- Trend Footer -->
        <div class="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs">
          <span class="text-slate-400 flex items-center gap-1">
            <UIcon
              v-if="stat.trend === 'up'"
              name="i-lucide-trending-up"
              class="w-3.5 h-3.5 text-emerald-400"
            />
            <UIcon
              v-else-if="stat.trend === 'warning'"
              name="i-lucide-alert-circle"
              class="w-3.5 h-3.5 text-amber-400"
            />
            <UIcon
              v-else
              name="i-lucide-check-circle-2"
              class="w-3.5 h-3.5 text-slate-400"
            />
            <span class="text-slate-300">{{ stat.change }}</span>
          </span>
          <UIcon
            name="i-lucide-arrow-right"
            class="w-3.5 h-3.5 text-slate-400 group-hover:text-indigo-400 group-hover:translate-x-1 transition-all"
          />
        </div>
      </div>
    </div>

    <!-- Main Content Split Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
      <!-- Left 2 Cols: Live Helpdesk Tickets & Quick Operations -->
      <div class="lg:col-span-2 space-y-8">
        <!-- Live Ticket Queue Stream -->
        <div class="rounded-3xl bg-slate-900/60 backdrop-blur-xl border border-slate-800/80 p-6 shadow-xl space-y-5">
          <!-- Section Header with Filter Tabs -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-slate-800/80">
            <div>
              <h2 class="text-lg font-bold text-white flex items-center gap-2">
                <UIcon
                  name="i-lucide-ticket"
                  class="w-5 h-5 text-indigo-400"
                />
                Active Help Desk Tickets
              </h2>
              <p class="text-xs text-slate-400 mt-0.5">
                Real-time incident response & SLA tracking
              </p>
            </div>

            <!-- Filter Pills -->
            <div class="flex items-center gap-1.5 p-1 rounded-xl bg-slate-950/80 border border-slate-800/80 text-xs">
              <button
                class="px-3 py-1 rounded-lg font-medium transition-all"
                :class="activeTab === 'all' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
                @click="activeTab = 'all'"
              >
                All ({{ ticketPage?.total ?? 0 }})
              </button>
              <button
                class="px-3 py-1 rounded-lg font-medium transition-all"
                :class="activeTab === 'urgent' ? 'bg-rose-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
                @click="activeTab = 'urgent'"
              >
                Urgent / High ({{ urgentTicketCount }})
              </button>
              <button
                class="px-3 py-1 rounded-lg font-medium transition-all"
                :class="activeTab === 'hardware' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-400 hover:text-slate-200'"
                @click="activeTab = 'hardware'"
              >
                Hardware ({{ hardwareTicketCount }})
              </button>
            </div>
          </div>

          <!-- Ticket Items List -->
          <div class="space-y-3">
            <p
              v-if="filteredTickets.length === 0"
              class="py-8 text-center text-xs text-slate-500"
            >
              No tickets returned by the Helpdesk API.
            </p>
            <div
              v-for="ticket in filteredTickets"
              :key="ticket.id"
              class="group p-4 rounded-2xl bg-slate-950/40 border border-slate-800/60 hover:border-indigo-500/40 hover:bg-slate-950/80 transition-all duration-200"
            >
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-2">
                <div class="flex items-center gap-2.5">
                  <span class="font-mono text-xs font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                    {{ ticket.ticket_number }}
                  </span>
                  <h3 class="text-sm font-semibold text-white group-hover:text-indigo-300 transition-colors">
                    {{ ticket.title }}
                  </h3>
                </div>

                <div class="flex items-center gap-2 shrink-0">
                  <span
                    class="text-[11px] font-semibold px-2 py-0.5 rounded-full border"
                    :class="ticket.priority === 'URGENT'
                      ? 'bg-rose-500/15 text-rose-300 border-rose-500/30 animate-pulse'
                      : ticket.priority === 'HIGH'
                        ? 'bg-amber-500/15 text-amber-300 border-amber-500/30'
                        : 'bg-slate-800 text-slate-300 border-slate-700'"
                  >
                    {{ formatLabel(ticket.priority) }}
                  </span>

                  <span
                    class="text-[11px] font-medium px-2 py-0.5 rounded-full"
                    :class="ticket.status === 'RESOLVED'
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                      : 'bg-indigo-500/10 text-indigo-300 border border-indigo-500/20'"
                  >
                    {{ formatLabel(ticket.status) }}
                  </span>
                </div>
              </div>

              <!-- Ticket Sub Metadata -->
              <div class="flex flex-wrap items-center justify-between gap-3 text-xs text-slate-400 pt-2 border-t border-slate-800/40">
                <div class="flex items-center gap-4">
                  <span class="flex items-center gap-1.5 text-slate-300">
                    <UIcon
                      name="i-lucide-user"
                      class="w-3.5 h-3.5 text-slate-400"
                    />
                    {{ ticket.requester_name }}<span v-if="ticket.department_id"> ({{ ticket.department_id }})</span>
                  </span>
                  <span class="hidden sm:inline-block text-slate-400 font-mono">
                    {{ ticket.category }}
                  </span>
                </div>

                <div class="flex items-center gap-3">
                  <span class="flex items-center gap-1 text-slate-400">
                    <UIcon
                      name="i-lucide-clock"
                      class="w-3.5 h-3.5"
                    />
                    {{ formatRelativeTime(ticket.created_at) }}
                  </span>
                  <span class="font-mono text-[11px] text-amber-400 bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/20">
                    SLA: {{ formatSLA(ticket) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer Action -->
          <div class="pt-2 text-center">
            <NuxtLink
              to="/helpdesk"
              class="inline-flex items-center gap-2 text-xs font-semibold text-indigo-400 hover:text-indigo-300 hover:underline"
            >
              <span>View All {{ ticketPage?.total ?? 0 }} Tickets in IT Help Desk</span>
              <UIcon
                name="i-lucide-arrow-right"
                class="w-4 h-4"
              />
            </NuxtLink>
          </div>
        </div>

        <!-- Quick Enterprise Actions Hub -->
        <div class="rounded-3xl bg-slate-900/60 backdrop-blur-xl border border-slate-800/80 p-6 shadow-xl space-y-4">
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-zap"
              class="w-5 h-5 text-amber-400"
            />
            Quick Operations Hub
          </h2>

          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <NuxtLink
              to="/helpdesk"
              class="flex flex-col items-center text-center p-4 rounded-2xl bg-slate-950/50 border border-slate-800/80 hover:border-indigo-500/50 hover:bg-indigo-600/10 transition-all group"
            >
              <div class="w-10 h-10 rounded-xl bg-indigo-500/10 text-indigo-400 flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
                <UIcon
                  name="i-lucide-ticket"
                  class="w-5 h-5"
                />
              </div>
              <span class="text-xs font-semibold text-slate-200">New Ticket</span>
              <span class="text-[10px] text-slate-400">Incident Triage</span>
            </NuxtLink>

            <NuxtLink
              to="/assets"
              class="flex flex-col items-center text-center p-4 rounded-2xl bg-slate-950/50 border border-slate-800/80 hover:border-emerald-500/50 hover:bg-emerald-600/10 transition-all group"
            >
              <div class="w-10 h-10 rounded-xl bg-emerald-500/10 text-emerald-400 flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
                <UIcon
                  name="i-lucide-laptop"
                  class="w-5 h-5"
                />
              </div>
              <span class="text-xs font-semibold text-slate-200">Register Asset</span>
              <span class="text-[10px] text-slate-400">Hardware & License</span>
            </NuxtLink>

            <NuxtLink
              to="/employees"
              class="flex flex-col items-center text-center p-4 rounded-2xl bg-slate-950/50 border border-slate-800/80 hover:border-blue-500/50 hover:bg-blue-600/10 transition-all group"
            >
              <div class="w-10 h-10 rounded-xl bg-blue-500/10 text-blue-400 flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
                <UIcon
                  name="i-lucide-user-plus"
                  class="w-5 h-5"
                />
              </div>
              <span class="text-xs font-semibold text-slate-200">Add Employee</span>
              <span class="text-[10px] text-slate-400">Org Assignment</span>
            </NuxtLink>

            <NuxtLink
              to="/ai"
              class="flex flex-col items-center text-center p-4 rounded-2xl bg-slate-950/50 border border-slate-800/80 hover:border-purple-500/50 hover:bg-purple-600/10 transition-all group"
            >
              <div class="w-10 h-10 rounded-xl bg-purple-500/10 text-purple-400 flex items-center justify-center mb-2 group-hover:scale-110 transition-transform">
                <UIcon
                  name="i-lucide-sparkles"
                  class="w-5 h-5"
                />
              </div>
              <span class="text-xs font-semibold text-slate-200">AI Assistant</span>
              <span class="text-[10px] text-slate-400">LLM Triage & RAG</span>
            </NuxtLink>
          </div>
        </div>
      </div>

      <!-- Right 1 Col: Infrastructure Health & AI Assistant Preview -->
      <div class="space-y-8">
        <!-- Infrastructure Cluster Live Matrix -->
        <div class="rounded-3xl bg-slate-900/60 backdrop-blur-xl border border-slate-800/80 p-6 shadow-xl space-y-4">
          <div class="flex items-center justify-between pb-3 border-b border-slate-800/80">
            <div>
              <h2 class="text-base font-bold text-white flex items-center gap-2">
                <UIcon
                  name="i-lucide-server"
                  class="w-4 h-4 text-emerald-400"
                />
                Microservice Cluster
              </h2>
              <p class="text-xs text-slate-400">
                Active health probes from the Gateway
              </p>
            </div>

            <button
              class="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 text-slate-300 hover:text-white transition-colors"
              title="Refresh Service Health"
              @click="refreshInfrastructure"
            >
              <UIcon
                name="i-lucide-refresh-cw"
                class="w-4 h-4"
                :class="isRefreshing && 'animate-spin text-indigo-400'"
              />
            </button>
          </div>

          <!-- Services Grid -->
          <div class="space-y-2">
            <div
              v-for="svc in infraServices"
              :key="svc.name"
              class="flex items-center justify-between p-2.5 rounded-xl bg-slate-950/50 border border-slate-800/60 hover:border-slate-700/80 transition-colors"
            >
              <div class="flex items-center gap-2.5">
                <span
                  class="w-2 h-2 rounded-full shadow-sm"
                  :class="serviceDotClass(svc.status)"
                />
                <div>
                  <p class="text-xs font-semibold text-slate-200">
                    {{ svc.name }}
                  </p>
                  <p class="text-[10px] font-mono text-slate-400">
                    Port: {{ svc.port }} · {{ svc.category }} · {{ formatLabel(svc.status) }}
                  </p>
                </div>
              </div>

              <div class="text-right">
                <span class="text-[10px] font-mono font-bold text-slate-300 bg-slate-500/10 px-1.5 py-0.5 rounded border border-slate-500/20">
                  {{ svc.latency_ms }}ms
                </span>
              </div>
            </div>
          </div>

          <p
            v-if="infraServices.length === 0"
            class="py-5 text-center text-xs text-slate-500"
          >
            Service health is unavailable.
          </p>

          <div class="pt-2 flex items-center justify-between text-xs text-slate-400">
            <span>Observed online: <strong class="text-emerald-400 font-mono">{{ onlineServiceCount }}/{{ infraServices.length }}</strong></span>
            <NuxtLink
              to="/monitoring"
              class="text-indigo-400 hover:underline"
            >
              Monitoring details &rarr;
            </NuxtLink>
          </div>
        </div>

        <!-- AI Assistant Quick Interaction Card -->
        <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-indigo-950/80 via-slate-900/90 to-purple-950/60 border border-indigo-500/30 p-6 shadow-xl space-y-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-500 to-purple-500 flex items-center justify-center shadow-lg shadow-indigo-500/30">
              <UIcon
                name="i-lucide-bot"
                class="w-5 h-5 text-white"
              />
            </div>
            <div>
              <h3 class="text-sm font-bold text-white">
                AI Ops Assistant
              </h3>
              <p class="text-[11px] text-indigo-200/70">
                Provider status is reported in AI Operations
              </p>
            </div>
          </div>

          <div class="p-3 rounded-2xl bg-slate-950/60 border border-indigo-500/20 text-xs text-slate-300 space-y-2">
            <p class="leading-relaxed">
              Open the AI workspace to submit a real prompt. This dashboard does not generate synthetic recommendations.
            </p>
          </div>

          <NuxtLink
            to="/ai"
            class="block w-full py-2 text-center rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs transition-colors shadow-md shadow-indigo-600/30"
          >
            Open AI Operations
          </NuxtLink>
        </div>

        <!-- Recent Audit Stream -->
        <div class="rounded-3xl bg-slate-900/60 backdrop-blur-xl border border-slate-800/80 p-6 shadow-xl space-y-3">
          <h2 class="text-sm font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-shield-check"
              class="w-4 h-4 text-cyan-400"
            />
            Recent Audit Trail
          </h2>

          <div class="space-y-3 pt-2">
            <p
              v-if="recentAuditEvents.length === 0"
              class="py-5 text-center text-xs text-slate-500"
            >
              No audit events are available for this account.
            </p>
            <div
              v-for="(event, idx) in recentAuditEvents"
              :key="idx"
              class="flex items-start gap-3 text-xs"
            >
              <div
                class="p-1.5 rounded-lg border shrink-0 mt-0.5"
                :class="event.color"
              >
                <UIcon
                  :name="event.icon"
                  class="w-3.5 h-3.5"
                />
              </div>
              <div class="flex-1 min-w-0">
                <p class="font-semibold text-slate-200 truncate">
                  {{ event.action }}
                </p>
                <p class="text-slate-400 text-[11px] leading-tight">
                  {{ event.detail }}
                </p>
                <span class="text-[10px] text-slate-400 font-mono">{{ event.time }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
