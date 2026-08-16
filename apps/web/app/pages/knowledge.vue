<script setup lang="ts">
definePageMeta({ layout: 'default' })

const docs = [
  { title: 'Standard Laptop Setup & Security Baseline (macOS/Windows)', category: 'IT Security', updated: '2 days ago', views: 412 },
  { title: 'Corporate VPN & 2FA Configuration Guide (GlobalProtect)', category: 'Network', updated: '1 week ago', views: 890 },
  { title: 'PostgreSQL Database Access & Staging Secrets Policy', category: 'DevOps', updated: '2 weeks ago', views: 245 },
  { title: 'Hardware Replacement & Warranty Claim Workflow', category: 'Asset Management', updated: '3 weeks ago', views: 180 }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-book-open"
            class="w-6 h-6 text-cyan-400"
          />
          Knowledge Base & Technical Manuals
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          Vector-Indexed Documentation, IT SOPs & Self-Service Resolution Guides
        </p>
      </div>

      <button class="flex items-center gap-2 px-4 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold shadow-lg shadow-cyan-500/20 hover:scale-105 transition-all">
        <UIcon
          name="i-lucide-plus-circle"
          class="w-4 h-4"
        />
        <span>+ New Document</span>
      </button>
    </div>

    <!-- Search -->
    <div class="relative">
      <UIcon
        name="i-lucide-search"
        class="w-5 h-5 absolute left-4 top-1/2 -translate-y-1/2 text-slate-400"
      />
      <input
        type="text"
        placeholder="Semantic search documentation with AI Qdrant Vector Engine..."
        class="w-full pl-12 pr-4 py-3 text-sm rounded-2xl bg-slate-900/80 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-cyan-500"
      >
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div
        v-for="d in docs"
        :key="d.title"
        class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-cyan-500/40 transition-all group"
      >
        <div class="flex items-start justify-between gap-3 mb-2">
          <span class="text-[10px] font-mono font-bold px-2 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
            {{ d.category }}
          </span>
          <span class="text-[11px] text-slate-400">{{ d.views }} views</span>
        </div>

        <h3 class="text-sm font-bold text-white group-hover:text-cyan-300 transition-colors mb-2">
          {{ d.title }}
        </h3>

        <div class="pt-3 border-t border-slate-800/60 flex items-center justify-between text-xs text-slate-400">
          <span>Updated {{ d.updated }}</span>
          <span class="text-cyan-400 group-hover:underline">Read Guide &rarr;</span>
        </div>
      </div>
    </div>
  </div>
</template>
