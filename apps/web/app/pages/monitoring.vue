<script setup lang="ts">
import type {
  ServiceHealthStatus,
  ClusterOverview,
  LogEntry
} from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()

// State
const overview = ref<ClusterOverview>({
  total_services: 11,
  online_services: 0,
  degraded_services: 0,
  offline_services: 0,
  unknown_services: 11,
  cluster_health_pct: 0,
  avg_probe_latency_ms: 0,
  metrics_available: false,
  data_source: 'active_health_probes',
  last_measurement_utc: ''
})

const services = ref<ServiceHealthStatus[]>([])
const logs = ref<LogEntry[]>([])
const loading = ref(false)
const probingServiceId = ref<string | null>(null)

// View State
const activeTab = ref<'grid' | 'logs' | 'red_metrics'>('grid')

// Live Log Streamer State
const selectedLogService = ref('all')
const selectedLogLevel = ref('all')
const logSearchQuery = ref('')
const isAutoScroll = ref(true)
const logContainerRef = ref<HTMLElement | null>(null)

// Auto-Refresh Timer
const isLivePolling = ref(true)
let pollTimer: ReturnType<typeof setInterval> | null = null

// Raw Prometheus Metrics Modal State
const isMetricsModalOpen = ref(false)
const selectedMetricService = ref<ServiceHealthStatus | null>(null)
const rawMetricsText = ref('')
const isLoadingMetrics = ref(false)

// Fetch Overview & Services
async function fetchData() {
  try {
    const [ovRes, srvRes, logRes] = await Promise.all([
      api.get<ClusterOverview>('/api/v1/monitoring/overview').catch(() => null),
      api.get<ServiceHealthStatus[]>('/api/v1/monitoring/services').catch(() => null),
      api.get<LogEntry[]>('/api/v1/monitoring/logs', {
        params: {
          service: selectedLogService.value === 'all' ? undefined : selectedLogService.value,
          level: selectedLogLevel.value === 'all' ? undefined : selectedLogLevel.value,
          limit: 50
        }
      }).catch(() => null)
    ])

    if (ovRes) overview.value = ovRes
    if (srvRes) services.value = srvRes

    if (logRes) logs.value = logRes
  } finally {
    loading.value = false
  }
}

// 1-Click Active Probe (Test Case 8.2)
async function handleProbeService(srv: ServiceHealthStatus) {
  probingServiceId.value = srv.id
  try {
    const updated = await api.post<ServiceHealthStatus>(`/api/v1/monitoring/probe/${srv.id}`)
    if (updated) {
      const idx = services.value.findIndex(s => s.id === srv.id)
      if (idx !== -1) services.value[idx] = updated
    }
    toast.add({
      title: 'Health Probe Complete',
      description: `${srv.name} is ${updated?.status || 'UNKNOWN'} (Latency: ${updated?.latency_ms ?? srv.latency_ms}ms).`,
      color: 'success'
    })
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Probe Error', description: errObj?.message || 'Health probe failed', color: 'error' })
  } finally {
    probingServiceId.value = null
  }
}

// Global 1-Click Probes
async function handleProbeAll() {
  loading.value = true
  try {
    await Promise.all(services.value.map(s => api.post(`/api/v1/monitoring/probe/${s.id}`).catch(() => null)))
    toast.add({ title: 'Cluster Probe Finished', description: 'All configured microservices were probed; review each observed status.', color: 'success' })
    fetchData()
  } finally {
    loading.value = false
  }
}

// View Raw Prometheus Metrics Modal (Test Case 8.1)
async function openMetricsModal(srv: ServiceHealthStatus) {
  selectedMetricService.value = srv
  isMetricsModalOpen.value = true
  isLoadingMetrics.value = true
  rawMetricsText.value = ''

  try {
    const res = await api.get<string>('/metrics', { responseType: 'text' }).catch(() => null)
    if (res && typeof res === 'string') {
      rawMetricsText.value = res
    } else {
      rawMetricsText.value = '# Metrics endpoint is unavailable. No synthetic metrics were generated.'
    }
  } finally {
    isLoadingMetrics.value = false
  }
}

// Copy Raw Metrics
function copyMetricsToClipboard() {
  if (!rawMetricsText.value) return
  navigator.clipboard.writeText(rawMetricsText.value)
  toast.add({ title: 'Copied', description: 'Prometheus metrics copied to clipboard.', color: 'success' })
}

// Toggle Live Polling
function toggleLivePolling() {
  isLivePolling.value = !isLivePolling.value
  if (isLivePolling.value) {
    pollTimer = setInterval(fetchData, 3000)
    toast.add({ title: 'Live Streaming Active', description: 'Polling cluster RED metrics every 3 seconds.', color: 'info' })
  } else {
    if (pollTimer) clearInterval(pollTimer)
    pollTimer = null
    toast.add({ title: 'Live Polling Paused', description: 'Metrics auto-refresh suspended.', color: 'warning' })
  }
}

// Filtered Logs
const filteredLogs = computed(() => {
  return logs.value.filter((l) => {
    const matchesService = selectedLogService.value === 'all' || l.service === selectedLogService.value
    const matchesLevel = selectedLogLevel.value === 'all' || l.level === selectedLogLevel.value
    const q = logSearchQuery.value.toLowerCase()
    const matchesQuery = !q || l.message.toLowerCase().includes(q) || l.service.toLowerCase().includes(q) || (l.caller && l.caller.toLowerCase().includes(q))
    return matchesService && matchesLevel && matchesQuery
  })
})

onMounted(() => {
  fetchData()
  pollTimer = setInterval(fetchData, 3000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-activity"
            class="w-7 h-7 text-emerald-400"
          />
          Enterprise Observability & SRE Health Mesh
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Active health probes for 11 services. RED metrics and logs are shown only when their real backends are connected.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <!-- Live Polling Toggle -->
        <button
          class="flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold border transition-all"
          :class="isLivePolling ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-slate-800 text-slate-400 border-slate-700'"
          @click="toggleLivePolling"
        >
          <span
            class="w-2 h-2 rounded-full"
            :class="isLivePolling ? 'bg-emerald-400 animate-ping' : 'bg-slate-500'"
          />
          <span>{{ isLivePolling ? 'Live Polling (3s)' : 'Polling Paused' }}</span>
        </button>

        <!-- 1-Click Global Health Probe -->
        <button
          :disabled="loading"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-cyan-600 hover:from-indigo-500 hover:to-cyan-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 hover:scale-105 transition-all disabled:opacity-50"
          @click="handleProbeAll"
        >
          <UIcon
            name="i-lucide-zap"
            class="w-4 h-4"
          />
          <span>Probe 11 Services</span>
        </button>
      </div>
    </div>

    <!-- 4 Top RED Metrics & Cluster Health KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-emerald-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Cluster Health</span>
          <UIcon
            name="i-lucide-shield-check"
            class="w-5 h-5 text-emerald-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2 flex items-center gap-2">
          <span>{{ overview.cluster_health_pct }}%</span>
          <span class="text-xs font-mono font-normal text-emerald-400">({{ overview.online_services }}/{{ overview.total_services }} Online)</span>
        </p>
        <span class="text-[10px] text-emerald-400 mt-1 block">Observed by active health probes</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-indigo-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Throughput (Rate)</span>
          <UIcon
            name="i-lucide-trending-up"
            class="w-5 h-5 text-indigo-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          Not connected
        </p>
        <span class="text-[10px] text-indigo-400 mt-1 block">Connect Prometheus query API for RED rate</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-cyan-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Avg Latency (Duration)</span>
          <UIcon
            name="i-lucide-clock"
            class="w-5 h-5 text-cyan-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ overview.avg_probe_latency_ms.toFixed(1) }} <span class="text-xs font-normal text-slate-400">ms mean</span>
        </p>
        <span class="text-[10px] text-cyan-400 mt-1 block">Latest active health probes</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-rose-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Cluster Error Rate</span>
          <UIcon
            name="i-lucide-alert-triangle"
            class="w-5 h-5 text-rose-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          Not connected
        </p>
        <span class="text-[10px] text-rose-400 mt-1 block">Connect Prometheus query API for RED errors</span>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex items-center justify-between gap-4 border-b border-slate-800 pb-3">
      <div class="flex items-center gap-1 p-1 bg-slate-900/80 border border-slate-800 rounded-xl w-fit">
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeTab === 'grid' ? 'bg-emerald-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeTab = 'grid'"
        >
          <UIcon
            name="i-lucide-grid"
            class="w-4 h-4"
          />
          <span>Ma Trận 11 Microservices (Service Grid)</span>
        </button>
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeTab === 'logs' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeTab = 'logs'"
        >
          <UIcon
            name="i-lucide-terminal"
            class="w-4 h-4"
          />
          <span>Live Log Streamer Console</span>
        </button>
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeTab === 'red_metrics' ? 'bg-cyan-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeTab = 'red_metrics'"
        >
          <UIcon
            name="i-lucide-bar-chart-2"
            class="w-4 h-4"
          />
          <span>Phân Tích Chỉ Số RED Method</span>
        </button>
      </div>
    </div>

    <!-- Tab 1: 11-Service Grid Matrix -->
    <div
      v-if="activeTab === 'grid'"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
    >
      <div
        v-for="srv in services"
        :key="srv.id"
        class="p-5 rounded-2xl bg-slate-900/70 border border-slate-800 backdrop-blur-xl hover:border-emerald-500/40 transition-all space-y-4 shadow-xl group"
      >
        <!-- Card Top -->
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-0.5">
            <div class="flex items-center gap-2">
              <span class="font-mono text-xs font-bold text-slate-400">:{{ srv.port }}</span>
              <span class="text-[10px] font-bold px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                {{ srv.category }}
              </span>
            </div>
            <h3 class="font-bold text-white text-sm group-hover:text-emerald-300 transition-colors">
              {{ srv.name }}
            </h3>
          </div>

          <!-- Status Indicator Badge -->
          <div
            class="flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold border"
            :class="{
              'bg-emerald-500/10 text-emerald-400 border-emerald-500/20': srv.status === 'ONLINE',
              'bg-amber-500/10 text-amber-400 border-amber-500/20': srv.status === 'DEGRADED',
              'bg-rose-500/10 text-rose-400 border-rose-500/20': srv.status === 'OFFLINE',
              'bg-slate-500/10 text-slate-400 border-slate-500/20': srv.status === 'UNKNOWN'
            }"
          >
            <span
              class="w-2 h-2 rounded-full bg-current"
              :class="{ 'animate-ping': srv.status === 'ONLINE' }"
            />
            <span>{{ srv.status }}</span>
          </div>
        </div>

        <!-- Metrics 2x2 Grid -->
        <div class="grid grid-cols-2 gap-2 text-xs">
          <div class="p-2.5 rounded-xl bg-slate-950/80 border border-slate-800">
            <span class="text-[10px] text-slate-500 font-semibold block">Uptime</span>
            <span class="font-mono font-bold text-white">{{ srv.uptime_pct == null ? '—' : `${srv.uptime_pct}%` }}</span>
          </div>
          <div class="p-2.5 rounded-xl bg-slate-950/80 border border-slate-800">
            <span class="text-[10px] text-slate-500 font-semibold block">Probe latency</span>
            <span class="font-mono font-bold text-cyan-300">{{ srv.latency_ms.toFixed(1) }} ms</span>
          </div>
          <div class="p-2.5 rounded-xl bg-slate-950/80 border border-slate-800">
            <span class="text-[10px] text-slate-500 font-semibold block">CPU / RAM</span>
            <span class="font-mono font-bold text-slate-200">{{ srv.cpu_pct == null || srv.memory_mb == null ? '—' : `${srv.cpu_pct}% / ${srv.memory_mb.toFixed(0)}MB` }}</span>
          </div>
          <div class="p-2.5 rounded-xl bg-slate-950/80 border border-slate-800">
            <span class="text-[10px] text-slate-500 font-semibold block">Error Rate</span>
            <span class="font-mono font-bold text-rose-300">{{ srv.error_rate_pct == null ? '—' : `${srv.error_rate_pct}%` }}</span>
          </div>
        </div>

        <!-- Card Bottom Actions -->
        <div class="flex items-center justify-between pt-2 border-t border-slate-800 text-xs">
          <button
            class="text-indigo-400 hover:text-indigo-300 font-semibold text-[11px] flex items-center gap-1"
            @click="openMetricsModal(srv)"
          >
            <UIcon
              name="i-lucide-code"
              class="w-3.5 h-3.5"
            />
            <span>Prometheus /metrics</span>
          </button>

          <button
            :disabled="probingServiceId === srv.id"
            class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-[11px] font-semibold flex items-center gap-1 transition-all"
            @click="handleProbeService(srv)"
          >
            <UIcon
              name="i-lucide-refresh-cw"
              class="w-3 h-3"
              :class="{ 'animate-spin': probingServiceId === srv.id }"
            />
            <span>Probe</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Tab 2: Live Log Streamer Terminal Console -->
    <div
      v-if="activeTab === 'logs'"
      class="p-6 rounded-3xl bg-slate-950 border border-slate-800 shadow-2xl space-y-4 font-mono text-xs"
    >
      <!-- Console Header & Filters -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 border-b border-slate-800 pb-3">
        <div class="flex items-center gap-2">
          <span class="w-3 h-3 rounded-full bg-rose-500/80 inline-block" />
          <span class="w-3 h-3 rounded-full bg-amber-500/80 inline-block" />
          <span class="w-3 h-3 rounded-full bg-emerald-500/80 inline-block" />
          <span class="text-slate-400 font-bold text-xs ml-2">Log query backend (not configured)</span>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs">
          <!-- Service Filter -->
          <select
            v-model="selectedLogService"
            class="px-2.5 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-slate-300 text-[11px] focus:outline-none focus:border-indigo-500"
            @change="fetchData"
          >
            <option value="all">
              Service: All (11)
            </option>
            <option value="gateway">
              gateway
            </option>
            <option value="auth">
              auth
            </option>
            <option value="helpdesk">
              helpdesk
            </option>
            <option value="workflow">
              workflow
            </option>
            <option value="ai">
              ai
            </option>
            <option value="knowledge">
              knowledge
            </option>
            <option value="asset">
              asset
            </option>
            <option value="notification">
              notification
            </option>
          </select>

          <!-- Level Filter -->
          <select
            v-model="selectedLogLevel"
            class="px-2.5 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-slate-300 text-[11px] focus:outline-none focus:border-indigo-500"
            @change="fetchData"
          >
            <option value="all">
              Level: All
            </option>
            <option value="INFO">
              INFO
            </option>
            <option value="WARN">
              WARN
            </option>
            <option value="ERROR">
              ERROR
            </option>
          </select>

          <!-- Search Filter -->
          <input
            v-model="logSearchQuery"
            type="text"
            placeholder="Search log messages..."
            class="px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-white placeholder-slate-500 text-[11px] focus:outline-none focus:border-indigo-500 w-44"
          >

          <!-- Auto Scroll Toggle -->
          <button
            class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold border transition-all"
            :class="isAutoScroll ? 'bg-indigo-600/20 text-indigo-300 border-indigo-500/30' : 'bg-slate-900 text-slate-400 border-slate-800'"
            @click="isAutoScroll = !isAutoScroll"
          >
            Auto-Scroll: {{ isAutoScroll ? 'ON' : 'OFF' }}
          </button>
        </div>
      </div>

      <!-- Logs Terminal Body -->
      <div
        ref="logContainerRef"
        class="h-96 overflow-y-auto space-y-1.5 p-3 rounded-xl bg-black/80 border border-slate-900 text-[11px] leading-relaxed select-text"
      >
        <div
          v-if="filteredLogs.length === 0"
          class="text-slate-600 text-center py-12"
        >
          Live logs are unavailable until a Loki-compatible query backend is configured.
        </div>

        <div
          v-for="l in filteredLogs"
          :key="l.id"
          class="flex items-start gap-2 hover:bg-slate-900/60 p-1 rounded transition-colors group"
        >
          <span class="text-slate-500 shrink-0 font-mono text-[10px]">[{{ l.timestamp }}]</span>
          <span
            class="px-1.5 py-0.2 rounded text-[9px] font-bold shrink-0"
            :class="{
              'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30': l.level === 'INFO',
              'bg-amber-500/20 text-amber-400 border border-amber-500/30': l.level === 'WARN',
              'bg-rose-500/20 text-rose-400 border border-rose-500/30': l.level === 'ERROR' || l.level === 'FATAL'
            }"
          >
            {{ l.level }}
          </span>
          <span class="font-bold text-indigo-400 shrink-0">[{{ l.service }}]</span>
          <span class="text-slate-200 flex-1 break-all">{{ l.message }}</span>
          <span
            v-if="l.caller"
            class="text-[9px] text-slate-600 shrink-0 font-mono hidden md:inline"
          >{{ l.caller }}</span>
        </div>
      </div>
    </div>

    <!-- Tab 3: RED Method Metrics Analyzer -->
    <div
      v-if="activeTab === 'red_metrics'"
      class="p-6 rounded-3xl bg-slate-900/70 border border-slate-800 backdrop-blur-xl space-y-5 shadow-2xl text-xs"
    >
      <div class="border-b border-slate-800 pb-3">
        <h2 class="text-base font-bold text-white flex items-center gap-2">
          <UIcon
            name="i-lucide-bar-chart-2"
            class="w-5 h-5 text-cyan-400"
          />
          RED Method Architecture Breakdown (Rate, Errors, Duration)
        </h2>
        <p class="text-xs text-slate-400 mt-0.5">
          Standardized microservice operational indicators recommended by Google SRE and Tom Wilkie.
        </p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <!-- Rate Tier -->
        <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between text-indigo-400 font-bold">
            <span class="flex items-center gap-1.5">
              <UIcon
                name="i-lucide-trending-up"
                class="w-4 h-4"
              />
              RATE (Throughput)
            </span>
            <span>Not connected</span>
          </div>
          <p class="text-slate-400 leading-relaxed text-[11px]">
            The number of requests per second that our application cluster is serving across all ingress ports.
          </p>
          <div class="space-y-1.5 pt-2 border-t border-slate-800 font-mono text-[10px]">
            <div class="flex justify-between text-slate-300">
              <span>Core Edge Gateway:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>Helpdesk Service:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>PostgreSQL Queries:</span><span class="text-slate-500">Unavailable</span>
            </div>
          </div>
        </div>

        <!-- Errors Tier -->
        <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between text-rose-400 font-bold">
            <span class="flex items-center gap-1.5">
              <UIcon
                name="i-lucide-alert-triangle"
                class="w-4 h-4"
              />
              ERRORS (Failures)
            </span>
            <span>Not connected</span>
          </div>
          <p class="text-slate-400 leading-relaxed text-[11px]">
            The number of failed requests (HTTP 5xx & unexpected timeouts) expressed as an overall percentage of traffic.
          </p>
          <div class="space-y-1.5 pt-2 border-t border-slate-800 font-mono text-[10px]">
            <div class="flex justify-between text-slate-300">
              <span>Gateway HTTP 500:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>AI Retriever Retries:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>Circuit Breaker Trips:</span><span class="text-slate-500">Unavailable</span>
            </div>
          </div>
        </div>

        <!-- Duration Tier -->
        <div class="p-4 rounded-2xl bg-slate-950 border border-slate-800 space-y-3">
          <div class="flex items-center justify-between text-cyan-400 font-bold">
            <span class="flex items-center gap-1.5">
              <UIcon
                name="i-lucide-clock"
                class="w-4 h-4"
              />
              DURATION (Latency)
            </span>
            <span>{{ overview.avg_probe_latency_ms.toFixed(1) }} ms probe mean</span>
          </div>
          <p class="text-slate-400 leading-relaxed text-[11px]">
            This is active health-probe latency, not an application request percentile. Connect Prometheus to obtain p50/p95/p99.
          </p>
          <div class="space-y-1.5 pt-2 border-t border-slate-800 font-mono text-[10px]">
            <div class="flex justify-between text-slate-300">
              <span>Gateway request latency:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>PostgreSQL query latency:</span><span class="text-slate-500">Unavailable</span>
            </div>
            <div class="flex justify-between text-slate-300">
              <span>Qdrant query latency:</span><span class="text-slate-500">Unavailable</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Prometheus Metrics Raw Inspector Modal (Test Case 8.1) -->
    <div
      v-if="isMetricsModalOpen && selectedMetricService"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-3xl max-h-[85vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-4 text-white shadow-2xl font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <div class="flex items-center gap-2">
            <UIcon
              name="i-lucide-activity"
              class="w-5 h-5 text-emerald-400"
            />
            <span class="font-bold text-sm">Gateway Prometheus text exposition</span>
          </div>
          <button
            class="text-slate-400 hover:text-white"
            @click="isMetricsModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="p-4 rounded-2xl bg-black/90 border border-slate-800 text-[11px] leading-relaxed overflow-x-auto text-emerald-400 whitespace-pre max-h-[50vh]">
          {{ rawMetricsText }}
        </div>

        <div class="flex items-center justify-between pt-2 border-t border-slate-800">
          <span class="text-[10px] text-slate-500">Standard: text/plain; version=0.0.4; charset=utf-8</span>
          <div class="flex items-center gap-2">
            <button
              class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 text-xs font-semibold"
              @click="isMetricsModalOpen = false"
            >
              Close
            </button>
            <button
              class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow flex items-center gap-1.5"
              @click="copyMetricsToClipboard"
            >
              <UIcon
                name="i-lucide-copy"
                class="w-3.5 h-3.5"
              />
              <span>Copy Prometheus Output</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
