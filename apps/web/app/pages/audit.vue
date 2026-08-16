<script setup lang="ts">
definePageMeta({ layout: 'default' })

const auditLogs = [
  { id: 'AUD-9912', event: 'AUTH_LOGIN_SUCCESS', actor: 'admin@eomp.local', ip: '192.168.1.10', service: 'auth-service', time: '5 mins ago', status: 'SUCCESS' },
  { id: 'AUD-9911', event: 'ASSET_REASSIGN', actor: 'sarah.tran@eomp.local', ip: '192.168.1.45', service: 'asset-service', time: '22 mins ago', status: 'SUCCESS' },
  { id: 'AUD-9910', event: 'TICKET_ESCALATE_TIER2', actor: 'system-ai', ip: '127.0.0.1', service: 'ai-service', time: '38 mins ago', status: 'SUCCESS' },
  { id: 'AUD-9909', event: 'WORKFLOW_APPROVE_STEP', actor: 'dept-head@eomp.local', ip: '192.168.1.12', service: 'workflow-service', time: '1 hour ago', status: 'SUCCESS' }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div>
      <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
        <UIcon
          name="i-lucide-shield-check"
          class="w-6 h-6 text-emerald-400"
        />
        Audit Logging & Compliance Trail
      </h1>
      <p class="text-xs text-slate-400 mt-1">
        Immutable Event Stream, Security Auditing, Role-Based Access Logs & Traceability
      </p>
    </div>

    <!-- Table -->
    <div class="overflow-hidden rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800/80">
            <tr>
              <th class="p-4">
                Audit ID
              </th>
              <th class="p-4">
                Event Type
              </th>
              <th class="p-4">
                Actor / User
              </th>
              <th class="p-4">
                Microservice
              </th>
              <th class="p-4">
                IP Address
              </th>
              <th class="p-4">
                Timestamp
              </th>
              <th class="p-4">
                Status
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <tr
              v-for="log in auditLogs"
              :key="log.id"
              class="hover:bg-slate-800/40 transition-colors"
            >
              <td class="p-4 font-mono font-bold text-indigo-400">
                {{ log.id }}
              </td>
              <td class="p-4 font-mono font-semibold text-white">
                {{ log.event }}
              </td>
              <td class="p-4 text-slate-300">
                {{ log.actor }}
              </td>
              <td class="p-4 font-mono text-cyan-400">
                {{ log.service }}
              </td>
              <td class="p-4 font-mono text-slate-400">
                {{ log.ip }}
              </td>
              <td class="p-4 text-slate-400">
                {{ log.time }}
              </td>
              <td class="p-4">
                <span class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
                  {{ log.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
