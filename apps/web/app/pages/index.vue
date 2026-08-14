<script setup lang="ts">
definePageMeta({
  layout: 'default'
})

const stats = [
  { label: 'Total Employees', value: '-', icon: 'i-lucide-users', color: 'text-blue-500' },
  { label: 'Active Assets', value: '-', icon: 'i-lucide-monitor', color: 'text-emerald-500' },
  { label: 'Open Tickets', value: '-', icon: 'i-lucide-ticket', color: 'text-amber-500' },
  { label: 'Pending Approvals', value: '-', icon: 'i-lucide-clock', color: 'text-purple-500' }
]
</script>

<template>
  <UDashboardPanel>
    <template #header>
      <div class="flex items-center justify-between w-full px-4">
        <h1 class="text-xl font-semibold">
          Dashboard
        </h1>
        <UBadge
          color="primary"
          variant="subtle"
          size="sm"
        >
          Setup Phase
        </UBadge>
      </div>
    </template>

    <div class="p-6 space-y-6">
      <!-- Stats Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard
          v-for="stat in stats"
          :key="stat.label"
        >
          <div class="flex items-center gap-3">
            <UIcon
              :name="stat.icon"
              :class="['size-8', stat.color]"
            />
            <div>
              <p class="text-sm text-muted">
                {{ stat.label }}
              </p>
              <p class="text-2xl font-bold">
                {{ stat.value }}
              </p>
            </div>
          </div>
        </UCard>
      </div>

      <!-- Welcome Section -->
      <UCard>
        <div class="text-center py-8">
          <UIcon
            name="i-lucide-boxes"
            class="size-12 text-primary mx-auto mb-4"
          />
          <h2 class="text-2xl font-bold mb-2">
            Welcome to EOMP
          </h2>
          <p class="text-muted max-w-lg mx-auto">
            Enterprise Operations Management Platform is being set up.
            Modules will be available once the development phase begins.
          </p>
        </div>
      </UCard>

      <!-- System Status -->
      <UCard>
        <template #header>
          <h3 class="font-semibold">
            System Status
          </h3>
        </template>
        <div class="space-y-2 text-sm">
          <div class="flex items-center justify-between">
            <span>Frontend</span>
            <UBadge
              color="success"
              variant="subtle"
              size="xs"
            >
              Running
            </UBadge>
          </div>
          <div class="flex items-center justify-between">
            <span>API Gateway</span>
            <UBadge
              color="neutral"
              variant="subtle"
              size="xs"
            >
              Not Started
            </UBadge>
          </div>
          <div class="flex items-center justify-between">
            <span>Database</span>
            <UBadge
              color="neutral"
              variant="subtle"
              size="xs"
            >
              Not Started
            </UBadge>
          </div>
        </div>
      </UCard>
    </div>
  </UDashboardPanel>
</template>
