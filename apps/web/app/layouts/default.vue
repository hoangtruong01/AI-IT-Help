<script setup lang="ts">
const navigation = [
  {
    label: 'Dashboard',
    icon: 'i-lucide-layout-dashboard',
    to: '/'
  },
  {
    label: 'Employees',
    icon: 'i-lucide-users',
    to: '/employees'
  },
  {
    label: 'Assets',
    icon: 'i-lucide-monitor',
    to: '/assets'
  },
  {
    label: 'Help Desk',
    icon: 'i-lucide-ticket',
    to: '/helpdesk'
  },
  {
    label: 'Workflows',
    icon: 'i-lucide-git-branch',
    to: '/workflows'
  },
  {
    label: 'Knowledge Base',
    icon: 'i-lucide-book-open',
    to: '/knowledge'
  },
  {
    label: 'Reports',
    icon: 'i-lucide-bar-chart-3',
    to: '/reports'
  }
]
</script>

<template>
  <UDashboardLayout>
    <UDashboardPanel>
      <template #header>
        <div class="flex items-center gap-2 px-4 py-3">
          <UIcon
            name="i-lucide-boxes"
            class="size-6 text-primary"
          />
          <span class="font-semibold text-lg">EOMP</span>
        </div>
      </template>

      <UNavigationMenu
        :items="navigation"
        orientation="vertical"
        class="px-2"
      />

      <template #footer>
        <div class="flex items-center gap-2 px-4 py-3 border-t border-default">
          <UAvatar
            text="A"
            size="sm"
          />
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium truncate">
              Admin User
            </p>
            <p class="text-xs text-muted truncate">
              admin@eomp.local
            </p>
          </div>
          <UColorModeButton size="xs" />
        </div>
      </template>
    </UDashboardPanel>

    <slot />
  </UDashboardLayout>
</template>
