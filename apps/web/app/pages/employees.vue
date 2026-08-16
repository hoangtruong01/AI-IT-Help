<script setup lang="ts">
definePageMeta({ layout: 'default' })

const employees = [
  { name: 'Alex Nguyen', role: 'Staff Software Engineer', department: 'Engineering', email: 'alex.nguyen@eomp.local', status: 'Active', devices: 3 },
  { name: 'Sarah Tran', role: 'Product Operations Lead', department: 'Product', email: 'sarah.tran@eomp.local', status: 'Active', devices: 2 },
  { name: 'David Tran', role: 'Senior DevOps Specialist', department: 'Infrastructure', email: 'david.tran@eomp.local', status: 'Active', devices: 4 },
  { name: 'Emily Davis', role: 'Lead UI/UX Designer', department: 'Design', email: 'emily.davis@eomp.local', status: 'Active', devices: 2 },
  { name: 'Michael Chang', role: 'Financial Controller', department: 'Finance', email: 'michael.chang@eomp.local', status: 'On Leave', devices: 1 }
]
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-2xl font-extrabold text-white flex items-center gap-2.5">
          <UIcon
            name="i-lucide-users"
            class="w-6 h-6 text-blue-400"
          />
          Organization & Employees
        </h1>
        <p class="text-xs text-slate-400 mt-1">
          1,248 Members across 8 Departments · Org Structure & Permission Assignment
        </p>
      </div>

      <button class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-xs font-semibold shadow-lg shadow-blue-500/20 hover:scale-105 transition-all">
        <UIcon
          name="i-lucide-user-plus"
          class="w-4 h-4"
        />
        <span>+ Add Employee</span>
      </button>
    </div>

    <!-- Cards Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="emp in employees"
        :key="emp.email"
        class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-blue-500/40 transition-all group"
      >
        <div class="flex items-start justify-between gap-3 mb-3">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center font-bold text-white text-sm shadow-md">
              {{ emp.name.split(' ').map(n => n[0]).join('') }}
            </div>
            <div>
              <h3 class="text-sm font-bold text-white group-hover:text-blue-300 transition-colors">
                {{ emp.name }}
              </h3>
              <p class="text-xs text-slate-400">
                {{ emp.role }}
              </p>
            </div>
          </div>
          <span
            class="text-[10px] font-semibold px-2 py-0.5 rounded-full border"
            :class="emp.status === 'Active' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-amber-500/10 text-amber-300 border-amber-500/30'"
          >
            {{ emp.status }}
          </span>
        </div>

        <div class="pt-3 border-t border-slate-800/60 flex items-center justify-between text-xs text-slate-400">
          <span class="font-mono text-[11px]">{{ emp.department }}</span>
          <span class="flex items-center gap-1 text-slate-300">
            <UIcon
              name="i-lucide-laptop"
              class="w-3.5 h-3.5 text-blue-400"
            />
            {{ emp.devices }} Assets
          </span>
        </div>
      </div>
    </div>
  </div>
</template>
