<script setup lang="ts">
definePageMeta({ layout: 'default' })
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-bar-chart-3"
            class="w-6 h-6 text-amber-400"
          />
          Operations BI & Analytics
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Service Level Agreements, Incident Resolution Times, Asset Depreciations & Grafana Live Links
        </p>
      </div>

      <a
        href="http://localhost:3002"
        target="_blank"
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white text-xs font-semibold shadow-md transition-all"
      >
        <UIcon
          name="i-lucide-external-link"
          class="w-4 h-4"
        />
        <span>Open Grafana Portal (:3002)</span>
      </a>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
      <div class="p-6 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3">
        <span class="text-xs font-semibold text-slate-400">Mean Time To Resolution (MTTR)</span>
        <p class="text-3xl font-extrabold text-white">
          42 <span class="text-base font-normal text-slate-400">minutes</span>
        </p>
        <span class="text-xs text-emerald-400 font-medium">↓ 18% improvement vs last month</span>
      </div>

      <div class="p-6 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3">
        <span class="text-xs font-semibold text-slate-400">First Contact Resolution Rate</span>
        <p class="text-3xl font-extrabold text-white">
          88.6%
        </p>
        <span class="text-xs text-emerald-400 font-medium">↑ 5.2% automated by AI</span>
      </div>

      <div class="p-6 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-3">
        <span class="text-xs font-semibold text-slate-400">Hardware Fleet Utilization</span>
        <p class="text-3xl font-extrabold text-white">
          96.4%
        </p>
        <span class="text-xs text-slate-300 font-medium">3,850 Total Tracked Assets</span>
      </div>
    </div>
  </div>
</template>
