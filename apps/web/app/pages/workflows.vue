<script setup lang="ts">
definePageMeta({ layout: 'default' })

const workflows = [
  { id: 'WF-301', title: 'Hardware Replacement Approval (Senior Dev)', requester: 'Alex Nguyen', step: 'Dept Head Review', status: 'Pending', time: '1 hour ago' },
  { id: 'WF-300', title: 'AWS Cloud Production Access Request', requester: 'David Tran', step: 'Security Lead Approval', status: 'Pending', time: '3 hours ago' },
  { id: 'WF-299', title: 'Annual Software License Purchase (Figma)', requester: 'Emily Davis', step: 'Finance Approval', status: 'Approved', time: 'Yesterday' }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div>
      <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
        <UIcon
          name="i-lucide-git-branch"
          class="w-6 h-6 text-purple-400"
        />
        Workflow & Approval Engine
      </h1>
      <p class="text-xs text-slate-400 mt-1">
        Multi-step Business Processes, State Machines & Incident Escalation Flows
      </p>
    </div>

    <div class="space-y-3">
      <div
        v-for="wf in workflows"
        :key="wf.id"
        class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex items-center justify-between gap-4"
      >
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <span class="font-mono text-xs font-bold text-purple-400 px-2 py-0.5 rounded bg-purple-500/10 border border-purple-500/20">{{ wf.id }}</span>
            <h3 class="text-sm font-bold text-white">
              {{ wf.title }}
            </h3>
          </div>
          <p class="text-xs text-slate-400">
            Requester: <span class="text-slate-300">{{ wf.requester }}</span> · Current Stage: <span class="text-purple-300 font-medium">{{ wf.step }}</span>
          </p>
        </div>

        <div class="flex items-center gap-3">
          <span
            class="text-xs font-semibold px-2.5 py-1 rounded-full border"
            :class="wf.status === 'Approved' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-amber-500/10 text-amber-300 border-amber-500/30'"
          >
            {{ wf.status }}
          </span>
          <button class="px-3 py-1.5 rounded-xl bg-purple-600 hover:bg-purple-500 text-white text-xs font-semibold transition-all">
            Review Step
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
