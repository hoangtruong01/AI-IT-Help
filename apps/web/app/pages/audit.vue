<script setup lang="ts">
import type {
  AuditLog,
  SecurityEvent,
  AuditStats,
  PaginatedResponse
} from '~/types'
import { classifyApiError, dataViewState, type ApiViewState } from '~/utils/api-view-state'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()

// State
const loading = ref(false)
const pageState = ref<ApiViewState>('loading')
const auditLogs = ref<AuditLog[]>([])
const totalRecords = ref(0)
const currentPage = ref(1)
const pageSize = ref(15)

// Stats
const stats = ref<AuditStats>({
  total_logs: 0,
  blocked_violations: 0,
  active_security_alerts: 0,
  immutable_proofs_count: 0,
  success_count: 0,
  forbidden_count: 0
})

const securityEvents = ref<SecurityEvent[]>([])

// Filter State
const search = ref('')
const selectedEventType = ref('ALL')
const selectedStatus = ref('ALL')
const selectedService = ref('ALL')

// Diff Modal State
const isDiffModalOpen = ref(false)
const selectedAuditLog = ref<AuditLog | null>(null)

// Fetch Audit Logs & Stats
async function fetchAuditData() {
  loading.value = true
  pageState.value = 'loading'
  try {
    const [listRes, statsRes, secRes] = await Promise.all([
      api.get<PaginatedResponse<AuditLog>>('/api/v1/audit/logs', {
        event_type: selectedEventType.value === 'ALL' ? undefined : selectedEventType.value,
        status: selectedStatus.value === 'ALL' ? undefined : selectedStatus.value,
        service: selectedService.value === 'ALL' ? undefined : selectedService.value,
        search: search.value || undefined,
        page: currentPage.value,
        page_size: pageSize.value
      }),
      api.get<AuditStats>('/api/v1/audit/stats').catch(() => null),
      api.get<SecurityEvent[]>('/api/v1/audit/security-events').catch(() => null)
    ])

    auditLogs.value = listRes?.data ?? []
    pageState.value = dataViewState(auditLogs.value)
    totalRecords.value = listRes?.total ?? 0
    stats.value = statsRes ?? {
      total_logs: 0,
      blocked_violations: 0,
      active_security_alerts: 0,
      immutable_proofs_count: 0,
      success_count: 0,
      forbidden_count: 0
    }
    securityEvents.value = secRes ?? []
    if (!listRes || !statsRes || !secRes) {
      toast.add({ title: 'Some audit data could not be loaded', color: 'warning' })
    }
  } catch (err: unknown) {
    auditLogs.value = []
    pageState.value = classifyApiError(err)
  } finally {
    loading.value = false
  }
}

// Open Diff Modal
function openDiffModal(log: AuditLog) {
  selectedAuditLog.value = log
  isDiffModalOpen.value = true
}

// Copy the persisted HMAC proof.
function copyChecksum(hash: string) {
  navigator.clipboard.writeText(hash)
  toast.add({
    title: 'Checksum Copied',
    description: 'HMAC-SHA256 chain proof copied to clipboard.',
    color: 'success'
  })
}

onMounted(() => {
  fetchAuditData()
})
</script>

<template>
  <div class="space-y-7 max-w-7xl mx-auto pb-12">
    <ApiStatePanel
      v-if="!loading && pageState !== 'loading' && pageState !== 'ready'"
      :state="pageState"
      resource="audit events"
      @retry="fetchAuditData"
    />
    <!-- Header with Breadcrumbs & Actions -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-emerald-400 mb-1">
          <UIcon
            name="i-lucide-shield-check"
            class="w-4 h-4 animate-pulse"
          />
          <span>SOC2 Type II & ISO 27001 Compliance Engine</span>
        </div>
        <h1 class="text-3xl font-extrabold text-white tracking-tight flex items-center gap-3">
          Security Hardening & Immutable Audit Trail
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Cryptographically Verified Audit Stream, Strict RBAC Policies, Data Masking & Threat Defense.
        </p>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-wrap items-center gap-2.5">
        <button
          :disabled="loading"
          class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-900/90 hover:bg-slate-800 border border-slate-700/80 text-white text-xs font-semibold shadow-md transition-all disabled:opacity-50"
          @click="fetchAuditData"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            :class="['w-4 h-4 text-cyan-400', loading ? 'animate-spin' : '']"
          />
          <span>Refresh Stream</span>
        </button>
      </div>
    </div>

    <!-- Security KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- 1. Total Events -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-emerald-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Immutable Audit Events</span>
          <UIcon
            name="i-lucide-database"
            class="w-4 h-4 text-emerald-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">{{ stats.total_logs }}</span>
          <span class="text-xs text-slate-400 font-medium">records</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-emerald-400 font-semibold">
          <UIcon
            name="i-lucide-check-circle"
            class="w-3.5 h-3.5"
          />
          <span>{{ stats.immutable_proofs_count }} verified HMAC proofs</span>
        </div>
      </div>

      <!-- 2. Blocked Violations -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-rose-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Blocked RBAC Violations</span>
          <UIcon
            name="i-lucide-shield-alert"
            class="w-4 h-4 text-rose-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-rose-400 tracking-tight">{{ stats.blocked_violations }}</span>
          <span class="text-xs text-slate-400 font-medium">attempts blocked</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-rose-400 font-semibold">
          <UIcon
            name="i-lucide-lock"
            class="w-3.5 h-3.5"
          />
          <span>HTTP 403 Forbidden enforced</span>
        </div>
      </div>

      <!-- 3. Active Security Alerts -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-amber-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Active Threat Signals</span>
          <UIcon
            name="i-lucide-zap"
            class="w-4 h-4 text-amber-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">{{ stats.active_security_alerts }}</span>
          <span class="text-xs text-slate-400 font-medium">signals</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-amber-400 font-semibold">
          <span class="w-2 h-2 rounded-full bg-amber-400 animate-ping" />
          <span>Sliding Window Rate Limiter ON</span>
        </div>
      </div>

      <!-- 4. Data Masking Engine -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-cyan-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Data Masking & Sanitization</span>
          <UIcon
            name="i-lucide-eye-off"
            class="w-4 h-4 text-cyan-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-cyan-400 tracking-tight">ACTIVE</span>
          <span class="text-xs text-slate-400 font-medium">Zero-Leak</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-cyan-400 font-semibold">
          <UIcon
            name="i-lucide-sparkles"
            class="w-3.5 h-3.5"
          />
          <span>Passwords & JWTs Sanitized</span>
        </div>
      </div>
    </div>

    <!-- Multi-Filter & Search Bar -->
    <div class="p-4 rounded-2xl bg-slate-900/80 border border-slate-800 flex flex-wrap items-center justify-between gap-3 shadow-lg">
      <div class="flex flex-wrap items-center gap-3 flex-1">
        <!-- Search Input -->
        <div class="relative w-full sm:w-64">
          <UIcon
            name="i-lucide-search"
            class="w-4 h-4 text-slate-400 absolute left-3 top-2.5"
          />
          <input
            v-model="search"
            type="text"
            placeholder="Search email, IP, action, resource..."
            class="w-full pl-9 pr-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-emerald-400"
            @input="fetchAuditData"
          >
        </div>

        <!-- Event Type Filter -->
        <select
          v-model="selectedEventType"
          class="px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white focus:outline-none focus:border-emerald-400"
          @change="fetchAuditData"
        >
          <option value="ALL">
            All Event Types
          </option>
          <option value="AUTH_LOGIN_SUCCESS">
            AUTH_LOGIN_SUCCESS
          </option>
          <option value="ROLE_CHANGE">
            ROLE_CHANGE
          </option>
          <option value="ASSET_DELETE">
            ASSET_DELETE
          </option>
          <option value="APPROVAL_DECISION">
            APPROVAL_DECISION
          </option>
          <option value="RBAC_ACCESS_DENIED">
            RBAC_ACCESS_DENIED
          </option>
        </select>

        <!-- Status Filter -->
        <select
          v-model="selectedStatus"
          class="px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white focus:outline-none focus:border-emerald-400"
          @change="fetchAuditData"
        >
          <option value="ALL">
            All Statuses
          </option>
          <option value="SUCCESS">
            SUCCESS
          </option>
          <option value="FORBIDDEN">
            FORBIDDEN
          </option>
          <option value="FAILED">
            FAILED
          </option>
        </select>

        <!-- Service Filter -->
        <select
          v-model="selectedService"
          class="px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white focus:outline-none focus:border-emerald-400"
          @change="fetchAuditData"
        >
          <option value="ALL">
            All Microservices
          </option>
          <option value="auth">
            auth
          </option>
          <option value="gateway">
            gateway
          </option>
          <option value="asset">
            asset
          </option>
          <option value="workflow">
            workflow
          </option>
          <option value="helpdesk">
            helpdesk
          </option>
          <option value="reporting">
            reporting
          </option>
        </select>
      </div>

      <span class="text-xs text-slate-400 font-mono">
        Showing <strong>{{ auditLogs.length }}</strong> of {{ totalRecords }} events
      </span>
    </div>

    <!-- Audit Stream Table -->
    <div class="overflow-hidden rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-2xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase text-[10px] font-extrabold tracking-wider border-b border-slate-800">
            <tr>
              <th class="py-3.5 px-4">
                Audit ID
              </th>
              <th class="py-3.5 px-4">
                Action / Event
              </th>
              <th class="py-3.5 px-4">
                Actor (Role)
              </th>
              <th class="py-3.5 px-4">
                Service
              </th>
              <th class="py-3.5 px-4">
                IP Address
              </th>
              <th class="py-3.5 px-4">
                Timestamp
              </th>
              <th class="py-3.5 px-4 text-center">
                Status
              </th>
              <th class="py-3.5 px-4 text-right">
                Details & Diffs
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 font-medium">
            <tr
              v-for="log in auditLogs"
              :key="log.id"
              class="hover:bg-slate-800/40 transition-colors group"
            >
              <!-- Audit ID with Hash indicator -->
              <td class="py-3.5 px-4 font-mono font-bold text-indigo-400 flex items-center gap-1.5">
                <UIcon
                  name="i-lucide-lock"
                  class="w-3.5 h-3.5 text-emerald-400"
                />
                <span class="truncate max-w-[120px]">{{ log.id }}</span>
              </td>

              <!-- Event Type -->
              <td class="py-3.5 px-4">
                <span class="font-mono font-extrabold text-white group-hover:text-emerald-400 transition-colors">
                  {{ log.event_type }}
                </span>
                <p class="text-[10px] text-slate-400 font-mono">
                  {{ log.resource_type }}: {{ log.resource_id }}
                </p>
              </td>

              <!-- Actor (Role) -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2">
                  <div class="w-6 h-6 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center text-[10px] font-bold text-white">
                    {{ log.actor_name ? log.actor_name[0] : 'U' }}
                  </div>
                  <div>
                    <span class="text-slate-200 font-bold block">{{ log.actor_name || log.actor_email }}</span>
                    <span class="text-[9px] px-1.5 py-0.2 rounded font-mono font-semibold bg-slate-800 text-slate-400">
                      {{ log.actor_role }}
                    </span>
                  </div>
                </div>
              </td>

              <!-- Service -->
              <td class="py-3.5 px-4 font-mono font-semibold text-cyan-400">
                {{ log.service_name }}
              </td>

              <!-- IP Address -->
              <td class="py-3.5 px-4 font-mono text-slate-300">
                {{ log.ip_address }}
              </td>

              <!-- Timestamp -->
              <td class="py-3.5 px-4 text-slate-400 whitespace-nowrap">
                {{ new Date(log.created_at).toLocaleTimeString() }}
                <span class="text-[10px] text-slate-400 block font-mono">{{ new Date(log.created_at).toISOString().slice(0, 10) }}</span>
              </td>

              <!-- Status Badge -->
              <td class="py-3.5 px-4 text-center">
                <span
                  :class="[
                    'px-2.5 py-0.5 rounded-full text-[10px] font-extrabold tracking-wider',
                    log.status === 'SUCCESS' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30'
                    : log.status === 'FORBIDDEN' ? 'bg-rose-500/20 text-rose-300 border border-rose-500/30'
                      : 'bg-amber-500/20 text-amber-300 border border-amber-500/30'
                  ]"
                >
                  {{ log.status }}
                </span>
              </td>

              <!-- View Diff Button -->
              <td class="py-3.5 px-4 text-right">
                <button
                  class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 hover:text-white text-[11px] font-semibold border border-slate-700 transition-all inline-flex items-center gap-1.5"
                  @click="openDiffModal(log)"
                >
                  <UIcon
                    name="i-lucide-git-compare"
                    class="w-3.5 h-3.5 text-cyan-400"
                  />
                  <span>View Diffs</span>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Visual Code Diff & Tamper Proof Modal -->
    <UModal
      v-model="isDiffModalOpen"
      :ui="{ content: 'sm:max-w-2xl' }"
    >
      <div
        v-if="selectedAuditLog"
        class="p-6 space-y-5 bg-slate-900 border border-slate-800 text-white rounded-2xl"
      >
        <!-- Modal Header -->
        <div class="flex items-center justify-between border-b border-slate-800 pb-4">
          <div>
            <span class="text-[10px] font-mono uppercase tracking-wider text-emerald-400 font-bold">Audit Detail & Code Diffs</span>
            <h3 class="text-lg font-extrabold text-white flex items-center gap-2 mt-0.5">
              <span>{{ selectedAuditLog.event_type }}</span>
              <span class="text-xs px-2 py-0.5 rounded font-mono font-bold bg-slate-800 text-cyan-400">
                {{ selectedAuditLog.service_name }}
              </span>
            </h3>
          </div>
          <button
            class="text-slate-400 hover:text-white"
            @click="isDiffModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <!-- Meta Info Grid -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 p-3 rounded-xl bg-slate-950/80 border border-slate-800 text-xs">
          <div>
            <span class="text-[10px] text-slate-400 block font-semibold">Actor Email</span>
            <span class="font-mono font-bold text-white truncate block">{{ selectedAuditLog.actor_email }}</span>
          </div>
          <div>
            <span class="text-[10px] text-slate-400 block font-semibold">Actor Role</span>
            <span class="font-mono font-bold text-emerald-400">{{ selectedAuditLog.actor_role }}</span>
          </div>
          <div>
            <span class="text-[10px] text-slate-400 block font-semibold">IP Address</span>
            <span class="font-mono text-slate-300">{{ selectedAuditLog.ip_address }}</span>
          </div>
          <div>
            <span class="text-[10px] text-slate-400 block font-semibold">Target Resource</span>
            <span class="font-mono text-cyan-400 truncate block">{{ selectedAuditLog.resource_id }}</span>
          </div>
        </div>

        <!-- Code Diffs Viewer (Old vs New Values) -->
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs font-bold text-slate-300">
            <span>State Changes (Old vs New Values Diff)</span>
            <span class="text-[10px] text-cyan-400 flex items-center gap-1 font-semibold">
              <UIcon
                name="i-lucide-eye-off"
                class="w-3.5 h-3.5"
              />
              <span>Masked Sensitive Fields</span>
            </span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <!-- Old Values -->
            <div class="p-3.5 rounded-xl bg-slate-950 border border-rose-500/20 text-xs font-mono">
              <div class="text-[10px] font-bold text-rose-400 uppercase tracking-wider mb-2 flex items-center gap-1">
                <UIcon
                  name="i-lucide-minus-circle"
                  class="w-3 h-3"
                />
                <span>Old Values (Before Action)</span>
              </div>
              <pre class="text-slate-300 text-[11px] overflow-x-auto whitespace-pre-wrap">{{ JSON.stringify(selectedAuditLog.old_values || {}, null, 2) }}</pre>
            </div>

            <!-- New Values -->
            <div class="p-3.5 rounded-xl bg-slate-950 border border-emerald-500/20 text-xs font-mono">
              <div class="text-[10px] font-bold text-emerald-400 uppercase tracking-wider mb-2 flex items-center gap-1">
                <UIcon
                  name="i-lucide-plus-circle"
                  class="w-3 h-3"
                />
                <span>New Values (After Action)</span>
              </div>
              <pre class="text-slate-200 text-[11px] overflow-x-auto whitespace-pre-wrap">{{ JSON.stringify(selectedAuditLog.new_values || {}, null, 2) }}</pre>
            </div>
          </div>
        </div>

        <!-- Chained HMAC tamper-evident proof -->
        <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-1.5">
          <div class="flex items-center justify-between text-xs font-bold text-slate-300">
            <span class="flex items-center gap-1.5 text-emerald-400">
              <UIcon
                name="i-lucide-shield-check"
                class="w-4 h-4"
              />
              <span>{{ selectedAuditLog.checksum_algorithm || 'Integrity proof' }}</span>
            </span>
            <button
              class="text-[10px] text-cyan-400 hover:underline flex items-center gap-1"
              @click="copyChecksum(selectedAuditLog.checksum_sha256)"
            >
              <UIcon
                name="i-lucide-copy"
                class="w-3 h-3"
              />
              <span>Copy Hash</span>
            </button>
          </div>
          <p class="font-mono text-[10px] text-slate-400 break-all bg-slate-900 p-2 rounded border border-slate-800 select-all">
            {{ selectedAuditLog.checksum_sha256 }}
          </p>
        </div>
      </div>
    </UModal>
  </div>
</template>
