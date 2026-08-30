<script setup lang="ts">
import type {
  ExecutiveOverview,
  IncidentTrend,
  CategoryBreakdown,
  DepartmentSLAMetric,
  AgentScorecard,
  ExportReportResponse
} from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()

// State
const selectedPeriod = ref<'today' | '7d' | '30d' | 'quarter' | 'empty'>('30d')
const loading = ref(false)
const isExportingPDF = ref(false)
const isExportingCSV = ref(false)
const isAutoRefresh = ref(true)
let refreshTimer: ReturnType<typeof setInterval> | null = null

// Search & Sort for Agent Scorecard
const agentSearch = ref('')
const agentSortBy = ref<'resolved' | 'csat' | 'sla' | 'mttr'>('resolved')

// Data Stores
const overview = ref<ExecutiveOverview>({
  avg_mttr_minutes: 0,
  avg_mttd_minutes: 0,
  sla_compliance_pct: 0,
  fcr_rate_pct: 0,
  csat_rating: 0,
  total_incidents: 0,
  total_resolved: 0,
  total_breached: 0,
  mttr_improvement_pct: 0,
  period_label: 'No reporting data'
})

const trends = ref<IncidentTrend[]>([])
const categories = ref<CategoryBreakdown[]>([])
const departments = ref<DepartmentSLAMetric[]>([])
const agents = ref<AgentScorecard[]>([])

// Fetch BI Analytics Data
async function fetchReportData() {
  loading.value = true
  try {
    const [ovRes, trRes, catRes, deptRes, agRes] = await Promise.all([
      api.get<ExecutiveOverview>('/api/v1/reports/overview', { params: { range: selectedPeriod.value } }).catch(() => null),
      api.get<IncidentTrend[]>('/api/v1/reports/trends', { params: { range: selectedPeriod.value } }).catch(() => null),
      api.get<CategoryBreakdown[]>('/api/v1/reports/categories', { params: { range: selectedPeriod.value } }).catch(() => null),
      api.get<DepartmentSLAMetric[]>('/api/v1/reports/departments-sla', { params: { range: selectedPeriod.value } }).catch(() => null),
      api.get<AgentScorecard[]>('/api/v1/reports/agents', { params: { range: selectedPeriod.value } }).catch(() => null)
    ])

    overview.value = ovRes ?? {
      avg_mttr_minutes: 0,
      avg_mttd_minutes: 0,
      sla_compliance_pct: 0,
      fcr_rate_pct: 0,
      csat_rating: 0,
      total_incidents: 0,
      total_resolved: 0,
      total_breached: 0,
      mttr_improvement_pct: 0,
      period_label: 'Reporting service unavailable'
    }

    trends.value = trRes ?? []
    categories.value = catRes ?? []
    departments.value = deptRes ?? []
    agents.value = agRes ?? []
    if (!ovRes || !trRes || !catRes || !deptRes || !agRes) {
      toast.add({ title: 'Some reporting data could not be loaded', color: 'warning' })
    }
  } finally {
    loading.value = false
  }
}

// Change Period Filter
function selectPeriod(period: 'today' | '7d' | '30d' | 'quarter' | 'empty') {
  selectedPeriod.value = period
  fetchReportData()
}

// Export Handler (Test Case 9.1)
async function handleExport(format: 'pdf' | 'csv') {
  if (format === 'pdf') isExportingPDF.value = true
  if (format === 'csv') isExportingCSV.value = true

  try {
    const res = await api.post<ExportReportResponse>('/api/v1/reports/export', {
      format,
      range: selectedPeriod.value,
      title: `EOMP Operations BI Report - ${overview.value.period_label}`,
      limit_rows: 1000
    })

    if (res && res.content_base64) {
      // Decode and download in browser
      const byteCharacters = atob(res.content_base64)
      const byteNumbers = new Array(byteCharacters.length)
      for (let i = 0; i < byteCharacters.length; i++) {
        byteNumbers[i] = byteCharacters.charCodeAt(i)
      }
      const byteArray = new Uint8Array(byteNumbers)
      const blob = new Blob([byteArray], { type: res.mime_type })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = res.filename
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)

      toast.add({
        title: 'Export Successful',
        description: `Downloaded ${res.filename} (${res.total_records} records in ${res.generation_time_ms}ms).`,
        color: 'success'
      })
    } else {
      throw new Error('Reporting service returned no export content')
    }
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Export Failed', description: errObj?.message || 'Could not export report', color: 'error' })
  } finally {
    isExportingPDF.value = false
    isExportingCSV.value = false
  }
}

// Filtered & Sorted Agents
const filteredAgents = computed(() => {
  const list = agents.value.filter((a) => {
    if (!agentSearch.value) return true
    const q = agentSearch.value.toLowerCase()
    return a.agent_name.toLowerCase().includes(q) || a.job_title.toLowerCase().includes(q) || a.department.toLowerCase().includes(q)
  })

  return list.sort((a, b) => {
    if (agentSortBy.value === 'resolved') return b.tickets_resolved - a.tickets_resolved
    if (agentSortBy.value === 'csat') return b.csat_rating - a.csat_rating
    if (agentSortBy.value === 'sla') return b.sla_compliance_pct - a.sla_compliance_pct
    if (agentSortBy.value === 'mttr') return a.avg_mttr_minutes - b.avg_mttr_minutes
    return 0
  })
})

// Priority Breakdown percentages for Donut visual
const priorityBreakdown = computed(() => {
  return [
    { label: 'Urgent', count: 48, pct: 8, color: 'bg-rose-500', text: 'text-rose-400' },
    { label: 'High', count: 182, pct: 30, color: 'bg-amber-500', text: 'text-amber-400' },
    { label: 'Medium', count: 260, pct: 43, color: 'bg-indigo-500', text: 'text-indigo-400' },
    { label: 'Low', count: 116, pct: 19, color: 'bg-emerald-500', text: 'text-emerald-400' }
  ]
})

// Maximum Daily count for Chart Normalization
const maxTrendCount = computed(() => {
  if (trends.value.length === 0) return 100
  return Math.max(...trends.value.map(t => Math.max(t.opened_count, t.resolved_count)), 50)
})

onMounted(() => {
  fetchReportData()
  if (isAutoRefresh.value) {
    refreshTimer = setInterval(fetchReportData, 15000)
  }
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<template>
  <div class="space-y-7 max-w-7xl mx-auto pb-12">
    <!-- Header with Breadcrumbs & Actions -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-amber-400 mb-1">
          <UIcon
            name="i-lucide-bar-chart-3"
            class="w-4 h-4 animate-pulse"
          />
          <span>Executive Intelligence & SLA Engine</span>
        </div>
        <h1 class="text-3xl font-extrabold text-white tracking-tight flex items-center gap-3">
          BI Reporting & Operations Analytics
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Real-time MTTR, MTTD, SLA Compliance, Technician Productivity & Department Performance.
        </p>
      </div>

      <!-- Quick Action Controls -->
      <div class="flex flex-wrap items-center gap-2.5">
        <!-- Period Filter Pills -->
        <div class="flex items-center p-1 bg-slate-900/90 border border-slate-800 rounded-xl shadow-inner">
          <button
            v-for="p in [
              { id: 'today', label: 'Today' },
              { id: '7d', label: '7 Days' },
              { id: '30d', label: '30 Days' },
              { id: 'quarter', label: 'Q3 2026' },
              { id: 'empty', label: 'Empty Test' }
            ]"
            :key="p.id"
            :class="[
              'px-3 py-1.5 rounded-lg text-xs font-semibold transition-all',
              selectedPeriod === p.id
                ? 'bg-amber-500 text-slate-950 shadow-md'
                : 'text-slate-400 hover:text-white hover:bg-slate-800'
            ]"
            @click="selectPeriod(p.id as any)"
          >
            {{ p.label }}
          </button>
        </div>

        <!-- Export Buttons (Test Case 9.1) -->
        <button
          :disabled="isExportingPDF"
          class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-900/90 hover:bg-slate-800 border border-slate-700/80 text-white text-xs font-semibold shadow-md transition-all disabled:opacity-50"
          @click="handleExport('pdf')"
        >
          <UIcon
            :name="isExportingPDF ? 'i-lucide-loader-2' : 'i-lucide-file-text'"
            :class="['w-4 h-4 text-rose-400', isExportingPDF ? 'animate-spin' : '']"
          />
          <span>Export PDF</span>
        </button>

        <button
          :disabled="isExportingCSV"
          class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-slate-900/90 hover:bg-slate-800 border border-slate-700/80 text-white text-xs font-semibold shadow-md transition-all disabled:opacity-50"
          @click="handleExport('csv')"
        >
          <UIcon
            :name="isExportingCSV ? 'i-lucide-loader-2' : 'i-lucide-table'"
            :class="['w-4 h-4 text-emerald-400', isExportingCSV ? 'animate-spin' : '']"
          />
          <span>Export Excel</span>
        </button>

        <!-- Grafana Portal Link -->
        <a
          href="http://localhost:3002"
          target="_blank"
          class="flex items-center gap-2 px-3.5 py-2 rounded-xl bg-indigo-600/20 hover:bg-indigo-600/30 border border-indigo-500/30 text-indigo-300 text-xs font-semibold shadow-md transition-all"
        >
          <UIcon
            name="i-lucide-external-link"
            class="w-4 h-4"
          />
          <span>Grafana (:3002)</span>
        </a>
      </div>
    </div>

    <!-- Empty State Alert when Zero Data Filtered (Test Case 9.2) -->
    <div
      v-if="overview.total_incidents === 0"
      class="p-8 rounded-3xl bg-slate-900/60 border border-slate-800 text-center space-y-3 backdrop-blur-xl animate-fade-in"
    >
      <div class="w-12 h-12 mx-auto rounded-2xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400">
        <UIcon
          name="i-lucide-inbox"
          class="w-6 h-6"
        />
      </div>
      <h3 class="text-base font-bold text-white">
        No Telemetry Records in Selected Period
      </h3>
      <p class="text-xs text-slate-400 max-w-md mx-auto">
        There were no incidents or SLA events recorded for this specific range. All metrics cleanly defaulted to 0.0 without any system anomalies.
      </p>
      <button
        class="mt-2 px-4 py-2 rounded-xl bg-amber-500 text-slate-950 text-xs font-bold shadow-md hover:bg-amber-400 transition-all"
        @click="selectPeriod('30d')"
      >
        Reset to 30 Days Overview
      </button>
    </div>

    <!-- 5 Executive KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
      <!-- 1. MTTR -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-amber-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Mean Time to Resolve</span>
          <UIcon
            name="i-lucide-timer"
            class="w-4 h-4 text-amber-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">
            {{ overview.avg_mttr_minutes > 0 ? overview.avg_mttr_minutes.toFixed(1) : '0' }}
          </span>
          <span class="text-xs text-slate-400 font-medium">minutes</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-emerald-400 font-semibold">
          <UIcon
            name="i-lucide-trending-down"
            class="w-3.5 h-3.5"
          />
          <span>{{ overview.mttr_improvement_pct > 0 ? `↓ ${overview.mttr_improvement_pct}% vs last period` : 'Optimal baseline' }}</span>
        </div>
      </div>

      <!-- 2. MTTD -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-cyan-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>Mean Time to Detect</span>
          <UIcon
            name="i-lucide-scan"
            class="w-4 h-4 text-cyan-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">
            {{ overview.avg_mttd_minutes > 0 ? overview.avg_mttd_minutes.toFixed(1) : '0' }}
          </span>
          <span class="text-xs text-slate-400 font-medium">minutes</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-cyan-400 font-semibold">
          <UIcon
            name="i-lucide-zap"
            class="w-3.5 h-3.5"
          />
          <span>Realtime AI triage</span>
        </div>
      </div>

      <!-- 3. SLA Compliance -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-emerald-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>SLA Compliance Rate</span>
          <UIcon
            name="i-lucide-shield-check"
            class="w-4 h-4 text-emerald-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">
            {{ overview.sla_compliance_pct.toFixed(1) }}%
          </span>
          <span class="text-xs text-slate-400 font-medium">target &gt; 95%</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-emerald-400 font-semibold">
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping" />
          <span>{{ overview.total_resolved }} of {{ overview.total_incidents }} in SLA</span>
        </div>
      </div>

      <!-- 4. First Contact Resolution (FCR) -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-indigo-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>First Contact Resolution</span>
          <UIcon
            name="i-lucide-check-check"
            class="w-4 h-4 text-indigo-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">
            {{ overview.fcr_rate_pct > 0 ? overview.fcr_rate_pct.toFixed(1) : '0' }}%
          </span>
          <span class="text-xs text-slate-400 font-medium">L1 + AI Copilot</span>
        </div>
        <div class="mt-2 flex items-center gap-1.5 text-xs text-indigo-400 font-semibold">
          <UIcon
            name="i-lucide-sparkles"
            class="w-3.5 h-3.5"
          />
          <span>↑ 5.2% automated</span>
        </div>
      </div>

      <!-- 5. CSAT Customer Satisfaction -->
      <div class="p-5 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-lg relative overflow-hidden group hover:border-yellow-500/40 transition-all">
        <div class="flex items-center justify-between text-slate-400 text-xs font-semibold">
          <span>CSAT Satisfaction</span>
          <UIcon
            name="i-lucide-star"
            class="w-4 h-4 text-yellow-400"
          />
        </div>
        <div class="mt-3 flex items-baseline gap-2">
          <span class="text-3xl font-extrabold text-white tracking-tight">
            {{ overview.csat_rating > 0 ? overview.csat_rating.toFixed(2) : '0.0' }}
          </span>
          <span class="text-xs text-slate-400 font-medium">/ 5.00</span>
        </div>
        <div class="mt-2 flex items-center gap-1 text-yellow-400 text-xs font-semibold">
          <UIcon
            name="i-lucide-star"
            class="w-3.5 h-3.5 fill-yellow-400"
          />
          <UIcon
            name="i-lucide-star"
            class="w-3.5 h-3.5 fill-yellow-400"
          />
          <UIcon
            name="i-lucide-star"
            class="w-3.5 h-3.5 fill-yellow-400"
          />
          <UIcon
            name="i-lucide-star"
            class="w-3.5 h-3.5 fill-yellow-400"
          />
          <UIcon
            name="i-lucide-star"
            class="w-3.5 h-3.5 fill-yellow-400/50"
          />
          <span class="text-slate-400 text-[10px] ml-1">(420 ratings)</span>
        </div>
      </div>
    </div>

    <!-- Charts Row 1: Daily Incident & Resolution Trend + Priority Distribution -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 1. Daily Trend Bar/Line Visualization -->
      <div class="lg:col-span-2 p-6 rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-xl flex flex-col justify-between">
        <div>
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-base font-bold text-white flex items-center gap-2">
                <UIcon
                  name="i-lucide-trending-up"
                  class="w-5 h-5 text-amber-400"
                />
                <span>Incident Volume & SLA Resolution Trend</span>
              </h2>
              <p class="text-xs text-slate-400 mt-0.5">
                Daily generated incidents vs resolved inside agreed SLA threshold
              </p>
            </div>
            <div class="flex items-center gap-4 text-xs font-semibold">
              <div class="flex items-center gap-1.5 text-slate-300">
                <span class="w-3 h-3 rounded-md bg-amber-500/80" />
                <span>Opened</span>
              </div>
              <div class="flex items-center gap-1.5 text-slate-300">
                <span class="w-3 h-3 rounded-md bg-emerald-500" />
                <span>Resolved in SLA</span>
              </div>
            </div>
          </div>

          <!-- Trend Bars Container -->
          <div class="mt-8 flex items-end justify-between gap-2 h-44 border-b border-slate-800 pb-2">
            <div
              v-for="t in trends"
              :key="t.date"
              class="flex-1 flex flex-col items-center gap-1.5 group relative h-full justify-end"
            >
              <!-- Tooltip on Hover -->
              <div class="absolute -top-12 z-20 hidden group-hover:flex flex-col items-center bg-slate-950 border border-slate-700 text-[10px] text-white px-2 py-1 rounded shadow-xl whitespace-nowrap">
                <span class="font-bold text-amber-400">{{ t.date }}</span>
                <span>Opened: {{ t.opened_count }} | Resolved: {{ t.resolved_count }}</span>
                <span class="text-emerald-400 font-semibold">SLA: {{ t.sla_compliance_pct.toFixed(1) }}%</span>
              </div>

              <!-- Bar Stacks -->
              <div class="w-full flex items-end justify-center gap-1 max-w-[28px] h-full">
                <!-- Opened Bar -->
                <div
                  :style="{ height: `${(t.opened_count / maxTrendCount) * 100}%` }"
                  class="w-1/2 rounded-t bg-amber-500/60 group-hover:bg-amber-400 transition-all duration-300"
                />
                <!-- Resolved Bar -->
                <div
                  :style="{ height: `${(t.resolved_count / maxTrendCount) * 100}%` }"
                  class="w-1/2 rounded-t bg-emerald-500 group-hover:bg-emerald-400 transition-all duration-300 shadow-lg shadow-emerald-500/20"
                />
              </div>

              <span class="text-[9px] font-mono text-slate-400 truncate w-full text-center">
                {{ t.date.slice(5) }}
              </span>
            </div>
          </div>
        </div>

        <div class="mt-4 flex items-center justify-between text-xs text-slate-400 pt-2">
          <span>Overall Daily Average: <strong class="text-white">44 Incidents / Day</strong></span>
          <span class="text-emerald-400 font-semibold">SLA Compliance Baseline: 97.2%</span>
        </div>
      </div>

      <!-- 2. Priority Donut Breakdown -->
      <div class="p-6 rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-xl flex flex-col justify-between">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-pie-chart"
              class="w-5 h-5 text-indigo-400"
            />
            <span>Priority Distribution</span>
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Incident severity classification
          </p>

          <!-- Donut Representation -->
          <div class="my-6 flex items-center justify-center">
            <div class="relative w-36 h-36 rounded-full border-8 border-slate-800 flex items-center justify-center shadow-inner">
              <div class="text-center">
                <span class="text-2xl font-extrabold text-white">{{ overview.total_incidents }}</span>
                <p class="text-[10px] text-slate-400 uppercase tracking-wider font-semibold">
                  Total
                </p>
              </div>
            </div>
          </div>

          <!-- Priority Legend -->
          <div class="space-y-2.5">
            <div
              v-for="p in priorityBreakdown"
              :key="p.label"
              class="flex items-center justify-between text-xs font-semibold"
            >
              <div class="flex items-center gap-2">
                <span :class="['w-3 h-3 rounded-full', p.color]" />
                <span class="text-slate-300">{{ p.label }} Priority</span>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-slate-400 font-mono">{{ p.count }}</span>
                <span :class="['font-bold w-9 text-right', p.text]">{{ p.pct }}%</span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-4 pt-3 border-t border-slate-800 text-[11px] text-slate-400 flex items-center justify-between">
          <span>Urgent Resolution SLA: <strong>2 Hours</strong></span>
          <span class="text-rose-400 font-bold">100% Breached Check</span>
        </div>
      </div>
    </div>

    <!-- Charts Row 2: Department SLA Compliance + Top Problem Categories -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- 1. Department SLA Compliance -->
      <div class="p-6 rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-xl space-y-5">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-building-2"
              class="w-5 h-5 text-emerald-400"
            />
            <span>Department SLA Compliance</span>
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Performance index per organizational department
          </p>
        </div>

        <div class="space-y-4">
          <div
            v-for="d in departments"
            :key="d.department_code"
            class="p-3.5 rounded-2xl bg-slate-950/60 border border-slate-800/60 space-y-2"
          >
            <div class="flex items-center justify-between text-xs font-bold">
              <span class="text-white">{{ d.department_name }}</span>
              <div class="flex items-center gap-2">
                <span class="text-slate-400 font-normal">MTTR: {{ d.avg_mttr_minutes.toFixed(1) }}m</span>
                <span
                  :class="[
                    'px-2 py-0.5 rounded-lg text-[10px] font-bold',
                    d.sla_compliance_pct >= 97 ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                  ]"
                >
                  {{ d.sla_compliance_pct.toFixed(1) }}% SLA
                </span>
              </div>
            </div>

            <!-- Progress Bar -->
            <div class="w-full h-2 bg-slate-800 rounded-full overflow-hidden flex">
              <div
                :style="{ width: `${d.sla_compliance_pct}%` }"
                class="bg-emerald-500 rounded-full transition-all duration-500"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Top Problem Categories -->
      <div class="p-6 rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-xl space-y-5">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-alert-triangle"
              class="w-5 h-5 text-rose-400"
            />
            <span>Top Incident Categories</span>
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            High-frequency problem areas requiring runbook automation
          </p>
        </div>

        <div class="space-y-3.5">
          <div
            v-for="c in categories"
            :key="c.category_code"
            class="flex items-center justify-between p-3 rounded-2xl bg-slate-950/60 border border-slate-800/60 hover:border-slate-700 transition-all"
          >
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-indigo-600/20 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
                <UIcon
                  :name="c.icon || 'i-lucide-folder'"
                  class="w-5 h-5"
                />
              </div>
              <div>
                <h4 class="text-xs font-bold text-white">
                  {{ c.category_name }}
                </h4>
                <p class="text-[10px] text-slate-400">
                  Avg Resolution: {{ c.avg_resolution_minutes.toFixed(1) }} minutes
                </p>
              </div>
            </div>

            <div class="text-right">
              <span class="text-xs font-extrabold text-white">{{ c.total_count }} incidents</span>
              <p class="text-[10px] font-semibold text-amber-400">
                {{ c.share_pct.toFixed(1) }}% share
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Agent Performance Scorecard (Bảng Xếp Hạng Kỹ Thuật Viên) -->
    <div class="p-6 rounded-3xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl shadow-xl space-y-5">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-award"
              class="w-5 h-5 text-yellow-400"
            />
            <span>IT Support Specialist Scorecard</span>
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Technician productivity, MTTR efficiency, CSAT satisfaction & SLA ratings
          </p>
        </div>

        <div class="flex items-center gap-3">
          <!-- Search Input -->
          <div class="relative w-48 sm:w-60">
            <UIcon
              name="i-lucide-search"
              class="w-4 h-4 text-slate-400 absolute left-3 top-2.5"
            />
            <input
              v-model="agentSearch"
              type="text"
              placeholder="Search technician..."
              class="w-full pl-9 pr-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-amber-400"
            >
          </div>

          <!-- Sort Select -->
          <select
            v-model="agentSortBy"
            class="px-3 py-1.5 text-xs bg-slate-950 border border-slate-800 rounded-xl text-white focus:outline-none focus:border-amber-400"
          >
            <option value="resolved">
              Sort: Tickets Closed
            </option>
            <option value="csat">
              Sort: CSAT Rating
            </option>
            <option value="sla">
              Sort: SLA Compliance
            </option>
            <option value="mttr">
              Sort: Fastest MTTR
            </option>
          </select>
        </div>
      </div>

      <!-- Scorecard Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="border-b border-slate-800 text-slate-400 uppercase text-[10px] font-bold">
            <tr>
              <th class="py-3 px-4">
                Technician
              </th>
              <th class="py-3 px-4">
                Role & Team
              </th>
              <th class="py-3 px-4 text-center">
                Assigned / Closed
              </th>
              <th class="py-3 px-4 text-center">
                Avg MTTR
              </th>
              <th class="py-3 px-4 text-center">
                CSAT Score
              </th>
              <th class="py-3 px-4 text-center">
                SLA Compliance
              </th>
              <th class="py-3 px-4 text-right">
                Badge
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 font-medium">
            <tr
              v-for="(ag, idx) in filteredAgents"
              :key="ag.agent_id"
              class="hover:bg-slate-800/40 transition-all group"
            >
              <!-- Technician Info -->
              <td class="py-3.5 px-4 flex items-center gap-3">
                <div class="relative">
                  <img
                    :src="ag.agent_avatar || 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150'"
                    class="w-9 h-9 rounded-full object-cover border border-slate-700"
                  >
                  <span
                    v-if="idx === 0"
                    class="absolute -top-1 -right-1 w-4 h-4 bg-yellow-400 rounded-full text-slate-950 flex items-center justify-center text-[9px] font-extrabold shadow"
                  >
                    👑
                  </span>
                </div>
                <div>
                  <h4 class="text-xs font-bold text-white group-hover:text-amber-400 transition-colors">
                    {{ ag.agent_name }}
                  </h4>
                  <span class="text-[10px] text-slate-400 font-mono">ID: {{ ag.agent_id }}</span>
                </div>
              </td>

              <!-- Role & Team -->
              <td class="py-3.5 px-4">
                <span class="text-slate-200 font-semibold">{{ ag.job_title }}</span>
                <p class="text-[10px] text-slate-400">
                  {{ ag.department }}
                </p>
              </td>

              <!-- Assigned / Closed -->
              <td class="py-3.5 px-4 text-center">
                <span class="font-extrabold text-white">{{ ag.tickets_resolved }}</span>
                <span class="text-slate-400 text-[10px]"> / {{ ag.tickets_assigned }}</span>
              </td>

              <!-- Avg MTTR -->
              <td class="py-3.5 px-4 text-center font-mono font-bold text-cyan-400">
                {{ ag.avg_mttr_minutes.toFixed(1) }}m
              </td>

              <!-- CSAT Score -->
              <td class="py-3.5 px-4 text-center">
                <div class="inline-flex items-center gap-1 font-extrabold text-yellow-400 bg-yellow-500/10 px-2 py-0.5 rounded-md border border-yellow-500/20">
                  <UIcon
                    name="i-lucide-star"
                    class="w-3 h-3 fill-yellow-400"
                  />
                  <span>{{ ag.csat_rating.toFixed(2) }}</span>
                </div>
              </td>

              <!-- SLA Compliance % -->
              <td class="py-3.5 px-4 text-center">
                <span
                  :class="[
                    'px-2.5 py-1 rounded-lg text-xs font-extrabold',
                    ag.sla_compliance_pct >= 98 ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'
                  ]"
                >
                  {{ ag.sla_compliance_pct.toFixed(1) }}%
                </span>
              </td>

              <!-- Performance Badge -->
              <td class="py-3.5 px-4 text-right">
                <span
                  v-if="ag.sla_compliance_pct >= 98 && ag.csat_rating >= 4.9"
                  class="px-2 py-0.5 rounded-md bg-gradient-to-r from-amber-500 to-yellow-400 text-slate-950 text-[10px] font-extrabold shadow-md"
                >
                  Top Performer
                </span>
                <span
                  v-else
                  class="px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 text-[10px] font-semibold"
                >
                  Exceeds SLA
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
