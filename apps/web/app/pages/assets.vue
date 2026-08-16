<script setup lang="ts">
definePageMeta({ layout: 'default' })

const assets = [
  { tag: 'AST-8092', model: 'MacBook Pro 16" M3 Max', type: 'Laptop', user: 'Sarah Tran', status: 'Assigned', purchaseDate: '2024-03-12', health: '99%' },
  { tag: 'AST-8091', model: 'Dell UltraSharp 32" 4K Monitor', type: 'Display', user: 'Alex Nguyen', status: 'Assigned', purchaseDate: '2023-11-20', health: '100%' },
  { tag: 'AST-8090', model: 'ThinkPad X1 Carbon Gen 11', type: 'Laptop', user: 'David Tran', status: 'Assigned', purchaseDate: '2023-08-15', health: '94%' },
  { tag: 'AST-8089', model: 'Cisco Catalyst 9300 Switch', type: 'Networking', user: 'Server Room A', status: 'In Service', purchaseDate: '2022-05-10', health: '98%' },
  { tag: 'AST-8088', model: 'Apple Studio Display 27"', type: 'Display', user: 'Inventory Storage', status: 'In Stock', purchaseDate: '2024-01-18', health: '100%' }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-laptop"
            class="w-6 h-6 text-emerald-400"
          />
          IT Asset Inventory
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Hardware Lifecycle, Asset Tagging, Maintenance Logs & Software Licensing
        </p>
      </div>

      <button class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-emerald-600 to-teal-600 text-white text-xs font-semibold shadow-lg shadow-emerald-500/20 hover:scale-105 transition-all">
        <UIcon
          name="i-lucide-plus-circle"
          class="w-4 h-4"
        />
        <span>+ Register New Asset</span>
      </button>
    </div>

    <!-- Assets Table -->
    <div class="overflow-hidden rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead class="bg-slate-950/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800/80">
            <tr>
              <th class="p-4">
                Asset Tag
              </th>
              <th class="p-4">
                Model & Description
              </th>
              <th class="p-4">
                Category
              </th>
              <th class="p-4">
                Assigned To
              </th>
              <th class="p-4">
                Status
              </th>
              <th class="p-4">
                Health
              </th>
              <th class="p-4 text-right">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <tr
              v-for="a in assets"
              :key="a.tag"
              class="hover:bg-slate-800/40 transition-colors group"
            >
              <td class="p-4 font-mono font-bold text-emerald-400">
                {{ a.tag }}
              </td>
              <td class="p-4 font-semibold text-white group-hover:text-emerald-300 transition-colors">
                {{ a.model }}
              </td>
              <td class="p-4 font-mono text-slate-400">
                {{ a.type }}
              </td>
              <td class="p-4 text-slate-300">
                {{ a.user }}
              </td>
              <td class="p-4">
                <span
                  class="px-2 py-0.5 rounded-full text-[10px] font-medium border"
                  :class="a.status === 'In Stock' ? 'bg-cyan-500/10 text-cyan-300 border-cyan-500/20' : 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20'"
                >
                  {{ a.status }}
                </span>
              </td>
              <td class="p-4 font-mono text-emerald-400">
                {{ a.health }}
              </td>
              <td class="p-4 text-right">
                <button class="px-2.5 py-1 rounded-lg bg-emerald-600/20 hover:bg-emerald-600/40 text-emerald-300 border border-emerald-500/30 text-[11px] font-medium transition-colors">
                  Details &rarr;
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
