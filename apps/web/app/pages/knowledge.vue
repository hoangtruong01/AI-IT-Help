<script setup lang="ts">
import type {
  KnowledgeArticle,
  KnowledgeCategory,
  KnowledgeRunbook,
  KnowledgeSearchResult,
  KnowledgeStats,
  CreateArticlePayload,
  CreateRunbookPayload,
  PaginatedResponse
} from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()

// State
const activeTab = ref<'articles' | 'runbooks'>('articles')
const selectedCategory = ref<string>('All')
const searchQuery = ref('')
const isLoading = ref(false)
const isSearching = ref(false)

// Modals & Drawers
const isArticleDrawerOpen = ref(false)
const selectedArticle = ref<KnowledgeArticle | null>(null)
const isCreateArticleModalOpen = ref(false)
const isCreateRunbookModalOpen = ref(false)
const isSubmitting = ref(false)

// Data Collections
const stats = ref<KnowledgeStats>({
  total_articles: 6,
  total_categories: 5,
  total_runbooks: 4,
  total_views: 4440
})

const categories = ref<KnowledgeCategory[]>([
  { id: 'c1', name: 'IT Security & Access', code: 'sec', icon: 'i-lucide-shield-check', description: 'MFA, Okta, Zero Trust' },
  { id: 'c2', name: 'Network & Connectivity', code: 'net', icon: 'i-lucide-network', description: 'VPN, DNS, Firewall' },
  { id: 'c3', name: 'Hardware & Equipment', code: 'hw', icon: 'i-lucide-laptop', description: 'Laptops, Monitors, RMA' },
  { id: 'c4', name: 'DevOps & Infrastructure', code: 'devops', icon: 'i-lucide-server', description: 'PostgreSQL, K8s, MinIO' },
  { id: 'c5', name: 'Software & Apps', code: 'soft', icon: 'i-lucide-layers', description: 'OS, Tools, Licenses' }
])

const articles = ref<KnowledgeArticle[]>([])
const runbooks = ref<KnowledgeRunbook[]>([])
const searchResults = ref<KnowledgeSearchResult[]>([])

// Create Article Form
const articleForm = ref<CreateArticlePayload>({
  category_id: '',
  title: '',
  summary: '',
  content: '',
  tags: []
})
const articleTagsInput = ref('')

// Create Runbook Form
const runbookForm = ref<CreateRunbookPayload>({
  code: '',
  title: '',
  category: 'IT Security',
  description: '',
  prerequisites: '',
  steps_json: [
    { step: 1, action: '', command: '', expected: '' }
  ],
  rollback_steps: ''
})

// Fetch initial data
async function loadData() {
  isLoading.value = true
  try {
    // 1. Fetch Stats
    const statsRes = await api.get<KnowledgeStats>('/api/v1/knowledge/stats').catch(() => null)
    if (statsRes) stats.value = statsRes

    // 2. Fetch Categories
    const catRes = await api.get<KnowledgeCategory[]>('/api/v1/knowledge/categories').catch(() => null)
    if (catRes && catRes.length > 0) {
      categories.value = catRes
      const firstCat = catRes[0]
      if (firstCat && !articleForm.value.category_id) {
        articleForm.value.category_id = firstCat.id
      }
    }

    // 3. Fetch Articles
    const artRes = await api.get<PaginatedResponse<KnowledgeArticle>>('/api/v1/knowledge/articles', {
      params: {
        category: selectedCategory.value === 'All' ? undefined : selectedCategory.value,
        page_size: 50
      }
    }).catch(() => null)
    if (artRes && artRes.data) {
      articles.value = artRes.data
    } else {
      // Fallback dummy articles
      articles.value = [
        {
          id: 'a1',
          category_id: 'c1',
          category_name: 'IT Security & Access',
          category_code: 'sec',
          title: 'How to Reset User MFA and Okta Verify Tokens',
          slug: 'how-to-reset-user-mfa-tokens',
          summary: 'Official standard operating procedure for IT Support Agents to verify employee identity and securely reset multi-factor authentication tokens in Okta / Keycloak.',
          content: '## Step 1: Verify Identity\nVerify identity via secondary corporate channel.\n\n## Step 2: Okta Admin\nOpen id.eomp.local/admin and clear factor enrollment.\n\n## Step 3: Issue QR Code\nGenerate 15-minute activation link.',
          tags: ['MFA', 'Okta', 'Security', 'SOP'],
          author_id: 'u1',
          author_name: 'Sarah Jenkins (IT Security Lead)',
          view_count: 1240,
          helpful_count: 85,
          is_published: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: 'a2',
          category_id: 'c2',
          category_name: 'Network & Connectivity',
          category_code: 'net',
          title: 'Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide',
          slug: 'vpn-troubleshooting-guide',
          summary: 'Comprehensive resolution guide for remote engineers experiencing VPN disconnects, handshake timeouts, and MTU routing packet loss.',
          content: '## Diagnostic Steps\n1. ping 10.8.0.1\n2. Verify MTU=1380\n3. Restart WireGuard daemon.',
          tags: ['VPN', 'WireGuard', 'Network', 'DNS'],
          author_id: 'u2',
          author_name: 'Alex Rivera (Network Architect)',
          view_count: 980,
          helpful_count: 62,
          is_published: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: 'a3',
          category_id: 'c3',
          category_name: 'Hardware & Equipment',
          category_code: 'hw',
          title: 'Standard Laptop Setup & Security Baseline (macOS & Windows)',
          slug: 'standard-laptop-setup-baseline',
          summary: 'Full checklist for provisioning MacBook Pro and ThinkPad laptops for new hires, including FileVault, BitLocker, EDR agent, and developer tooling.',
          content: '## Provisioning Checklist\n1. FileVault / BitLocker disk encryption.\n2. CrowdStrike EDR installation.\n3. Corporate Root CA enrollment.',
          tags: ['Hardware', 'Laptop', 'Setup', 'Security'],
          author_id: 'u3',
          author_name: 'David Chen (IT Asset Lead)',
          view_count: 850,
          helpful_count: 45,
          is_published: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ]
    }

    // 4. Fetch Runbooks
    const rbRes = await api.get<PaginatedResponse<KnowledgeRunbook>>('/api/v1/knowledge/runbooks').catch(() => null)
    if (rbRes && rbRes.data) {
      runbooks.value = rbRes.data
    } else {
      runbooks.value = [
        {
          id: 'r1',
          code: 'RB-SEC-02',
          title: 'User MFA Token Reset and Identity Verification SOP',
          category: 'IT Security',
          description: 'Standardized operational procedure for identity authentication and Okta/Keycloak multi-factor token reissue.',
          prerequisites: '1. Active IT Support Agent credentials.\n2. Manager written approval or video verification log.',
          steps_json: [
            { step: 1, action: 'Verify Employee Identity', command: 'Check /employees directory', expected: 'Active status confirmed' },
            { step: 2, action: 'Open Okta / Auth Admin console', command: 'Navigate to /admin/users/{email}/factors', expected: 'Active factors displayed' },
            { step: 3, action: 'Revoke existing MFA token registration', command: 'POST /api/v1/auth/mfa/revoke', expected: 'Token status changed to REVOKED' },
            { step: 4, action: 'Generate temporary QR code', command: 'POST /api/v1/auth/mfa/re-enroll', expected: '15-min activation token issued' }
          ],
          rollback_steps: 'If user fails verification, lock account for 30 minutes and alert SOC.',
          author_name: 'Sarah Jenkins (IT Security Lead)',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        },
        {
          id: 'r2',
          code: 'RB-NET-01',
          title: 'Emergency VPN Tunnel Failover SOP',
          category: 'Network',
          description: 'Failover procedure when primary WireGuard VPN gateway server exhibits packet loss or hardware crash.',
          prerequisites: '1. Access to Secondary VPN Gateway (10.8.1.1).\n2. Cloudflare DNS admin permissions.',
          steps_json: [
            { step: 1, action: 'Check primary VPN tunnel status', command: 'wg show', expected: 'Identify disconnected peers' },
            { step: 2, action: 'Switch DNS endpoint record', command: 'cf-cli dns update vpn.eomp.local --ip 198.51.100.2', expected: 'DNS propagated within 60s' },
            { step: 3, action: 'Restart WireGuard client daemon', command: 'systemctl restart wg-quick@wg0', expected: 'Handshake re-established' }
          ],
          rollback_steps: 'Revert DNS A record back to primary gateway IP once upstream network recovers.',
          author_name: 'Alex Rivera (Network Architect)',
          is_active: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }
      ]
    }
  } finally {
    isLoading.value = false
  }
}

// Perform search
let searchTimer: ReturnType<typeof setTimeout> | null = null
function onSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  if (!searchQuery.value.trim()) {
    searchResults.value = []
    isSearching.value = false
    return
  }

  isSearching.value = true
  searchTimer = setTimeout(async () => {
    try {
      const res = await api.get<{
        query: string
        total: number
        results: KnowledgeSearchResult[]
      }>(
        '/api/v1/knowledge/search',
        { params: { q: searchQuery.value.trim() } }
      )
      searchResults.value = res.results || []
    } catch {
      // Local filter fallback
      const q = searchQuery.value.toLowerCase()
      searchResults.value = articles.value
        .filter(a => a.title.toLowerCase().includes(q) || a.summary.toLowerCase().includes(q))
        .map(a => ({
          id: a.id,
          type: 'article',
          title: a.title,
          snippet: a.summary,
          category: a.category_name || 'General',
          score: 0.92,
          tags: a.tags,
          slug_or_code: a.slug,
          view_count: a.view_count,
          updated_time: a.updated_at
        }))
    } finally {
      isSearching.value = false
    }
  }, 300)
}

// Select Article to read
function openArticle(art: KnowledgeArticle) {
  selectedArticle.value = art
  isArticleDrawerOpen.value = true
}

function openSearchResult(item: KnowledgeSearchResult) {
  if (item.type === 'article') {
    const found = articles.value.find(a => a.id === item.id || a.slug === item.slug_or_code)
    if (found) {
      openArticle(found)
    } else {
      selectedArticle.value = {
        id: item.id,
        category_id: '',
        category_name: item.category,
        title: item.title,
        slug: item.slug_or_code,
        summary: item.snippet,
        content: `# ${item.title}\n\n${item.snippet}\n\n*(Full content indexed in Qdrant Vector Store)*`,
        tags: item.tags || [],
        author_id: 'system',
        author_name: 'IT Knowledge Author',
        view_count: item.view_count || 100,
        helpful_count: 10,
        is_published: true,
        created_at: new Date().toISOString(),
        updated_at: item.updated_time
      }
      isArticleDrawerOpen.value = true
    }
  } else {
    activeTab.value = 'runbooks'
  }
}

// Add / Remove step in create runbook form
function addRunbookStep() {
  runbookForm.value.steps_json.push({
    step: runbookForm.value.steps_json.length + 1,
    action: '',
    command: '',
    expected: ''
  })
}

function removeRunbookStep(index: number) {
  if (runbookForm.value.steps_json.length > 1) {
    runbookForm.value.steps_json.splice(index, 1)
    runbookForm.value.steps_json.forEach((s, i) => {
      s.step = i + 1
    })
  }
}

// Submit Article
async function submitArticle() {
  if (!articleForm.value.title.trim() || !articleForm.value.content.trim()) {
    toast.add({ title: 'Validation Error', description: 'Title and content are required', color: 'error' })
    return
  }

  if (articleTagsInput.value) {
    articleForm.value.tags = articleTagsInput.value.split(',').map(t => t.trim()).filter(Boolean)
  }

  isSubmitting.value = true
  try {
    await api.post('/api/v1/knowledge/articles', articleForm.value)
    toast.add({ title: 'Article Published', description: 'Article added to knowledge base and queued for Qdrant vector indexing.', color: 'success' })
    isCreateArticleModalOpen.value = false
    const firstCatVal = categories.value[0]
    const defaultCatId = firstCatVal ? firstCatVal.id : ''
    articleForm.value = { category_id: defaultCatId, title: '', summary: '', content: '', tags: [] }
    articleTagsInput.value = ''
    loadData()
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Error', description: errObj?.message || 'Failed to create article', color: 'error' })
  } finally {
    isSubmitting.value = false
  }
}

// Submit Runbook
async function submitRunbook() {
  if (!runbookForm.value.code.trim() || !runbookForm.value.title.trim()) {
    toast.add({ title: 'Validation Error', description: 'Runbook code and title are required', color: 'error' })
    return
  }

  isSubmitting.value = true
  try {
    await api.post('/api/v1/knowledge/runbooks', runbookForm.value)
    toast.add({ title: 'Runbook Created', description: `Runbook ${runbookForm.value.code} published successfully.`, color: 'success' })
    isCreateRunbookModalOpen.value = false
    runbookForm.value = {
      code: '',
      title: '',
      category: 'IT Security',
      description: '',
      prerequisites: '',
      steps_json: [{ step: 1, action: '', command: '', expected: '' }],
      rollback_steps: ''
    }
    loadData()
  } catch (err: unknown) {
    const errObj = err as { message?: string }
    toast.add({ title: 'Error', description: errObj?.message || 'Failed to create runbook', color: 'error' })
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-book-open"
            class="w-7 h-7 text-cyan-400"
          />
          Knowledge Base & SOP Runbooks
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Qdrant Vector-Indexed IT Manuals, Standard Operating Procedures & Self-Service Guides
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <button
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 text-xs font-semibold shadow-md transition-all"
          @click="isCreateRunbookModalOpen = true"
        >
          <UIcon
            name="i-lucide-file-code"
            class="w-4 h-4 text-amber-400"
          />
          <span>+ New SOP Runbook</span>
        </button>

        <button
          class="flex items-center gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-cyan-600 to-cyan-500 hover:from-cyan-500 hover:to-cyan-400 text-white text-xs font-semibold shadow-lg shadow-cyan-500/20 hover:scale-105 transition-all"
          @click="isCreateArticleModalOpen = true"
        >
          <UIcon
            name="i-lucide-plus-circle"
            class="w-4 h-4"
          />
          <span>+ New Article</span>
        </button>
      </div>
    </div>

    <!-- 4 KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-cyan-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Total Articles</span>
          <UIcon
            name="i-lucide-file-text"
            class="w-5 h-5 text-cyan-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.total_articles }}
        </p>
        <span class="text-[10px] text-cyan-400 mt-1 block">Vector Synced to Qdrant</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-indigo-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">SOP Runbooks</span>
          <UIcon
            name="i-lucide-terminal"
            class="w-5 h-5 text-indigo-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.total_runbooks }}
        </p>
        <span class="text-[10px] text-indigo-400 mt-1 block">Step-by-step Procedures</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-emerald-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Categories</span>
          <UIcon
            name="i-lucide-folder-tree"
            class="w-5 h-5 text-emerald-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.total_categories }}
        </p>
        <span class="text-[10px] text-emerald-400 mt-1 block">IT Domain Trees</span>
      </div>

      <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-amber-500/30 transition-all">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-slate-400">Total Consultations</span>
          <UIcon
            name="i-lucide-eye"
            class="w-5 h-5 text-amber-400"
          />
        </div>
        <p class="text-2xl font-black text-white mt-2">
          {{ stats.total_views.toLocaleString() }}
        </p>
        <span class="text-[10px] text-amber-400 mt-1 block">Self-Service Views</span>
      </div>
    </div>

    <!-- Semantic Search Bar -->
    <div class="relative">
      <div class="relative">
        <UIcon
          name="i-lucide-search"
          class="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-cyan-400"
        />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Semantic search IT SOPs, MFA resets, VPN configuration, database failover with Qdrant vector engine..."
          class="w-full pl-12 pr-12 py-3.5 text-sm rounded-2xl bg-slate-900/90 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-cyan-500 shadow-xl backdrop-blur-xl transition-all"
          @input="onSearchInput"
        >
        <button
          v-if="searchQuery"
          class="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white"
          @click="searchQuery = ''; searchResults = []"
        >
          <UIcon
            name="i-lucide-x"
            class="w-4 h-4"
          />
        </button>
      </div>

      <!-- Live Search Instant Results Popover -->
      <div
        v-if="searchQuery.trim() && searchResults.length > 0"
        class="absolute left-0 right-0 top-full mt-2 p-3 rounded-2xl bg-slate-900/95 border border-cyan-500/30 shadow-2xl backdrop-blur-2xl z-30 space-y-2"
      >
        <div class="flex items-center justify-between px-3 py-1 text-[11px] font-mono text-cyan-400 border-b border-slate-800">
          <span>Qdrant Vector Ranked Results (Top {{ searchResults.length }})</span>
          <span class="text-slate-400">Click to view article / SOP</span>
        </div>

        <div
          v-for="res in searchResults"
          :key="res.id"
          class="p-3 rounded-xl hover:bg-slate-800/80 cursor-pointer transition-colors flex items-start justify-between gap-3"
          @click="openSearchResult(res)"
        >
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span
                class="text-[9px] font-mono px-1.5 py-0.5 rounded font-bold"
                :class="res.type === 'runbook' ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' : 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20'"
              >
                {{ res.type.toUpperCase() }}
              </span>
              <span class="text-xs font-bold text-white hover:text-cyan-300">{{ res.title }}</span>
            </div>
            <p class="text-[11px] text-slate-400 line-clamp-1">
              {{ res.snippet }}
            </p>
          </div>

          <div class="text-right shrink-0">
            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Score: {{ (res.score * 100).toFixed(0) }}%
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tabs & Category Filters -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-3">
      <!-- Main Tab Switcher -->
      <div class="flex items-center gap-1 p-1 bg-slate-900/80 border border-slate-800 rounded-xl w-fit">
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeTab === 'articles' ? 'bg-cyan-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeTab = 'articles'"
        >
          <UIcon
            name="i-lucide-book-open"
            class="w-4 h-4"
          />
          <span>Cẩm Nang Hướng Dẫn (Articles)</span>
        </button>
        <button
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold transition-all"
          :class="activeTab === 'runbooks' ? 'bg-amber-600 text-white shadow' : 'text-slate-400 hover:text-white'"
          @click="activeTab = 'runbooks'"
        >
          <UIcon
            name="i-lucide-terminal"
            class="w-4 h-4"
          />
          <span>SOP Runbooks (Kịch Bản Kỹ Thuật)</span>
        </button>
      </div>

      <!-- Category Filter Pills -->
      <div class="flex items-center gap-1.5 overflow-x-auto pb-1">
        <button
          class="px-3 py-1.5 rounded-xl text-xs font-semibold whitespace-nowrap transition-all"
          :class="selectedCategory === 'All' ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
          @click="selectedCategory = 'All'; loadData()"
        >
          Tất cả (All)
        </button>
        <button
          v-for="cat in categories"
          :key="cat.id"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold whitespace-nowrap transition-all"
          :class="selectedCategory === cat.code ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30' : 'bg-slate-900 text-slate-400 hover:text-white border border-slate-800'"
          @click="selectedCategory = cat.code; loadData()"
        >
          <UIcon
            :name="cat.icon"
            class="w-3.5 h-3.5"
          />
          <span>{{ cat.name }}</span>
        </button>
      </div>
    </div>

    <!-- Tab 1: Articles Grid -->
    <div
      v-if="activeTab === 'articles'"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5"
    >
      <div
        v-for="art in articles"
        :key="art.id"
        class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-cyan-500/40 hover:-translate-y-1 transition-all group flex flex-col justify-between cursor-pointer"
        @click="openArticle(art)"
      >
        <div>
          <div class="flex items-start justify-between gap-3 mb-3">
            <span class="text-[10px] font-mono font-bold px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
              {{ art.category_name || 'IT Guide' }}
            </span>
            <div class="flex items-center gap-1 text-[11px] text-slate-400">
              <UIcon
                name="i-lucide-eye"
                class="w-3.5 h-3.5"
              />
              <span>{{ art.view_count }}</span>
            </div>
          </div>

          <h3 class="text-sm font-bold text-white group-hover:text-cyan-300 transition-colors mb-2 line-clamp-2">
            {{ art.title }}
          </h3>

          <p class="text-xs text-slate-400 leading-relaxed line-clamp-3 mb-4">
            {{ art.summary }}
          </p>
        </div>

        <div class="space-y-3 pt-3 border-t border-slate-800/60">
          <div class="flex flex-wrap gap-1">
            <span
              v-for="t in art.tags"
              :key="t"
              class="text-[9px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700"
            >
              #{{ t }}
            </span>
          </div>

          <div class="flex items-center justify-between text-[11px] text-slate-400">
            <span class="truncate max-w-[140px]">{{ art.author_name }}</span>
            <span class="text-cyan-400 group-hover:underline flex items-center gap-1">
              Read Guide <UIcon
                name="i-lucide-arrow-right"
                class="w-3 h-3"
              />
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab 2: SOP Runbooks Grid -->
    <div
      v-if="activeTab === 'runbooks'"
      class="grid grid-cols-1 md:grid-cols-2 gap-5"
    >
      <div
        v-for="rb in runbooks"
        :key="rb.id"
        class="p-6 rounded-2xl bg-slate-900/70 border border-slate-800/80 backdrop-blur-xl hover:border-amber-500/40 transition-all space-y-4"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="space-y-1">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono font-bold px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                {{ rb.code }}
              </span>
              <span class="text-[10px] text-slate-400 font-mono">{{ rb.category }}</span>
            </div>
            <h3 class="text-sm font-bold text-white">
              {{ rb.title }}
            </h3>
          </div>

          <span class="px-2 py-0.5 text-[10px] font-semibold rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            SOP Active
          </span>
        </div>

        <p class="text-xs text-slate-300 leading-relaxed">
          {{ rb.description }}
        </p>

        <!-- Prerequisites Box -->
        <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800/80 text-[11px] space-y-1">
          <span class="text-amber-400 font-semibold flex items-center gap-1">
            <UIcon
              name="i-lucide-alert-circle"
              class="w-3.5 h-3.5"
            /> Điều Kiện Tiên Quyết (Prerequisites):
          </span>
          <p class="text-slate-400 whitespace-pre-line">
            {{ rb.prerequisites }}
          </p>
        </div>

        <!-- Steps Timeline -->
        <div class="space-y-2">
          <span class="text-xs font-semibold text-slate-200">Các Bước Thực Hiện (Execution Steps):</span>
          <div class="space-y-2">
            <div
              v-for="st in rb.steps_json"
              :key="st.step"
              class="p-2.5 rounded-xl bg-slate-950/60 border border-slate-800/60 flex items-start gap-3"
            >
              <span class="w-5 h-5 rounded-full bg-amber-500/20 text-amber-300 flex items-center justify-center text-[10px] font-mono font-bold shrink-0 mt-0.5">
                {{ st.step }}
              </span>
              <div class="space-y-1 text-xs flex-1">
                <p class="text-slate-200 font-medium">
                  {{ st.action }}
                </p>
                <p
                  v-if="st.command"
                  class="font-mono text-[11px] text-cyan-300 bg-slate-900 px-2 py-1 rounded border border-slate-800"
                >
                  {{ st.command }}
                </p>
                <p
                  v-if="st.expected"
                  class="text-[10px] text-slate-400"
                >
                  Expected: {{ st.expected }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Rollback Box -->
        <div
          v-if="rb.rollback_steps"
          class="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-[11px]"
        >
          <span class="text-rose-400 font-semibold">Kịch Bản Phục Hồi (Rollback):</span>
          <p class="text-slate-300 mt-0.5">
            {{ rb.rollback_steps }}
          </p>
        </div>
      </div>
    </div>

    <!-- Article Reader Modal Overlay -->
    <div
      v-if="isArticleDrawerOpen && selectedArticle"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-3xl max-h-[85vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-5 text-white shadow-2xl">
        <div class="flex items-start justify-between gap-4 border-b border-slate-800 pb-4">
          <div class="space-y-1">
            <span class="text-[10px] font-mono font-bold px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
              {{ selectedArticle.category_name || 'IT Documentation' }}
            </span>
            <h2 class="text-xl font-extrabold text-white mt-1">
              {{ selectedArticle.title }}
            </h2>
            <div class="flex items-center gap-3 text-xs text-slate-400 pt-1">
              <span>By {{ selectedArticle.author_name }}</span>
              <span>•</span>
              <span>{{ selectedArticle.view_count }} views</span>
            </div>
          </div>
          <button
            class="p-2 text-slate-400 hover:text-white rounded-lg bg-slate-800"
            @click="isArticleDrawerOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <!-- Summary Box -->
        <div class="p-4 rounded-xl bg-cyan-950/30 border border-cyan-500/30 text-xs text-cyan-200 leading-relaxed">
          <span class="font-bold block mb-1">Tóm Tắt Tài Liệu (Executive Summary):</span>
          {{ selectedArticle.summary }}
        </div>

        <!-- Full Content -->
        <div class="prose prose-invert max-w-none text-xs md:text-sm text-slate-300 leading-relaxed whitespace-pre-line space-y-4">
          {{ selectedArticle.content }}
        </div>

        <!-- Tags -->
        <div class="flex items-center gap-2 pt-4 border-t border-slate-800">
          <span class="text-xs text-slate-400">Tags:</span>
          <span
            v-for="t in selectedArticle.tags"
            :key="t"
            class="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700"
          >
            #{{ t }}
          </span>
        </div>
      </div>
    </div>

    <!-- Create Article Modal Overlay -->
    <div
      v-if="isCreateArticleModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-2xl max-h-[85vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-4 text-white shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h2 class="text-lg font-bold flex items-center gap-2 text-white">
            <UIcon
              name="i-lucide-file-plus"
              class="w-5 h-5 text-cyan-400"
            />
            Tạo Tài Liệu Knowledge Base Mới
          </h2>
          <button
            class="text-slate-400 hover:text-white"
            @click="isCreateArticleModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="space-y-3 text-xs">
          <div>
            <label class="block text-slate-300 font-semibold mb-1">Danh Mục (Category) *</label>
            <select
              v-model="articleForm.category_id"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500"
            >
              <option
                v-for="c in categories"
                :key="c.id"
                :value="c.id"
              >
                {{ c.name }}
              </option>
            </select>
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Tiêu Đề Bài Viết (Title) *</label>
            <input
              v-model="articleForm.title"
              type="text"
              placeholder="VD: Hướng dẫn cấu hình VPN WireGuard cho nhân viên mới"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500"
            >
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Tóm Tắt Ngắn (Summary)</label>
            <textarea
              v-model="articleForm.summary"
              rows="2"
              placeholder="Tóm tắt ngắn gọn mục tiêu của tài liệu..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Nội Dung Chi Tiết (Markdown Content) *</label>
            <textarea
              v-model="articleForm.content"
              rows="6"
              placeholder="Soạn thảo nội dung theo chuẩn Markdown..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono focus:outline-none focus:border-cyan-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Thẻ Từ Khóa (Tags - phân cách bởi dấu phẩy)</label>
            <input
              v-model="articleTagsInput"
              type="text"
              placeholder="VPN, WireGuard, Network, SOP"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500"
            >
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-3 border-t border-slate-800">
          <button
            class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:bg-slate-700 text-xs font-semibold"
            @click="isCreateArticleModalOpen = false"
          >
            Hủy
          </button>
          <button
            class="px-5 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold shadow-lg shadow-cyan-500/20 disabled:opacity-50"
            :disabled="isSubmitting"
            @click="submitArticle"
          >
            {{ isSubmitting ? 'Đang Lưu & Tạo Vector...' : 'Xuất Bản Bài Viết' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Create Runbook Modal Overlay -->
    <div
      v-if="isCreateRunbookModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm"
    >
      <div class="w-full max-w-2xl max-h-[85vh] overflow-y-auto p-6 bg-slate-900 border border-slate-800 rounded-3xl space-y-4 text-white shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h2 class="text-lg font-bold flex items-center gap-2 text-white">
            <UIcon
              name="i-lucide-file-code"
              class="w-5 h-5 text-amber-400"
            />
            Tạo Kịch Bản Vận Hành SOP Runbook
          </h2>
          <button
            class="text-slate-400 hover:text-white"
            @click="isCreateRunbookModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div class="space-y-3 text-xs">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Mã Runbook (Code) *</label>
              <input
                v-model="runbookForm.code"
                type="text"
                placeholder="VD: RB-NET-05"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white font-mono uppercase focus:outline-none focus:border-amber-500"
              >
            </div>
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Danh Mục (Category) *</label>
              <input
                v-model="runbookForm.category"
                type="text"
                placeholder="Network, IT Security, DevOps"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-amber-500"
              >
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Tên Quy Trình (Title) *</label>
            <input
              v-model="runbookForm.title"
              type="text"
              placeholder="VD: Quy trình xử lý lỗi kết nối VPN Gateway"
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-amber-500"
            >
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Mô Tả Mục Tiêu (Description)</label>
            <textarea
              v-model="runbookForm.description"
              rows="2"
              placeholder="Mô tả mục đích xử lý của quy trình..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-amber-500"
            />
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Điều Kiện Tiên Quyết (Prerequisites)</label>
            <textarea
              v-model="runbookForm.prerequisites"
              rows="2"
              placeholder="1. Quyền quản trị VPN... 2. Địa chỉ IP secondary..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-amber-500"
            />
          </div>

          <!-- Steps Builder -->
          <div class="space-y-2 pt-2">
            <div class="flex items-center justify-between">
              <label class="text-slate-300 font-semibold">Các Bước Thực Hiện (Steps)</label>
              <button
                type="button"
                class="text-[11px] text-amber-400 hover:underline flex items-center gap-1"
                @click="addRunbookStep"
              >
                <UIcon
                  name="i-lucide-plus"
                  class="w-3 h-3"
                /> Thêm Bước
              </button>
            </div>

            <div
              v-for="(st, idx) in runbookForm.steps_json"
              :key="idx"
              class="p-3 rounded-xl bg-slate-950 border border-slate-800 space-y-2"
            >
              <div class="flex items-center justify-between">
                <span class="text-[10px] font-mono font-bold text-amber-400">Bước {{ st.step }}</span>
                <button
                  v-if="runbookForm.steps_json.length > 1"
                  type="button"
                  class="text-rose-400 hover:text-rose-300 text-[10px]"
                  @click="removeRunbookStep(idx)"
                >
                  Xóa
                </button>
              </div>
              <input
                v-model="st.action"
                type="text"
                placeholder="Hành động thực hiện..."
                class="w-full px-2.5 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-white"
              >
              <input
                v-model="st.command"
                type="text"
                placeholder="Lệnh thực thi (Command)..."
                class="w-full px-2.5 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-white font-mono"
              >
              <input
                v-model="st.expected"
                type="text"
                placeholder="Kết quả kỳ vọng (Expected)..."
                class="w-full px-2.5 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-white"
              >
            </div>
          </div>

          <div>
            <label class="block text-slate-300 font-semibold mb-1">Kịch Bản Phục Hồi (Rollback Steps)</label>
            <textarea
              v-model="runbookForm.rollback_steps"
              rows="2"
              placeholder="Các bước hoàn tác nếu xử lý thất bại..."
              class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-amber-500"
            />
          </div>
        </div>

        <div class="flex items-center justify-end gap-3 pt-3 border-t border-slate-800">
          <button
            class="px-4 py-2 rounded-xl bg-slate-800 text-slate-300 hover:bg-slate-700 text-xs font-semibold"
            @click="isCreateRunbookModalOpen = false"
          >
            Hủy
          </button>
          <button
            class="px-5 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white text-xs font-semibold shadow-lg shadow-amber-500/20 disabled:opacity-50"
            :disabled="isSubmitting"
            @click="submitRunbook"
          >
            {{ isSubmitting ? 'Đang Tạo Runbook...' : 'Lưu & Kích Hoạt SOP' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
