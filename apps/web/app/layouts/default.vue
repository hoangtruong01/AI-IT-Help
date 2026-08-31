<script setup lang="ts">
import type { Notification, PaginatedResponse, NotificationStats } from '~/types'

const route = useRoute()
const authStore = useAuthStore()
const api = useApi()

const isSidebarOpen = ref(true)

// Notification Dropdown State
const isNotificationsOpen = ref(false)
const notifications = ref<Notification[]>([])
const notifStats = ref<NotificationStats>({ total: 0, unread: 0, incidents: 0, approvals: 0 })
const notifLoading = ref(false)

type NavigationItem = {
  label: string
  icon: string
  to: string
  roles?: string[]
}

const navigationGroups = computed(() => {
  const role = authStore.role
  const groups: Array<{ title: string, items: NavigationItem[] }> = [
    {
      title: 'Core Operations',
      items: [
        {
          label: 'Dashboard',
          icon: 'i-lucide-layout-dashboard',
          to: '/'
        },
        {
          label: 'IT Help Desk',
          icon: 'i-lucide-ticket',
          to: '/helpdesk'
        },
        {
          label: 'Problem Management',
          icon: 'i-lucide-alert-octagon',
          to: '/problems',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER', 'ROLE_AGENT']
        },
        {
          label: 'Change Advisory (CAB)',
          icon: 'i-lucide-git-pull-request',
          to: '/changes',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER']
        },
        {
          label: 'AI Ops Assistant',
          icon: 'i-lucide-bot',
          to: '/ai'
        }
      ]
    },
    {
      title: 'Resource Management',
      items: [
        {
          label: 'Employees & Orgs',
          icon: 'i-lucide-users',
          to: '/employees'
        },
        {
          label: 'IT Asset Inventory',
          icon: 'i-lucide-laptop',
          to: '/assets',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER', 'ROLE_AGENT']
        },
        {
          label: 'Workflow & Approvals',
          icon: 'i-lucide-git-branch',
          to: '/workflows'
        }
      ]
    },
    {
      title: 'Intelligence & Insights',
      items: [
        {
          label: 'Observability & SRE',
          icon: 'i-lucide-activity',
          to: '/monitoring',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER']
        },
        {
          label: 'Knowledge Base',
          icon: 'i-lucide-book-open',
          to: '/knowledge'
        },
        {
          label: 'Reports & Analytics',
          icon: 'i-lucide-bar-chart-3',
          to: '/reports',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER']
        },
        {
          label: 'Audit & Compliance',
          icon: 'i-lucide-shield-check',
          to: '/audit',
          roles: ['ROLE_ADMIN', 'ROLE_MANAGER']
        }
      ]
    }
  ]

  return groups
    .map(group => ({
      ...group,
      items: group.items.filter(item => !item.roles || item.roles.includes(role))
    }))
    .filter(group => group.items.length > 0)
})

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

async function fetchNotifications() {
  notifLoading.value = true
  try {
    const [listRes, statsRes] = await Promise.all([
      api.get<PaginatedResponse<Notification>>('/api/v1/notifications', { page: 1, page_size: 10 }),
      api.get<NotificationStats>('/api/v1/notifications/stats')
    ])
    notifications.value = listRes?.data || []
    if (statsRes) notifStats.value = statsRes
  } catch (err) {
    console.error('Failed to load notifications:', err)
  } finally {
    notifLoading.value = false
  }
}

async function markNotificationAsRead(id: string) {
  try {
    await api.patch(`/api/v1/notifications/${id}/read`)
    const target = notifications.value.find(n => n.id === id)
    if (target) target.is_read = true
    if (notifStats.value.unread > 0) notifStats.value.unread--
  } catch (err) {
    console.error('Failed to mark read:', err)
  }
}

async function markAllNotificationsAsRead() {
  try {
    await api.post('/api/v1/notifications/read-all')
    notifications.value.forEach((n) => {
      n.is_read = true
    })
    notifStats.value.unread = 0
  } catch (err) {
    console.error('Failed to mark all read:', err)
  }
}

function getCategoryColor(cat: string) {
  switch (cat) {
    case 'INCIDENT': return 'text-rose-400 bg-rose-500/15 border-rose-500/30'
    case 'APPROVAL': return 'text-amber-400 bg-amber-500/15 border-amber-500/30'
    case 'ASSET': return 'text-emerald-400 bg-emerald-500/15 border-emerald-500/30'
    case 'SLA': return 'text-purple-400 bg-purple-500/15 border-purple-500/30'
    default: return 'text-cyan-400 bg-cyan-500/15 border-cyan-500/30'
  }
}

onMounted(() => {
  if (authStore.isAuthenticated) {
    fetchNotifications()
  }
})
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 flex font-sans antialiased selection:bg-indigo-500/30 selection:text-indigo-200">
    <!-- Ambient Background Glows -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden z-0">
      <div class="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/10 rounded-full blur-3xl" />
      <div class="absolute top-1/3 -right-40 w-[500px] h-[500px] bg-cyan-600/10 rounded-full blur-3xl" />
      <div class="absolute -bottom-40 left-1/3 w-96 h-96 bg-emerald-600/10 rounded-full blur-3xl" />
    </div>

    <!-- Sidebar -->
    <aside
      class="fixed inset-y-0 left-0 z-40 flex flex-col border-r border-slate-800/80 bg-slate-900/80 backdrop-blur-xl transition-all duration-300 ease-in-out"
      :class="isSidebarOpen ? 'w-64' : 'w-20'"
    >
      <!-- Brand Header -->
      <div class="h-16 flex items-center justify-between px-4 border-b border-slate-800/80">
        <NuxtLink
          to="/"
          class="flex items-center gap-3 group min-w-0"
        >
          <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-indigo-600 via-indigo-500 to-cyan-400 p-[1px] shadow-lg shadow-indigo-500/20 group-hover:shadow-indigo-500/40 transition-all duration-300 shrink-0">
            <div class="w-full h-full bg-slate-950 rounded-[11px] flex items-center justify-center">
              <UIcon
                name="i-lucide-sparkles"
                class="w-5 h-5 text-indigo-400 group-hover:scale-110 transition-transform duration-300"
              />
            </div>
          </div>

          <div
            v-if="isSidebarOpen"
            class="flex flex-col min-w-0 transition-opacity duration-200"
          >
            <div class="flex items-center gap-1.5">
              <span class="font-bold text-base tracking-tight bg-gradient-to-r from-white via-slate-100 to-slate-300 bg-clip-text text-transparent">
                EOMP
              </span>
              <span class="text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-indigo-500/15 text-indigo-300 border border-indigo-500/30">
                Enterprise
              </span>
            </div>
            <span class="text-xs text-slate-400 truncate">IT & Ops Platform</span>
          </div>
        </NuxtLink>

        <button
          class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition-colors"
          title="Toggle Sidebar"
          @click="toggleSidebar"
        >
          <UIcon
            :name="isSidebarOpen ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'"
            class="w-4 h-4"
          />
        </button>
      </div>

      <!-- Navigation Links -->
      <div class="flex-1 overflow-y-auto px-3 py-4 space-y-6 scrollbar-thin scrollbar-thumb-slate-800">
        <div
          v-for="group in navigationGroups"
          :key="group.title"
          class="space-y-1"
        >
          <div
            v-if="isSidebarOpen"
            class="px-3 text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2"
          >
            {{ group.title }}
          </div>

          <NuxtLink
            v-for="item in group.items"
            :key="item.to"
            :to="item.to"
            class="flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all duration-200 group relative"
            :class="route.path === item.to
              ? 'bg-gradient-to-r from-indigo-600/20 to-indigo-500/10 text-white border border-indigo-500/30 shadow-sm shadow-indigo-500/10'
              : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 border border-transparent'"
          >
            <!-- Active Indicator Pill -->
            <div
              v-if="route.path === item.to"
              class="absolute left-0 w-1 h-5 bg-indigo-500 rounded-r-full"
            />

            <UIcon
              :name="item.icon"
              class="w-5 h-5 shrink-0 transition-transform duration-200 group-hover:scale-110"
              :class="route.path === item.to ? 'text-indigo-400' : 'text-slate-400 group-hover:text-slate-200'"
            />

            <span
              v-if="isSidebarOpen"
              class="truncate flex-1"
            >
              {{ item.label }}
            </span>

          </NuxtLink>
        </div>
      </div>

      <!-- User Profile Footer -->
      <div class="p-3 border-t border-slate-800/80 bg-slate-950/40">
        <div
          v-if="authStore.isAuthenticated && authStore.user"
          class="flex items-center gap-3 p-2 rounded-xl bg-slate-900/80 border border-slate-800/80 hover:border-slate-700/80 transition-colors"
        >
          <div class="relative shrink-0">
            <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-indigo-500 to-cyan-500 flex items-center justify-center font-bold text-white text-sm shadow-md">
              {{ authStore.user.full_name ? authStore.user.full_name.split(' ').map(n => n[0]).join('').slice(0, 2) : 'US' }}
            </div>
            <div class="absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 bg-emerald-500 rounded-full border-2 border-slate-900" />
          </div>

          <div
            v-if="isSidebarOpen"
            class="flex-1 min-w-0"
          >
            <p class="text-xs font-bold text-white truncate">
              {{ authStore.user.full_name }}
            </p>
            <p class="text-[10px] text-indigo-400 font-semibold truncate">
              {{ authStore.user.role }}
            </p>
          </div>

          <div
            v-if="isSidebarOpen"
            class="flex items-center gap-1"
          >
            <button
              class="p-1 text-slate-400 hover:text-rose-400 transition-colors"
              title="Sign Out"
              @click="authStore.logout"
            >
              <UIcon
                name="i-lucide-log-out"
                class="w-4 h-4"
              />
            </button>
          </div>
        </div>

        <div
          v-else
          class="p-2"
        >
          <NuxtLink
            to="/login"
            class="w-full py-2 px-3 rounded-xl bg-indigo-600/20 hover:bg-indigo-600/40 text-indigo-300 border border-indigo-500/30 text-xs font-semibold flex items-center justify-center gap-2 transition-all"
          >
            <UIcon
              name="i-lucide-log-in"
              class="w-4 h-4"
            />
            <span v-if="isSidebarOpen">Sign In</span>
          </NuxtLink>
        </div>
      </div>
    </aside>

    <!-- Main Content Area -->
    <div
      class="flex-1 flex flex-col min-w-0 transition-all duration-300 z-10"
      :class="isSidebarOpen ? 'pl-64' : 'pl-20'"
    >
      <!-- Top Navigation Bar -->
      <header class="h-16 sticky top-0 z-30 flex items-center justify-between px-6 border-b border-slate-800/80 bg-slate-950/70 backdrop-blur-xl">
        <!-- Search & Command Palette -->
        <div class="flex items-center gap-4 flex-1 max-w-lg">
          <div class="relative w-full">
            <UIcon
              name="i-lucide-search"
              class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none"
            />
            <input
              type="text"
              placeholder="Search tickets, assets, employees, docs... (Ctrl + K)"
              class="w-full pl-9 pr-12 py-1.5 text-sm rounded-xl bg-slate-900/90 border border-slate-800/90 text-slate-200 placeholder-slate-400 focus:outline-none focus:border-indigo-500/60 focus:ring-2 focus:ring-indigo-500/20 transition-all"
            >
            <kbd class="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-mono font-semibold px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700/60">
              ⌘K
            </kbd>
          </div>
        </div>

        <!-- Right Quick Actions -->
        <div class="flex items-center gap-3">
          <!-- Live Service Status Pill -->
          <div class="hidden md:flex items-center gap-2 px-3 py-1.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs font-medium">
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
            <span>11 Microservices Online</span>
          </div>

          <!-- Notification Bell & Dropdown -->
          <div class="relative">
            <button
              class="relative p-2 rounded-xl bg-slate-900/80 border border-slate-800/80 text-slate-300 hover:text-white hover:border-slate-700 transition-colors"
              title="Notifications"
              @click="isNotificationsOpen = !isNotificationsOpen; if (isNotificationsOpen) fetchNotifications()"
            >
              <UIcon
                name="i-lucide-bell"
                class="w-4 h-4"
              />
              <span
                v-if="notifStats.unread > 0"
                class="absolute -top-1 -right-1 px-1.5 py-0.2 rounded-full bg-rose-500 text-white font-bold text-[9px] ring-2 ring-slate-950 animate-bounce"
              >
                {{ notifStats.unread }}
              </span>
            </button>

            <!-- Notifications Dropdown Popover -->
            <div
              v-if="isNotificationsOpen"
              class="absolute right-0 mt-2 w-80 sm:w-96 rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-4 z-50 space-y-3"
            >
              <div class="flex items-center justify-between border-b border-slate-800 pb-2.5">
                <div class="flex items-center gap-2">
                  <h4 class="text-xs font-bold text-white uppercase tracking-wider">
                    Notifications
                  </h4>
                  <span
                    v-if="notifStats.unread > 0"
                    class="text-[10px] font-bold px-2 py-0.5 rounded-full bg-rose-500/20 text-rose-300"
                  >
                    {{ notifStats.unread }} new
                  </span>
                </div>
                <button
                  class="text-[11px] text-indigo-400 hover:text-indigo-300 font-semibold"
                  @click="markAllNotificationsAsRead"
                >
                  Mark all read
                </button>
              </div>

              <!-- Notifications List -->
              <div class="space-y-2 max-h-72 overflow-y-auto">
                <div
                  v-if="notifLoading && notifications.length === 0"
                  class="p-6 text-center text-slate-500 text-xs animate-pulse"
                >
                  Loading alerts...
                </div>

                <div
                  v-else-if="notifications.length === 0"
                  class="p-6 text-center text-slate-500 text-xs"
                >
                  No notifications yet.
                </div>

                <div
                  v-for="n in notifications"
                  :key="n.id"
                  class="p-3 rounded-2xl border text-xs space-y-1 transition-colors cursor-pointer"
                  :class="n.is_read ? 'bg-slate-950/40 border-slate-800/60 text-slate-400' : 'bg-slate-950 border-slate-800 text-slate-200 hover:border-indigo-500/40'"
                  @click="markNotificationAsRead(n.id)"
                >
                  <div class="flex items-center justify-between text-[11px]">
                    <span
                      class="px-2 py-0.5 rounded-full font-bold text-[9px] border"
                      :class="getCategoryColor(n.category)"
                    >
                      {{ n.category }}
                    </span>
                    <span class="text-slate-500 text-[10px]">{{ new Date(n.created_at).toLocaleTimeString() }}</span>
                  </div>
                  <p class="font-bold text-white text-[12px]">
                    {{ n.title }}
                  </p>
                  <p class="text-slate-300 text-[11px] line-clamp-2">
                    {{ n.message }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Quick Create Action Button -->
          <NuxtLink
            to="/helpdesk"
            class="flex items-center gap-2 px-3.5 py-1.5 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white font-medium text-xs shadow-md shadow-indigo-600/20 hover:shadow-indigo-600/40 transition-all duration-200"
          >
            <UIcon
              name="i-lucide-plus"
              class="w-4 h-4"
            />
            <span>Create Ticket</span>
          </NuxtLink>
        </div>
      </header>

      <!-- Page View Body -->
      <main class="flex-1 p-6 md:p-8">
        <slot />
      </main>
    </div>
  </div>
</template>
