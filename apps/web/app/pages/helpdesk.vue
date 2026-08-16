<script setup lang="ts">
definePageMeta({ layout: 'default' })

const searchQuery = ref('')
const selectedPriority = ref('All')

const tickets = [
  { id: 'TK-1094', title: 'VPN Connection Failure on Windows 11', category: 'Network', requester: 'Alex Nguyen', priority: 'Urgent', status: 'In Progress', createdAt: '10 mins ago', sla: '50m left' },
  { id: 'TK-1093', title: 'Request Dual Monitor Setup & Docking Station', category: 'Hardware', requester: 'Emily Davis', priority: 'Normal', status: 'Assigned', createdAt: '25 mins ago', sla: '3h left' },
  { id: 'TK-1092', title: 'Cannot access PostgreSQL Staging Cluster', category: 'DevOps', requester: 'David Tran', priority: 'High', status: 'Investigating', createdAt: '45 mins ago', sla: '1h 15m left' },
  { id: 'TK-1091', title: 'Microsoft 365 License renewal & 2FA reset', category: 'Software', requester: 'Michael Chang', priority: 'Normal', status: 'Resolved', createdAt: '2 hours ago', sla: 'Completed' },
  { id: 'TK-1090', title: 'New Employee Laptop Provisioning (MacBook Pro)', category: 'Hardware', requester: 'HR Operations', priority: 'High', status: 'Pending', createdAt: '3 hours ago', sla: '4h left' },
  { id: 'TK-1089', title: 'Printer in 4th floor finance room offline', category: 'Office IT', requester: 'Jessica Vo', priority: 'Low', status: 'Open', createdAt: '5 hours ago', sla: '7h left' }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-ticket"
            class="w-6 h-6 text-indigo-400"
          />
          IT Help Desk Management
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Service Request Management, SLA Monitoring, Incident Triage & Automated Dispatch
        </p>
      </div>

      <button class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 transition-all hover:scale-105">
        <UIcon
          name="i-lucide-plus-circle"
          class="w-4 h-4"
        />
        <span>+ Create New Ticket</span>
      </button>
    </div>

    <!-- Filter & Search Bar -->
    <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row items-center justify-between gap-3">
      <div class="relative w-full sm:w-96">
        <UIcon
          name="i-lucide-search"
          class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
        />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter by ticket ID, user, keyword..."
          class="w-full pl-9 pr-4 py-1.5 text-xs rounded-xl bg-slate-950/80 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-indigo-500"
        >
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <span class="text-xs text-slate-400 font-medium">Priority:</span>
        <select
          v-model="selectedPriority"
          class="px-3 py-1.5 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
        >
          <option value="All">
            All Priorities
          </option>
          <option value="Urgent">
            Urgent
          </option>
          <option value="High">
            High
          </option>
          <option value="Normal">
            Normal
          </option>
          <option value="Low">
            Low
          </option>
        </select>
      </div>
    </div>

    <!-- Ticket Table / Cards -->
    <div class="overflow-hidden rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800/80">
            <tr>
              <th class="p-4">
                Ticket ID & Title
              </th>
              <th class="p-4">
                Category
              </th>
              <th class="p-4">
                Requester
              </th>
              <th class="p-4">
                Priority
              </th>
              <th class="p-4">
                Status
              </th>
              <th class="p-4">
                SLA Countdown
              </th>
              <th class="p-4 text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <tr
              v-for="t in tickets"
              :key="t.id"
              class="hover:bg-slate-800/40 transition-colors group"
            >
              <td class="p-4">
                <div class="flex items-center gap-2.5">
                  <span class="font-mono text-[11px] font-bold text-indigo-400 px-2 py-0.5 rounded bg-indigo-500/10 border border-indigo-500/20">
                    {{ t.id }}
                  </span>
                  <span class="font-semibold text-white group-hover:text-indigo-300 transition-colors">
                    {{ t.title }}
                  </span>
                </div>
              </td>
              <td class="p-4 font-mono text-slate-400">
                {{ t.category }}
              </td>
              <td class="p-4 text-white">
                {{ t.requester }}
              </td>
              <td class="p-4">
                <span
                  class="px-2 py-0.5 rounded-full font-semibold text-[10px] border"
                  :class="t.priority === 'Urgent' ? 'bg-rose-500/15 text-rose-300 border-rose-500/30' : t.priority === 'High' ? 'bg-amber-500/15 text-amber-300 border-amber-500/30' : 'bg-slate-800 text-slate-300 border-slate-700'"
                >
                  {{ t.priority }}
                </span>
              </td>
              <td class="p-4">
                <span
                  class="px-2 py-0.5 rounded-full text-[10px]"
                  :class="t.status === 'Resolved' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-indigo-500/10 text-indigo-300 border border-indigo-500/20'"
                >
                  {{ t.status }}
                </span>
              </td>
              <td class="p-4 font-mono text-amber-400">
                {{ t.sla }}
              </td>
              <td class="p-4 text-right">
                <button class="px-2.5 py-1 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/40 text-indigo-300 border border-indigo-500/30 text-[11px] font-medium transition-colors">
                  Open &rarr;
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
