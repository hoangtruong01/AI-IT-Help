<script setup lang="ts">
import type { Asset, AssetStats, AssetAssignment, ConfigurationItem, CMDBTopologyGraph, PaginatedResponse, CreateAssetPayload } from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()

// Active Tab
const activeTab = ref<'inventory' | 'cmdb'>('inventory')

// Asset Inventory State
const assets = ref<Asset[]>([])
const stats = ref<AssetStats>({
  total_assets: 0,
  in_use: 0,
  in_stock: 0,
  in_maintenance: 0,
  total_value: 0
})
const loading = ref(false)
const error = ref<string | null>(null)

// Filters
const searchQuery = ref('')
const selectedCategory = ref('All')
const selectedStatus = ref('All')

// Register Modal State
const isRegisterModalOpen = ref(false)
const registering = ref(false)
const registerError = ref<string | null>(null)
const newAsset = reactive<CreateAssetPayload>({
  asset_tag: '',
  name: '',
  category: 'LAPTOP',
  model: '',
  serial_number: '',
  purchase_date: new Date().toISOString().split('T')[0],
  purchase_cost: 1500,
  warranty_expiry: '',
  location: 'Headquarters Warehouse',
  notes: ''
})

// Assign Modal State
const isAssignModalOpen = ref(false)
const assigningAsset = ref<Asset | null>(null)
const assignUserID = ref('')
const assignUserName = ref('')
const assignCondition = ref('EXCELLENT')
const assignNotes = ref('')
const assignLoading = ref(false)

// History & Incidents Modal State
interface AssetIncidentItem {
  ticket_id: string
  ticket_number: string
  title: string
  category: string
  priority: string
  status: string
  requester_id: string
  requester_name: string
  assignee_name?: string
  created_at: string
  resolved_at?: string
}

const isHistoryModalOpen = ref(false)
const historyAsset = ref<Asset | null>(null)
const historyRecords = ref<AssetAssignment[]>([])
const incidentRecords = ref<AssetIncidentItem[]>([])
const activeHistoryTab = ref<'assignments' | 'incidents'>('assignments')
const loadingHistory = ref(false)

// CMDB State
const topology = ref<CMDBTopologyGraph | null>(null)
const cmdbLoading = ref(false)
const isCIModalOpen = ref(false)
const creatingCI = ref(false)
const newCI = reactive({
  ci_code: '',
  name: '',
  ci_type: 'APPLICATION',
  environment: 'PRODUCTION',
  ip_address: '',
  description: ''
})

async function fetchStats() {
  try {
    const res = await api.get<AssetStats>('/api/v1/assets/stats')
    if (res) stats.value = res
  } catch (err) {
    console.error('Failed to load asset stats:', err)
  }
}

async function fetchAssets() {
  loading.value = true
  error.value = null
  try {
    const params: Record<string, unknown> = {
      page: 1,
      page_size: 50
    }
    if (searchQuery.value) params.search = searchQuery.value
    if (selectedCategory.value && selectedCategory.value !== 'All') params.category = selectedCategory.value
    if (selectedStatus.value && selectedStatus.value !== 'All') params.status = selectedStatus.value

    const res = await api.get<PaginatedResponse<Asset>>('/api/v1/assets', params)
    assets.value = res?.data || []
  } catch (err: unknown) {
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    error.value = errObj?.data?.error?.message || errObj?.message || 'Failed to fetch assets from Asset service.'
  } finally {
    loading.value = false
  }
}

async function fetchTopology() {
  cmdbLoading.value = true
  try {
    const res = await api.get<CMDBTopologyGraph>('/api/v1/cmdb/topology')
    topology.value = res || { nodes: [], edges: [] }
  } catch (err) {
    console.error('Failed to load CMDB topology:', err)
  } finally {
    cmdbLoading.value = false
  }
}

async function handleRegisterAsset() {
  if (!newAsset.asset_tag || !newAsset.name || !newAsset.category) {
    registerError.value = 'Asset tag, name and category are required.'
    return
  }

  registering.value = true
  registerError.value = null
  try {
    await api.post<Asset>('/api/v1/assets', { ...newAsset })
    isRegisterModalOpen.value = false
    Object.assign(newAsset, {
      asset_tag: '',
      name: '',
      category: 'LAPTOP',
      model: '',
      serial_number: '',
      purchase_cost: 1500,
      notes: ''
    })
    await Promise.all([fetchAssets(), fetchStats()])
  } catch (err: unknown) {
    const errObj = err as { data?: { error?: { message?: string } }, message?: string }
    registerError.value = errObj?.data?.error?.message || errObj?.message || 'Failed to register asset.'
  } finally {
    registering.value = false
  }
}

function openAssignModal(asset: Asset) {
  assigningAsset.value = asset
  assignUserID.value = ''
  assignUserName.value = ''
  assignCondition.value = 'EXCELLENT'
  assignNotes.value = ''
  isAssignModalOpen.value = true
}

async function handleAssignSubmit() {
  if (!assigningAsset.value || !assignUserID.value || !assignUserName.value) return
  assignLoading.value = true
  try {
    await api.post(`/api/v1/assets/${assigningAsset.value.id}/assign`, {
      user_id: assignUserID.value,
      user_name: assignUserName.value,
      condition_on_assign: assignCondition.value,
      notes: assignNotes.value,
      version: assigningAsset.value.version ?? 1
    })
    isAssignModalOpen.value = false
    await Promise.all([fetchAssets(), fetchStats()])
  } catch (err) {
    console.error('Failed to assign asset:', err)
  } finally {
    assignLoading.value = false
  }
}

async function handleReturnAsset(asset: Asset) {
  if (!confirm(`Confirm return asset ${asset.asset_tag} (${asset.name}) to inventory stock?`)) return
  try {
    await api.post(`/api/v1/assets/${asset.id}/return`, {
      condition: 'GOOD',
      notes: 'Returned to warehouse stock',
      version: asset.version ?? 1
    })
    await Promise.all([fetchAssets(), fetchStats()])
  } catch (err) {
    console.error('Failed to return asset:', err)
  }
}

async function openHistoryModal(asset: Asset) {
  historyAsset.value = asset
  isHistoryModalOpen.value = true
  historyRecords.value = []
  incidentRecords.value = []
  activeHistoryTab.value = 'assignments'
  loadingHistory.value = true
  try {
    const [assignments, incidents] = await Promise.all([
      api.get<AssetAssignment[]>(`/api/v1/assets/${asset.id}/assignments`).catch(() => []),
      api.get<AssetIncidentItem[]>(`/api/v1/assets/${asset.id}/incidents`).catch(() => [])
    ])
    historyRecords.value = assignments || []
    incidentRecords.value = incidents || []
  } catch (err) {
    console.error('Failed to load asset history & incidents:', err)
  } finally {
    loadingHistory.value = false
  }
}

async function handleCreateCI() {
  if (!newCI.ci_code || !newCI.name || !newCI.ci_type) return
  creatingCI.value = true
  try {
    await api.post<ConfigurationItem>('/api/v1/cmdb/ci', { ...newCI })
    isCIModalOpen.value = false
    Object.assign(newCI, {
      ci_code: '',
      name: '',
      ci_type: 'APPLICATION',
      environment: 'PRODUCTION',
      ip_address: '',
      description: ''
    })
    await fetchTopology()
  } catch (err) {
    console.error('Failed to create CI:', err)
  } finally {
    creatingCI.value = false
  }
}

// Watch filters
let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch([searchQuery, selectedCategory, selectedStatus], () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchAssets()
  }, 250)
})

onMounted(() => {
  fetchStats()
  fetchAssets()
  fetchTopology()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-laptop"
            class="w-6 h-6 text-emerald-400"
          />
          IT Asset & CMDB Management
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Hardware/License Lifecycle State Machine & Configuration Item (CI) Dependency Topology (:8083)
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          v-if="activeTab === 'inventory'"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white text-xs font-semibold shadow-lg shadow-emerald-500/20 hover:scale-105 transition-all"
          @click="isRegisterModalOpen = true"
        >
          <UIcon
            name="i-lucide-plus-circle"
            class="w-4 h-4"
          />
          <span>+ Register New Asset</span>
        </button>

        <button
          v-else
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 hover:scale-105 transition-all"
          @click="isCIModalOpen = true"
        >
          <UIcon
            name="i-lucide-plus-circle"
            class="w-4 h-4"
          />
          <span>+ Add CI Node</span>
        </button>
      </div>
    </div>

    <!-- Live KPI Metrics Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Total Inventory
        </p>
        <p class="text-2xl font-extrabold text-white mt-1">
          {{ stats.total_assets }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          In Active Use
        </p>
        <p class="text-2xl font-extrabold text-emerald-400 mt-1">
          {{ stats.in_use }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Ready In Stock
        </p>
        <p class="text-2xl font-extrabold text-cyan-400 mt-1">
          {{ stats.in_stock }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Maintenance
        </p>
        <p class="text-2xl font-extrabold text-amber-400 mt-1">
          {{ stats.in_maintenance }}
        </p>
      </div>
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl col-span-2 sm:col-span-1">
        <p class="text-slate-500 text-[11px] font-semibold uppercase tracking-wider">
          Fleet Valuation
        </p>
        <p class="text-2xl font-extrabold text-purple-400 mt-1 font-mono">
          ${{ stats.total_value.toLocaleString() }}
        </p>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex items-center gap-2 border-b border-slate-800 pb-2">
      <button
        class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
        :class="activeTab === 'inventory' ? 'bg-emerald-600/20 text-emerald-300 border border-emerald-500/30' : 'text-slate-400 hover:text-white'"
        @click="activeTab = 'inventory'"
      >
        <UIcon
          name="i-lucide-laptop"
          class="w-4 h-4"
        />
        <span>Hardware & License Inventory</span>
      </button>

      <button
        class="px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-2"
        :class="activeTab === 'cmdb' ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30' : 'text-slate-400 hover:text-white'"
        @click="activeTab = 'cmdb'"
      >
        <UIcon
          name="i-lucide-network"
          class="w-4 h-4"
        />
        <span>CMDB & Dependency Topology Map</span>
      </button>
    </div>

    <!-- TAB 1: ASSET INVENTORY -->
    <div
      v-if="activeTab === 'inventory'"
      class="space-y-4"
    >
      <!-- Filter Bar -->
      <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row items-center justify-between gap-3">
        <div class="relative w-full sm:w-80">
          <UIcon
            name="i-lucide-search"
            class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
          />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by tag, name, model, serial..."
            class="w-full pl-9 pr-4 py-2 text-xs rounded-xl bg-slate-950/80 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-emerald-500"
          >
        </div>

        <div class="flex items-center gap-2.5 w-full sm:w-auto">
          <!-- Category Filter -->
          <select
            v-model="selectedCategory"
            class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
          >
            <option value="All">
              All Categories
            </option>
            <option value="LAPTOP">
              Laptop
            </option>
            <option value="DESKTOP">
              Desktop
            </option>
            <option value="SERVER">
              Server
            </option>
            <option value="MONITOR">
              Monitor
            </option>
            <option value="NETWORK">
              Network
            </option>
            <option value="LICENSE">
              License
            </option>
          </select>

          <!-- Status Filter -->
          <select
            v-model="selectedStatus"
            class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
          >
            <option value="All">
              All Statuses
            </option>
            <option value="IN_USE">
              In Use
            </option>
            <option value="IN_STOCK">
              In Stock
            </option>
            <option value="MAINTENANCE">
              Maintenance
            </option>
            <option value="RETIRED">
              Retired
            </option>
          </select>

          <button
            class="p-2 rounded-xl bg-slate-800/60 hover:bg-slate-800 text-slate-300 transition-colors"
            title="Refresh"
            @click="fetchAssets"
          >
            <UIcon
              name="i-lucide-refresh-cw"
              class="w-4 h-4"
              :class="{ 'animate-spin': loading }"
            />
          </button>
        </div>
      </div>

      <!-- Error alert -->
      <div
        v-if="error"
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
          @click="fetchAssets"
        >
          Retry
        </button>
      </div>

      <!-- Table -->
      <div class="overflow-hidden rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-950/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800/80">
              <tr>
                <th class="p-4">
                  Asset Tag
                </th>
                <th class="p-4">
                  Asset Name & Specs
                </th>
                <th class="p-4">
                  Category
                </th>
                <th class="p-4">
                  Assigned User / Location
                </th>
                <th class="p-4">
                  Cost / Value
                </th>
                <th class="p-4">
                  Status
                </th>
                <th class="p-4 text-right">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-slate-300">
              <template v-if="loading && assets.length === 0">
                <tr
                  v-for="i in 5"
                  :key="i"
                >
                  <td
                    colspan="7"
                    class="p-4 text-center text-slate-500 animate-pulse"
                  >
                    Loading asset inventory from service...
                  </td>
                </tr>
              </template>

              <tr v-else-if="assets.length === 0">
                <td
                  colspan="7"
                  class="p-8 text-center text-slate-500"
                >
                  No assets found matching current filter.
                </td>
              </tr>

              <tr
                v-for="a in assets"
                :key="a.id"
                class="hover:bg-slate-800/40 transition-colors group"
              >
                <td class="p-4">
                  <span class="font-mono font-bold text-emerald-400 px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
                    {{ a.asset_tag }}
                  </span>
                </td>
                <td class="p-4">
                  <p class="font-semibold text-white group-hover:text-emerald-300 transition-colors">
                    {{ a.name }}
                  </p>
                  <p
                    v-if="a.model"
                    class="text-[11px] text-slate-400 font-mono mt-0.5 truncate max-w-xs"
                  >
                    {{ a.model }}
                  </p>
                </td>
                <td class="p-4 font-mono text-slate-400">
                  {{ a.category }}
                </td>
                <td class="p-4">
                  <p class="text-white font-medium">
                    {{ a.assigned_to_user_name || 'In Warehouse' }}
                  </p>
                  <p class="text-[11px] text-slate-400 mt-0.5">
                    {{ a.location }}
                  </p>
                </td>
                <td class="p-4 font-mono">
                  <span class="text-slate-300 font-medium">${{ a.current_value.toLocaleString() }}</span>
                </td>
                <td class="p-4">
                  <span
                    class="px-2 py-0.5 rounded-full text-[10px] font-medium border"
                    :class="a.status === 'IN_USE' ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30' : a.status === 'IN_STOCK' ? 'bg-cyan-500/15 text-cyan-300 border-cyan-500/30' : 'bg-amber-500/15 text-amber-300 border-amber-500/30'"
                  >
                    {{ a.status }}
                  </span>
                </td>
                <td class="p-4 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <button
                      v-if="a.status === 'IN_STOCK'"
                      class="px-2.5 py-1 rounded-lg bg-emerald-600/20 hover:bg-emerald-600/40 text-emerald-300 border border-emerald-500/30 text-[11px] font-medium transition-colors"
                      @click="openAssignModal(a)"
                    >
                      Assign
                    </button>
                    <button
                      v-else-if="a.status === 'IN_USE'"
                      class="px-2.5 py-1 rounded-lg bg-amber-600/20 hover:bg-amber-600/40 text-amber-300 border border-amber-500/30 text-[11px] font-medium transition-colors"
                      @click="handleReturnAsset(a)"
                    >
                      Return
                    </button>
                    <button
                      class="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-[11px]"
                      title="Assignment History"
                      @click="openHistoryModal(a)"
                    >
                      <UIcon
                        name="i-lucide-history"
                        class="w-3.5 h-3.5"
                      />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: CMDB TOPOLOGY -->
    <div
      v-else-if="activeTab === 'cmdb'"
      class="space-y-6"
    >
      <div
        v-if="cmdbLoading"
        class="p-12 text-center text-slate-500 animate-pulse"
      >
        Loading CMDB Dependency Topology...
      </div>

      <div
        v-else-if="topology"
        class="space-y-6"
      >
        <!-- CI Grid -->
        <div>
          <h3 class="text-sm font-bold text-white uppercase tracking-wider mb-3 flex items-center gap-2">
            <UIcon
              name="i-lucide-server"
              class="w-4 h-4 text-indigo-400"
            />
            Registered Configuration Items (CIs)
          </h3>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div
              v-for="node in topology.nodes"
              :key="node.id"
              class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-2 hover:border-indigo-500/40 transition-colors"
            >
              <div class="flex items-center justify-between">
                <span class="font-mono text-xs font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                  {{ node.ci_code }}
                </span>
                <span class="text-[10px] px-2 py-0.5 rounded-full font-semibold bg-emerald-500/15 text-emerald-300 border border-emerald-500/30">
                  {{ node.status }}
                </span>
              </div>
              <h4 class="text-xs font-bold text-white">
                {{ node.name }}
              </h4>
              <div class="text-[11px] text-slate-400 space-y-0.5">
                <p>Type: <span class="text-slate-200 font-mono">{{ node.ci_type }}</span></p>
                <p>Environment: <span class="text-indigo-300 font-medium">{{ node.environment }}</span></p>
                <p v-if="node.ip_address">
                  IP: <span class="text-slate-300 font-mono">{{ node.ip_address }}</span>
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Relationship Dependencies Map -->
        <div class="p-6 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
          <h3 class="text-sm font-bold text-white uppercase tracking-wider flex items-center gap-2">
            <UIcon
              name="i-lucide-git-branch"
              class="w-4 h-4 text-purple-400"
            />
            Infrastructure CI Dependency Graph (Topology Map)
          </h3>
          <div class="space-y-3">
            <div
              v-for="edge in topology.edges"
              :key="edge.id"
              class="p-3.5 rounded-2xl bg-slate-950/80 border border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
            >
              <!-- Parent Node -->
              <div class="flex items-center gap-2">
                <span class="p-2 rounded-xl bg-indigo-500/10 text-indigo-400 font-bold font-mono">
                  {{ edge.parent_ci_name }}
                </span>
                <span class="text-slate-500">({{ edge.parent_ci_type }})</span>
              </div>

              <!-- Relationship Edge -->
              <div class="flex items-center gap-2 text-purple-300 font-mono font-semibold px-3 py-1 rounded-full bg-purple-500/10 border border-purple-500/20 text-[11px]">
                <span>&mdash; [ {{ edge.relationship_type }} ] &rarr;</span>
              </div>

              <!-- Child Node -->
              <div class="flex items-center gap-2">
                <span class="p-2 rounded-xl bg-teal-500/10 text-teal-400 font-bold font-mono">
                  {{ edge.child_ci_name }}
                </span>
                <span class="text-slate-500">({{ edge.child_ci_type }})</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Register Asset Modal -->
    <div
      v-if="isRegisterModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-5">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-plus-circle"
              class="w-5 h-5 text-emerald-400"
            />
            Register Hardware / Software Asset
          </h3>
          <button
            class="text-slate-400 hover:text-white"
            @click="isRegisterModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div
          v-if="registerError"
          class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs"
        >
          {{ registerError }}
        </div>

        <form
          class="space-y-3 text-xs"
          @submit.prevent="handleRegisterAsset"
        >
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Asset Tag *</label>
              <input
                v-model="newAsset.asset_tag"
                type="text"
                required
                placeholder="AST-2001"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500 font-mono"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Category *</label>
              <select
                v-model="newAsset.category"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
              >
                <option value="LAPTOP">
                  Laptop
                </option>
                <option value="DESKTOP">
                  Desktop
                </option>
                <option value="SERVER">
                  Server
                </option>
                <option value="MONITOR">
                  Monitor
                </option>
                <option value="NETWORK">
                  Network Device
                </option>
                <option value="LICENSE">
                  Software License
                </option>
              </select>
            </div>
          </div>

          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Asset Name *</label>
            <input
              v-model="newAsset.name"
              type="text"
              required
              placeholder="e.g. MacBook Pro 16 M3 Max"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
            >
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Model / Specs</label>
              <input
                v-model="newAsset.model"
                type="text"
                placeholder="36GB RAM / 1TB SSD"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Serial Number</label>
              <input
                v-model="newAsset.serial_number"
                type="text"
                placeholder="C02G81YQMD6R"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500 font-mono"
              >
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Purchase Cost ($)</label>
              <input
                v-model.number="newAsset.purchase_cost"
                type="number"
                step="0.01"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500 font-mono"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Location</label>
              <input
                v-model="newAsset.location"
                type="text"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
              >
            </div>
          </div>

          <div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold"
              @click="isRegisterModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="registering"
              class="px-5 py-2 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-semibold flex items-center gap-2 shadow-lg shadow-emerald-500/25 disabled:opacity-50"
            >
              <UIcon
                v-if="registering"
                name="i-lucide-loader-2"
                class="w-4 h-4 animate-spin"
              />
              <span>{{ registering ? 'Saving...' : 'Register Asset' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Assign Modal -->
    <div
      v-if="isAssignModalOpen && assigningAsset"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-md rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <h3 class="text-base font-bold text-white">
          Assign Asset: {{ assigningAsset.asset_tag }}
        </h3>
        <p class="text-slate-400">
          Allocate {{ assigningAsset.name }} to an employee.
        </p>

        <form
          class="space-y-3"
          @submit.prevent="handleAssignSubmit"
        >
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Employee ID *</label>
            <input
              v-model="assignUserID"
              type="text"
              required
              placeholder="e0000000-0000-..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
            >
          </div>
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Employee Full Name *</label>
            <input
              v-model="assignUserName"
              type="text"
              required
              placeholder="Emily Davis"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
            >
          </div>
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Condition on Handover</label>
            <select
              v-model="assignCondition"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-emerald-500"
            >
              <option value="EXCELLENT">
                Brand New / Excellent
              </option>
              <option value="GOOD">
                Good Condition
              </option>
              <option value="FAIR">
                Fair / Minor Wear
              </option>
            </select>
          </div>
          <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300"
              @click="isAssignModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="assignLoading"
              class="px-5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-semibold"
            >
              {{ assignLoading ? 'Assigning...' : 'Confirm Assignment' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- History Modal -->
    <div
      v-if="isHistoryModalOpen && historyAsset"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-base font-bold text-white">
            Assignment Audit Trail: {{ historyAsset.asset_tag }}
          </h3>
          <button
            class="text-slate-400 hover:text-white"
            @click="isHistoryModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div
          v-if="historyRecords.length === 0"
          class="p-6 text-center text-slate-500"
        >
          No historical handovers found for this asset.
        </div>

        <div
          v-else
          class="space-y-3 max-h-60 overflow-y-auto"
        >
          <div
            v-for="rec in historyRecords"
            :key="rec.id"
            class="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-1"
          >
            <div class="flex items-center justify-between font-semibold text-white">
              <span>{{ rec.user_name }}</span>
              <span class="text-slate-500">{{ new Date(rec.assigned_at).toLocaleDateString() }}</span>
            </div>
            <p class="text-[11px] text-slate-400">
              Condition on Assign: <span class="text-slate-200">{{ rec.condition_on_assign }}</span> | Status: <span :class="rec.returned_at ? 'text-cyan-400' : 'text-emerald-400'">{{ rec.returned_at ? 'Returned' : 'Currently Holding' }}</span>
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Add CI Modal -->
    <div
      v-if="isCIModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-md rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 text-xs">
        <h3 class="text-base font-bold text-white">
          Register Configuration Item (CMDB)
        </h3>

        <form
          class="space-y-3"
          @submit.prevent="handleCreateCI"
        >
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">CI Code *</label>
            <input
              v-model="newCI.ci_code"
              type="text"
              required
              placeholder="CI-SRV-REDIS"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono"
            >
          </div>
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">Name *</label>
            <input
              v-model="newCI.name"
              type="text"
              required
              placeholder="Redis Cache Cluster"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white"
            >
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Type</label>
              <select
                v-model="newCI.ci_type"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white"
              >
                <option value="APPLICATION">
                  Application
                </option>
                <option value="API_SERVICE">
                  API Service
                </option>
                <option value="SERVER">
                  Server
                </option>
                <option value="DATABASE">
                  Database
                </option>
                <option value="CLOUD_RESOURCE">
                  Cloud Resource
                </option>
              </select>
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Environment</label>
              <select
                v-model="newCI.environment"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white"
              >
                <option value="PRODUCTION">
                  Production
                </option>
                <option value="STAGING">
                  Staging
                </option>
                <option value="DEVELOPMENT">
                  Development
                </option>
              </select>
            </div>
          </div>
          <div class="space-y-1">
            <label class="font-semibold text-slate-300">IP Address</label>
            <input
              v-model="newCI.ip_address"
              type="text"
              placeholder="10.0.2.30"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono"
            >
          </div>
          <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300"
              @click="isCIModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="creatingCI"
              class="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white font-semibold"
            >
              {{ creatingCI ? 'Adding...' : 'Register CI' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Asset History & Incident Tickets Modal -->
    <div
      v-if="isHistoryModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-2xl rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-4 animate-scale-up">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <div>
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <UIcon
                name="i-lucide-history"
                class="w-5 h-5 text-emerald-400"
              />
              <span>Asset Lifecycle & Incident History</span>
            </h3>
            <p
              v-if="historyAsset"
              class="text-xs text-slate-400 mt-0.5"
            >
              Hardware: <span class="font-mono text-emerald-400 font-bold">{{ historyAsset.asset_tag }}</span> — {{ historyAsset.name }} ({{ historyAsset.status }})
            </p>
          </div>
          <button
            class="text-slate-400 hover:text-white"
            @click="isHistoryModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <!-- Tabs: Assignments vs Incidents -->
        <div class="flex border-b border-slate-800 gap-4 text-xs font-semibold">
          <button
            class="pb-2 flex items-center gap-1.5 transition-colors border-b-2"
            :class="activeHistoryTab === 'assignments' ? 'border-emerald-500 text-emerald-400' : 'border-transparent text-slate-400 hover:text-slate-200'"
            @click="activeHistoryTab = 'assignments'"
          >
            <UIcon
              name="i-lucide-users"
              class="w-4 h-4"
            />
            <span>Assignments ({{ historyRecords.length }})</span>
          </button>
          <button
            class="pb-2 flex items-center gap-1.5 transition-colors border-b-2"
            :class="activeHistoryTab === 'incidents' ? 'border-amber-500 text-amber-400' : 'border-transparent text-slate-400 hover:text-slate-200'"
            @click="activeHistoryTab = 'incidents'"
          >
            <UIcon
              name="i-lucide-alert-triangle"
              class="w-4 h-4"
            />
            <span>Incident Tickets ({{ incidentRecords.length }})</span>
          </button>
        </div>

        <div
          v-if="loadingHistory"
          class="py-12 flex flex-col items-center justify-center space-y-3"
        >
          <UIcon
            name="i-lucide-loader-2"
            class="w-7 h-7 text-emerald-400 animate-spin"
          />
          <span class="text-xs text-slate-400">Loading asset logs...</span>
        </div>

        <!-- Assignments Tab -->
        <div v-else-if="activeHistoryTab === 'assignments'">
          <div
            v-if="historyRecords.length === 0"
            class="py-8 text-center text-slate-500 text-xs"
          >
            No assignment history found for this asset.
          </div>
          <div
            v-else
            class="max-h-80 overflow-y-auto space-y-2.5 pr-1"
          >
            <div
              v-for="rec in historyRecords"
              :key="rec.id"
              class="p-3 rounded-2xl bg-slate-950/60 border border-slate-800 text-xs space-y-1.5"
            >
              <div class="flex items-center justify-between">
                <span class="font-semibold text-white">{{ rec.user_name }}</span>
                <span
                  class="text-[10px] px-2 py-0.5 rounded-full border"
                  :class="rec.returned_at ? 'bg-slate-800 text-slate-400 border-slate-700' : 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30'"
                >
                  {{ rec.returned_at ? 'RETURNED' : 'IN USE' }}
                </span>
              </div>
              <div class="flex items-center gap-4 text-[11px] text-slate-400">
                <span>Assigned: {{ new Date(rec.assigned_at).toLocaleDateString() }} ({{ rec.condition_on_assign }})</span>
                <span v-if="rec.returned_at">Returned: {{ new Date(rec.returned_at).toLocaleDateString() }}</span>
              </div>
              <p
                v-if="rec.notes"
                class="text-[11px] text-slate-400 italic"
              >
                "{{ rec.notes }}"
              </p>
            </div>
          </div>
        </div>

        <!-- Incidents Tab -->
        <div v-else-if="activeHistoryTab === 'incidents'">
          <div
            v-if="incidentRecords.length === 0"
            class="py-8 text-center text-slate-500 text-xs"
          >
            No incident tickets reported on this hardware.
          </div>
          <div
            v-else
            class="max-h-80 overflow-y-auto space-y-2.5 pr-1"
          >
            <div
              v-for="inc in incidentRecords"
              :key="inc.ticket_id"
              class="p-3 rounded-2xl bg-slate-950/60 border border-slate-800 text-xs space-y-1.5"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <span class="font-mono text-[11px] font-bold text-amber-400">{{ inc.ticket_number }}</span>
                  <span class="font-semibold text-white">{{ inc.title }}</span>
                </div>
                <span
                  class="text-[10px] px-2 py-0.5 rounded-full border"
                  :class="inc.status === 'RESOLVED' || inc.status === 'CLOSED' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-amber-500/10 text-amber-300 border-amber-500/30'"
                >
                  {{ inc.status }}
                </span>
              </div>
              <div class="flex items-center gap-4 text-[11px] text-slate-400">
                <span>Requester: {{ inc.requester_name }}</span>
                <span>Priority: <strong class="text-white">{{ inc.priority }}</strong></span>
                <span>Logged: {{ new Date(inc.created_at).toLocaleDateString() }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-3 flex justify-end border-t border-slate-800">
          <button
            class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold text-xs"
            @click="isHistoryModalOpen = false"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
