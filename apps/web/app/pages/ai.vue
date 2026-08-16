<script setup lang="ts">
definePageMeta({ layout: 'default' })

const messages = ref([
  { role: 'assistant', text: 'Hello Administrator! I am the EOMP AI Operations Copilot, integrated with your 11 microservices and Qdrant Vector Knowledge Base. How can I assist you with operations, incident triage, or log analysis today?' }
])

const inputQuery = ref('')

function sendMessage() {
  if (!inputQuery.value.trim()) return
  const q = inputQuery.value
  messages.value.push({ role: 'user', text: q })
  inputQuery.value = ''

  setTimeout(() => {
    messages.value.push({
      role: 'assistant',
      text: `[AI Operations Analysis for: "${q}"]\n\n🔍 Scanned Helpdesk & Microservices:\n- Analyzed current ticket cluster TK-1094 (VPN failure).\n- Root Cause Probability: DNS resolver timeout on gateway subnet.\n- Automated Resolution: Applied cached routing table update via Gateway Service.`
    })
  }, 600)
}
</script>

<template>
  <div class="space-y-6 max-w-5xl mx-auto h-[calc(100vh-140px)] flex flex-col">
    <!-- Header -->
    <div class="flex items-center justify-between pb-2">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-bot"
            class="w-6 h-6 text-indigo-400"
          />
          AI Operations Copilot
        </h1>
        <p class="text-xs text-slate-400 mt-0.5">
          LLM-Powered Ticket Triage, Automated Incident Diagnostics & RAG Knowledge Retrieval
        </p>
      </div>

      <div class="flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-xs font-mono text-indigo-300">
        <span class="w-2 h-2 rounded-full bg-indigo-400 animate-pulse" />
        <span>Port: 8088 / Qdrant RAG</span>
      </div>
    </div>

    <!-- Chat Messages Window -->
    <div class="flex-1 overflow-y-auto p-6 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
      <div
        v-for="(msg, i) in messages"
        :key="i"
        class="flex items-start gap-3"
        :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <div
          v-if="msg.role === 'assistant'"
          class="w-8 h-8 rounded-xl bg-gradient-to-tr from-indigo-600 to-purple-600 flex items-center justify-center text-white shrink-0 shadow-md"
        >
          <UIcon
            name="i-lucide-sparkles"
            class="w-4 h-4"
          />
        </div>

        <div
          class="p-4 rounded-2xl max-w-2xl text-xs md:text-sm leading-relaxed whitespace-pre-line shadow-lg"
          :class="msg.role === 'user'
            ? 'bg-indigo-600 text-white rounded-br-none'
            : 'bg-slate-950/80 border border-slate-800/80 text-slate-200 rounded-bl-none'"
        >
          {{ msg.text }}
        </div>
      </div>
    </div>

    <!-- Chat Input Box -->
    <div class="p-2 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl flex items-center gap-2">
      <input
        v-model="inputQuery"
        type="text"
        placeholder="Ask AI to triage tickets, search IT manuals, or diagnose server errors..."
        class="flex-1 px-4 py-2.5 text-xs md:text-sm bg-transparent text-white placeholder-slate-400 focus:outline-none"
        @keydown.enter="sendMessage"
      >
      <button
        class="px-5 py-2.5 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white text-xs font-semibold shadow-md transition-all"
        @click="sendMessage"
      >
        Send &rarr;
      </button>
    </div>
  </div>
</template>
