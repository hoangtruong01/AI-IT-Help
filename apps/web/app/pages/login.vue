<script setup lang="ts">
definePageMeta({
  layout: false
})

const authStore = useAuthStore()
const route = useRoute()

const email = ref('')
const password = ref('')
const showPassword = ref(false)

async function handleLogin() {
  if (!email.value || !password.value) return
  const success = await authStore.login(email.value, password.value)
  if (success) {
    await navigateTo(safeLocalRedirect(route.query.redirect))
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-100 flex flex-col justify-center items-center p-4 relative overflow-hidden selection:bg-indigo-500 selection:text-white">
    <!-- Ambient Background Glows -->
    <div class="absolute -top-40 -left-40 w-96 h-96 bg-indigo-600/20 rounded-full blur-3xl pointer-events-none" />
    <div class="absolute -bottom-40 -right-40 w-96 h-96 bg-purple-600/20 rounded-full blur-3xl pointer-events-none" />

    <!-- Main Login Card -->
    <div class="w-full max-w-md relative z-10">
      <!-- Brand Header -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-gradient-to-tr from-indigo-600 to-purple-600 shadow-xl shadow-indigo-500/20 mb-4 border border-indigo-400/30">
          <UIcon
            name="i-lucide-shield-check"
            class="w-7 h-7 text-white"
          />
        </div>
        <h1 class="text-2xl font-black tracking-tight text-white flex items-center justify-center gap-2">
          EOMP <span class="text-xs font-semibold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">ENTERPRISE</span>
        </h1>
        <p class="text-xs text-slate-400 mt-1.5">
          Enterprise Operations Management & ITSM Platform
        </p>
      </div>

      <!-- Card Box -->
      <div class="p-8 rounded-3xl bg-slate-900/70 border border-slate-800/80 backdrop-blur-2xl shadow-2xl shadow-black/60 space-y-6">
        <div>
          <h2 class="text-lg font-bold text-white">
            Sign In to Workspace
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            Enter your credentials to access operations dashboard
          </p>
        </div>

        <!-- Error Message Alert -->
        <div
          v-if="authStore.error"
          class="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center gap-2.5 animate-shake"
        >
          <UIcon
            name="i-lucide-alert-circle"
            class="w-4 h-4 shrink-0 text-rose-400"
          />
          <span>{{ authStore.error }}</span>
        </div>

        <form
          class="space-y-4"
          @submit.prevent="handleLogin"
        >
          <!-- Email Input -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-slate-300">Email Address</label>
            <div class="relative">
              <UIcon
                name="i-lucide-mail"
                class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                v-model="email"
                type="email"
                required
                placeholder="name@company.com"
                class="w-full pl-10 pr-4 py-2.5 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
              >
            </div>
          </div>

          <!-- Password Input -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <label class="text-xs font-semibold text-slate-300">Password</label>
              <a
                href="#"
                class="text-[11px] text-indigo-400 hover:text-indigo-300 transition-colors"
              >Forgot?</a>
            </div>
            <div class="relative">
              <UIcon
                name="i-lucide-lock"
                class="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400"
              />
              <input
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                required
                placeholder="••••••••"
                class="w-full pl-10 pr-10 py-2.5 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 transition-all"
              >
              <button
                type="button"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200 text-xs"
                @click="showPassword = !showPassword"
              >
                <UIcon
                  :name="showPassword ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  class="w-4 h-4"
                />
              </button>
            </div>
          </div>

          <!-- Submit Button -->
          <button
            type="submit"
            :disabled="authStore.loading"
            class="w-full mt-2 py-3 px-4 rounded-xl bg-gradient-to-r from-indigo-600 via-indigo-500 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white text-xs font-bold shadow-lg shadow-indigo-500/25 transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed hover:scale-[1.01]"
          >
            <UIcon
              v-if="authStore.loading"
              name="i-lucide-loader-2"
              class="w-4 h-4 animate-spin"
            />
            <span v-else>Sign In to Account</span>
            <UIcon
              v-if="!authStore.loading"
              name="i-lucide-arrow-right"
              class="w-4 h-4"
            />
          </button>
        </form>
      </div>

      <!-- Footer Info -->
      <p class="text-center text-[11px] text-slate-500 mt-6">
        Protected by EOMP JWT Auth & RBAC Security Engine &copy; 2026
      </p>
    </div>
  </div>
</template>
