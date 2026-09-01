<script setup lang="ts">
import type {
  AIChatMessage,
  AIChatResponse,
  AITicketAnalysis,
  AIRuntimeStatus
} from '~/types'
import { classifyApiError, dataViewState, type ApiViewState } from '~/utils/api-view-state'

definePageMeta({ layout: 'default' })

const api = useApi()
const toast = useToast()

// Active Mode: 'chat' | 'triage' | 'rag-status'
const activeMode = ref<'chat' | 'triage' | 'rag-status'>('chat')
const runtimeStatus = ref<AIRuntimeStatus | null>(null)
const pageState = ref<ApiViewState>('loading')

// Chat State
const inputQuery = ref('')
const isSending = ref(false)
const chatMessages = ref<AIChatMessage[]>([
  {
    role: 'assistant',
    content: `👋 Xin chào! Tôi là **EOMP AI Operations Copilot**.

Tôi có thể hỗ trợ bạn:
- 🔍 **Tra cứu ngữ cảnh kỹ thuật** khi kho vector có sẵn.
- 🏷️ **Đề xuất phân loại Ticket và nguyên nhân cần kiểm tra**.
- ⚡ **Gợi ý hướng xử lý để kỹ thuật viên đánh giá**.

*Câu trả lời và nguồn trích dẫn chỉ xuất hiện sau khi backend trả về kết quả thật.*`,
    timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
])

// Quick Suggestions
const quickPrompts = [
  'How to reset user MFA token?',
  'Cannot connect to VPN Staging Server',
  'PostgreSQL database connection pool exhausted',
  'Standard laptop setup and baseline security'
]

// Auto-Triage State
const triageTitle = ref('')
const triageDescription = ref('')
const isTriaging = ref(false)
const triageResult = ref<AITicketAnalysis | null>(null)

// Send Chat Message
async function sendMessage(overrideQuery?: string) {
  const query = overrideQuery || inputQuery.value
  if (!query.trim() || isSending.value) return

  // Push User Message
  chatMessages.value.push({
    role: 'user',
    content: query,
    timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  })

  inputQuery.value = ''
  isSending.value = true

  // Scroll to bottom
  nextTick(() => {
    const el = document.getElementById('chat-window')
    if (el) el.scrollTop = el.scrollHeight
  })

  try {
    const payload = {
      messages: chatMessages.value.map(m => ({ role: m.role, content: m.content }))
    }

    const res = await api.post<AIChatResponse>('/api/v1/ai/chat', payload).catch(() => null)

    if (res) {
      chatMessages.value.push({
        role: 'assistant',
        content: res.answer,
        citations: res.citations,
        confidence: res.confidence,
        fallback_mode: res.fallback_mode,
        timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      })
    } else {
      toast.add({ title: 'AI service is unavailable', description: 'No response was generated.', color: 'error' })
    }
  } finally {
    isSending.value = false
    nextTick(() => {
      const el = document.getElementById('chat-window')
      if (el) el.scrollTop = el.scrollHeight
    })
  }
}

// Analyze Ticket
async function runAnalyzeTicket() {
  if (!triageTitle.value.trim()) {
    toast.add({ title: 'Validation', description: 'Ticket title cannot be empty', color: 'error' })
    return
  }

  isTriaging.value = true
  try {
    const res = await api.post<AITicketAnalysis>('/api/v1/ai/analyze-ticket', {
      title: triageTitle.value,
      description: triageDescription.value
    }).catch(() => null)

    if (res) {
      triageResult.value = res
      toast.add({ title: 'Triage Complete', description: `AI classified category as ${res.suggested_category} (${(res.confidence * 100).toFixed(0)}% confidence).`, color: 'success' })
    } else {
      triageResult.value = null
      toast.add({ title: 'AI triage is unavailable', color: 'error' })
    }
  } finally {
    isTriaging.value = false
  }
}

// Copy to clipboard
function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  toast.add({ title: 'Copied', description: 'Text copied to clipboard', color: 'success' })
}

// Quick Sample Tickets
function loadSampleTicket(type: string) {
  if (type === 'vpn') {
    triageTitle.value = 'Cannot connect to VPN Staging Server'
    triageDescription.value = 'Remote engineer is unable to reach internal staging services via WireGuard tunnel. Handshake times out every 10 minutes.'
  } else if (type === 'mfa') {
    triageTitle.value = 'Lost phone with Okta Verify - User locked out'
    triageDescription.value = 'Employee got a new replacement phone and cannot receive push notifications or MFA OTP codes.'
  } else if (type === 'db') {
    triageTitle.value = 'HTTP 500 error: PostgreSQL connection pool exhausted'
    triageDescription.value = 'API Gateway reporting server errors with message: remaining connection slots are reserved for non-replication superuser connections.'
  } else if (type === 'laptop') {
    triageTitle.value = 'Provision MacBook Pro 16" M3 for new hire in Engineering'
    triageDescription.value = 'New Senior Full-stack Developer starting next Monday. Needs standard developer baseline tooling and encryption.'
  }
  runAnalyzeTicket()
}

async function loadRuntimeStatus() {
  pageState.value = 'loading'
  try {
    runtimeStatus.value = await api.get<AIRuntimeStatus>('/api/v1/ai/status')
    pageState.value = dataViewState(runtimeStatus.value ? [runtimeStatus.value] : [])
  } catch (err: unknown) {
    runtimeStatus.value = null
    pageState.value = classifyApiError(err)
  }
}

onMounted(loadRuntimeStatus)
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto h-[calc(100vh-120px)] flex flex-col">
    <ApiStatePanel
      v-if="pageState !== 'loading' && pageState !== 'ready'"
      :state="pageState"
      resource="AI runtime"
      @retry="loadRuntimeStatus"
    />
    <!-- Top Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-bot"
            class="w-7 h-7 text-indigo-400"
          />
          AI Operations Copilot & RAG Engine
        </h1>
        <p class="text-xs text-slate-400 mt-0.5">
          Provider-backed ticket triage and knowledge assistance with human review
        </p>
      </div>

      <!-- Mode Switcher Pills -->
      <div class="flex items-center gap-2">
        <div class="flex items-center p-1 bg-slate-900/90 border border-slate-800 rounded-xl">
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
            :class="activeMode === 'chat' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'"
            @click="activeMode = 'chat'"
          >
            <UIcon
              name="i-lucide-message-square"
              class="w-3.5 h-3.5"
            />
            <span>IT Ops Chat</span>
          </button>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
            :class="activeMode === 'triage' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'"
            @click="activeMode = 'triage'"
          >
            <UIcon
              name="i-lucide-sparkles"
              class="w-3.5 h-3.5"
            />
            <span>Ticket Auto-Triage</span>
          </button>
          <button
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all"
            :class="activeMode === 'rag-status' ? 'bg-indigo-600 text-white shadow' : 'text-slate-400 hover:text-white'"
            @click="activeMode = 'rag-status'"
          >
            <UIcon
              name="i-lucide-database"
              class="w-3.5 h-3.5"
            />
            <span>AI Runtime Status</span>
          </button>
        </div>

        <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-xs font-mono text-indigo-300">
          <span
            class="w-2 h-2 rounded-full"
            :class="runtimeStatus ? 'bg-cyan-400' : 'bg-slate-500'"
          />
          <span>{{ runtimeStatus ? `${runtimeStatus.provider} / ${runtimeStatus.service_status}` : 'Status unavailable' }}</span>
        </div>
      </div>
    </div>

    <!-- MODE 1: IT OPS LIVE CHAT -->
    <div
      v-if="activeMode === 'chat'"
      class="flex-1 flex flex-col min-h-0 space-y-3"
    >
      <!-- Quick Prompt Suggestions -->
      <div class="flex items-center gap-2 overflow-x-auto pb-1 shrink-0">
        <span class="text-[11px] text-slate-400 font-semibold flex items-center gap-1 whitespace-nowrap">
          <UIcon
            name="i-lucide-lightbulb"
            class="w-3.5 h-3.5 text-amber-400"
          /> Gợi ý:
        </span>
        <button
          v-for="p in quickPrompts"
          :key="p"
          class="px-3 py-1 rounded-full text-xs font-medium bg-slate-900/80 hover:bg-indigo-600/20 border border-slate-800 hover:border-indigo-500/40 text-slate-300 hover:text-indigo-200 whitespace-nowrap transition-all"
          @click="sendMessage(p)"
        >
          {{ p }}
        </button>
      </div>

      <!-- Messages Window -->
      <div
        id="chat-window"
        class="flex-1 overflow-y-auto p-5 rounded-3xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-5"
      >
        <div
          v-for="(msg, i) in chatMessages"
          :key="i"
          class="flex items-start gap-3.5"
          :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
        >
          <!-- Assistant Avatar -->
          <div
            v-if="msg.role === 'assistant'"
            class="w-8 h-8 rounded-xl bg-gradient-to-tr from-cyan-500 to-indigo-600 flex items-center justify-center text-white shrink-0 shadow-lg shadow-indigo-500/20 mt-1"
          >
            <UIcon
              name="i-lucide-bot"
              class="w-4 h-4 text-cyan-200 animate-pulse"
            />
          </div>

          <!-- Message Container -->
          <div
            class="max-w-3xl space-y-2.5"
            :class="msg.role === 'user' ? 'items-end' : 'items-start'"
          >
            <!-- Bubble -->
            <div
              class="p-4 rounded-2xl text-xs md:text-sm leading-relaxed whitespace-pre-line shadow-xl"
              :class="msg.role === 'user'
                ? 'bg-indigo-600/40 border border-indigo-500/40 text-white rounded-br-none ml-auto'
                : 'bg-slate-900/90 border border-slate-800 text-slate-100 rounded-bl-none'"
            >
              {{ msg.content }}
            </div>

            <!-- Citations & Action Bar for Assistant Messages -->
            <div
              v-if="msg.role === 'assistant'"
              class="space-y-2 px-1"
            >
              <!-- Source References Pill (RAG Grounding) -->
              <div
                v-if="msg.citations && msg.citations.length > 0"
                class="flex flex-wrap items-center gap-1.5"
              >
                <span class="text-[10px] text-slate-400 font-mono flex items-center gap-1">
                  <UIcon
                    name="i-lucide-link"
                    class="w-3 h-3 text-cyan-400"
                  /> RAG Sources:
                </span>
                <span
                  v-for="c in msg.citations"
                  :key="c.article_id"
                  class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-mono bg-cyan-500/10 text-cyan-300 border border-cyan-500/20 hover:bg-cyan-500/20 transition-all cursor-pointer"
                  @click="navigateTo('/knowledge')"
                >
                  <span>{{ c.title }}</span>
                  <span class="text-emerald-400 font-bold">({{ (c.score * 100).toFixed(0) }}%)</span>
                </span>
              </div>

              <!-- Action Buttons & Confidence Score -->
              <div class="flex items-center justify-between text-[11px] text-slate-400 pt-1">
                <div class="flex items-center gap-2">
                  <span
                    v-if="msg.confidence"
                    class="px-2 py-0.5 rounded-full text-[10px] font-mono font-semibold"
                    :class="msg.confidence >= 0.9 ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'"
                  >
                    Confidence: {{ (msg.confidence * 100).toFixed(0) }}%
                  </span>
                  <span
                    v-if="msg.fallback_mode"
                    class="text-[10px] text-amber-400 font-mono"
                  >
                    [Fallback Catalog]
                  </span>
                  <span>{{ msg.timestamp }}</span>
                </div>

                <div class="flex items-center gap-1">
                  <button
                    class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white flex items-center gap-1 transition-all"
                    @click="copyToClipboard(msg.content)"
                  >
                    <UIcon
                      name="i-lucide-copy"
                      class="w-3 h-3"
                    />
                    <span>Copy Solution</span>
                  </button>
                  <button
                    class="px-2.5 py-1 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/40 text-indigo-300 hover:text-white border border-indigo-500/30 flex items-center gap-1 transition-all"
                    @click="navigateTo('/helpdesk')"
                  >
                    <UIcon
                      name="i-lucide-send"
                      class="w-3 h-3"
                    />
                    <span>Apply to Ticket</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Typing Indicator -->
        <div
          v-if="isSending"
          class="flex items-center gap-3"
        >
          <div class="w-8 h-8 rounded-xl bg-slate-800 flex items-center justify-center text-cyan-400">
            <UIcon
              name="i-lucide-loader-2"
              class="w-4 h-4 animate-spin"
            />
          </div>
          <div class="px-4 py-2.5 rounded-2xl bg-slate-900 border border-slate-800 text-xs text-slate-400 font-mono flex items-center gap-2">
            <span>Waiting for the configured AI provider...</span>
          </div>
        </div>
      </div>

      <!-- Chat Input Box -->
      <div class="p-2.5 rounded-2xl bg-slate-900/90 border border-slate-800/90 backdrop-blur-xl flex items-center gap-2 shadow-2xl shrink-0">
        <input
          v-model="inputQuery"
          type="text"
          placeholder="Ask AI to triage tickets, search SOP runbooks, or diagnose server errors... (Ctrl + Enter to send)"
          class="flex-1 px-4 py-2.5 text-xs md:text-sm bg-transparent text-white placeholder-slate-400 focus:outline-none"
          :disabled="isSending"
          @keydown.enter.ctrl="sendMessage()"
          @keydown.enter.exact="sendMessage()"
        >
        <button
          class="px-5 py-2.5 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white text-xs font-semibold shadow-lg shadow-indigo-500/20 disabled:opacity-50 transition-all flex items-center gap-1.5"
          :disabled="isSending || !inputQuery.trim()"
          @click="sendMessage()"
        >
          <span>Send</span>
          <UIcon
            name="i-lucide-send"
            class="w-3.5 h-3.5"
          />
        </button>
      </div>
    </div>

    <!-- MODE 2: TICKET AUTO-TRIAGE STUDIO -->
    <div
      v-if="activeMode === 'triage'"
      class="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-0 overflow-y-auto"
    >
      <!-- Left: Ticket Input & Sample Picker -->
      <div class="lg:col-span-5 space-y-4">
        <div class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
          <div class="flex items-center justify-between">
            <h2 class="text-sm font-bold text-white flex items-center gap-2">
              <UIcon
                name="i-lucide-file-text"
                class="w-4 h-4 text-cyan-400"
              />
              Nhập Thông Tin Sự Cố (Ticket Details)
            </h2>
          </div>

          <!-- Quick Samples -->
          <div class="space-y-1.5">
            <span class="text-[11px] text-slate-400 font-semibold">Tải kịch bản mẫu:</span>
            <div class="grid grid-cols-2 gap-2 text-[11px]">
              <button
                class="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700 text-slate-300 hover:text-cyan-300 text-left transition-all"
                @click="loadSampleTicket('vpn')"
              >
                🌐 VPN Timeout (Network)
              </button>
              <button
                class="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700 text-slate-300 hover:text-cyan-300 text-left transition-all"
                @click="loadSampleTicket('mfa')"
              >
                🛡️ MFA Reset (Security)
              </button>
              <button
                class="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700 text-slate-300 hover:text-cyan-300 text-left transition-all"
                @click="loadSampleTicket('db')"
              >
                🗄️ DB Pool Crash (DevOps)
              </button>
              <button
                class="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-800 border border-slate-700 text-slate-300 hover:text-cyan-300 text-left transition-all"
                @click="loadSampleTicket('laptop')"
              >
                💻 Laptop Setup (Hardware)
              </button>
            </div>
          </div>

          <div class="space-y-3 text-xs">
            <div>
              <label class="block text-slate-300 font-semibold mb-1">Tiêu Đề Ticket (Title) *</label>
              <input
                v-model="triageTitle"
                type="text"
                class="w-full px-3 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500"
              >
            </div>

            <div>
              <label class="block text-slate-300 font-semibold mb-1">Mô Tả Lỗi (Description) *</label>
              <textarea
                v-model="triageDescription"
                rows="5"
                class="w-full px-3 py-2.5 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-cyan-500 leading-relaxed"
              />
            </div>

            <button
              class="w-full py-3 rounded-xl bg-gradient-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 text-white font-semibold shadow-lg shadow-indigo-500/20 flex items-center justify-center gap-2 transition-all disabled:opacity-50"
              :disabled="isTriaging"
              @click="runAnalyzeTicket"
            >
              <UIcon
                name="i-lucide-sparkles"
                class="w-4 h-4"
              />
              <span>{{ isTriaging ? 'Đang Phân Tích RAG...' : 'Tự Động Phân Loại Với AI Copilot' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Right: AI Analysis Results Card -->
      <div class="lg:col-span-7 space-y-4">
        <div
          v-if="triageResult"
          class="p-6 rounded-2xl bg-slate-900/80 border border-slate-800/80 backdrop-blur-xl space-y-5"
        >
          <!-- Header Bar -->
          <div class="flex items-start justify-between gap-4 border-b border-slate-800 pb-4">
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <span class="text-xs font-mono font-bold px-2.5 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                  {{ triageResult.ticket_id }}
                </span>
                <span class="text-xs text-slate-400">AI Diagnostic Report</span>
              </div>
              <h3 class="text-base font-bold text-white">
                {{ triageResult.summary }}
              </h3>
            </div>

            <div class="text-right">
              <span class="px-3 py-1 rounded-full text-xs font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                Confidence: {{ (triageResult.confidence * 100).toFixed(0) }}%
              </span>
            </div>
          </div>

          <!-- Classification Badges -->
          <div class="grid grid-cols-2 gap-4">
            <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-1">
              <span class="text-[11px] text-slate-400">Danh Mục Gợi Ý (Suggested Category):</span>
              <p class="text-sm font-bold text-cyan-300">
                {{ triageResult.suggested_category }}
              </p>
            </div>

            <div class="p-3 rounded-xl bg-slate-950/80 border border-slate-800 space-y-1">
              <span class="text-[11px] text-slate-400">Độ Ưu Tiên Khuyến Nghị (Priority):</span>
              <div class="flex items-center gap-2">
                <span
                  class="px-2.5 py-0.5 rounded text-xs font-bold font-mono"
                  :class="triageResult.priority === 'URGENT' || triageResult.priority === 'HIGH' ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'"
                >
                  {{ triageResult.priority }}
                </span>
              </div>
            </div>
          </div>

          <!-- Root Cause Probability -->
          <div class="p-4 rounded-xl bg-indigo-950/30 border border-indigo-500/30 space-y-1.5 text-xs">
            <span class="font-bold text-indigo-300 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-activity"
                class="w-4 h-4"
              />
              Nguyên Nhân Gốc Rễ Dự Đoán (Root Cause Probability):
            </span>
            <p class="text-slate-200 leading-relaxed">
              {{ triageResult.root_cause }}
            </p>
          </div>

          <!-- Suggested Resolution -->
          <div class="space-y-2 text-xs">
            <span class="font-bold text-slate-200 flex items-center gap-1.5">
              <UIcon
                name="i-lucide-check-circle"
                class="w-4 h-4 text-emerald-400"
              />
              Kịch Bản Xử Lý Khuyến Nghị (Suggested SOP Resolution):
            </span>
            <div class="p-4 rounded-xl bg-slate-950/80 border border-slate-800 text-slate-300 font-mono whitespace-pre-line leading-relaxed">
              {{ triageResult.suggested_resolution }}
            </div>
          </div>

          <!-- Citations Grounding -->
          <div
            v-if="triageResult.citations && triageResult.citations.length > 0"
            class="space-y-2 text-xs"
          >
            <span class="font-bold text-slate-300">Nguồn do AI Service trả về:</span>
            <div class="space-y-1.5">
              <div
                v-for="c in triageResult.citations"
                :key="c.article_id"
                class="p-2.5 rounded-xl bg-slate-950 border border-slate-800 flex items-center justify-between gap-3 text-xs"
              >
                <div class="flex items-center gap-2">
                  <UIcon
                    :name="c.type === 'runbook' ? 'i-lucide-terminal' : 'i-lucide-file-text'"
                    class="w-4 h-4 text-cyan-400"
                  />
                  <span class="text-white font-medium">{{ c.title }}</span>
                </div>
                <span class="text-[10px] font-mono text-emerald-400 font-bold px-2 py-0.5 rounded bg-emerald-500/10 border border-emerald-500/20">
                  Similarity: {{ (c.score * 100).toFixed(0) }}%
                </span>
              </div>
            </div>
          </div>

          <!-- Safety Policy Alert -->
          <div class="p-3 rounded-xl bg-slate-950 border border-slate-800 text-[11px] text-slate-400 flex items-center justify-between">
            <span class="flex items-center gap-1.5 text-amber-300 font-semibold">
              <UIcon
                name="i-lucide-shield-alert"
                class="w-4 h-4"
              />
              Rule: Requires Human Review (AI can never execute destructive actions autonomously)
            </span>
            <button
              class="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs shadow-md transition-all"
              @click="navigateTo('/helpdesk')"
            >
              Áp Dụng Vào Helpdesk
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- MODE 3: OBSERVED AI RUNTIME STATUS -->
    <div
      v-if="activeMode === 'rag-status'"
      class="flex-1 space-y-6 overflow-y-auto"
    >
      <div class="grid grid-cols-1 md:grid-cols-3 gap-5">
        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs text-slate-400 font-semibold">Qdrant Vector Store</span>
            <span
              class="w-2.5 h-2.5 rounded-full"
              :class="runtimeStatus?.qdrant_status === 'ONLINE' ? 'bg-emerald-400' : 'bg-rose-400'"
            />
          </div>
          <p class="text-xl font-bold text-white">
            {{ runtimeStatus?.qdrant_status ?? 'Unavailable' }}
          </p>
          <span class="text-xs text-slate-400">Collection: <code class="text-cyan-400">{{ runtimeStatus?.qdrant_collection ?? 'unknown' }}</code></span>
        </div>

        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs text-slate-400 font-semibold">Configured Provider</span>
            <UIcon
              name="i-lucide-cpu"
              class="w-5 h-5 text-indigo-400"
            />
          </div>
          <p class="text-xl font-bold text-white">
            {{ runtimeStatus?.provider ?? 'Unavailable' }}
          </p>
          <span class="text-xs text-slate-400">Model: <code class="text-indigo-400">{{ runtimeStatus?.model ?? 'unknown' }}</code> · Embedding: <code class="text-indigo-400">{{ runtimeStatus?.embedding_model ?? 'unknown' }}</code></span>
        </div>

        <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs text-slate-400 font-semibold">Explicit Mock Fallback</span>
            <UIcon
              name="i-lucide-shield-check"
              class="w-5 h-5 text-emerald-400"
            />
          </div>
          <p class="text-xl font-bold text-emerald-400">
            {{ runtimeStatus ? (runtimeStatus.mock_fallback_enabled ? 'Enabled' : 'Disabled') : 'Unavailable' }}
          </p>
          <span class="text-xs text-slate-400">Controlled by provider choice and ALLOW_MOCK_AI</span>
        </div>
      </div>

      <div class="p-6 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl space-y-4">
        <h3 class="text-sm font-bold text-white">
          Runtime Evidence
        </h3>
        <div class="space-y-2 text-xs">
          <div class="p-3 rounded-xl bg-slate-950 border border-slate-800 text-slate-300">
            Last checked: <span class="font-mono text-cyan-400">{{ runtimeStatus?.last_checked_at ? new Date(runtimeStatus.last_checked_at).toLocaleString() : 'Unavailable' }}</span>
          </div>
          <div class="p-3 rounded-xl bg-slate-950 border border-slate-800 text-slate-400">
            Automatic ingestion: <strong class="text-slate-200">{{ runtimeStatus ? (runtimeStatus.auto_ingest_enabled ? 'Enabled' : 'Disabled') : 'Unavailable' }}</strong>.
            Indexed document counts are not exposed by the runtime endpoint, so no synthetic totals are displayed.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
