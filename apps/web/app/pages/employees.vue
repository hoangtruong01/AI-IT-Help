<script setup lang="ts">
import type { Employee, Department, PaginatedResponse, CreateEmployeePayload } from '~/types'

definePageMeta({ layout: 'default' })

const api = useApi()

// State
const employees = ref<Employee[]>([])
const departments = ref<Department[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')
const selectedDepartment = ref('')
const selectedStatus = ref('')

// Modal state
const isAddModalOpen = ref(false)
const submitting = ref(false)
const formError = ref<string | null>(null)
const newEmployee = reactive<CreateEmployeePayload>({
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  job_title: '',
  department_id: '',
  status: 'ACTIVE',
  location: 'Headquarters (Building A)'
})

async function fetchDepartments() {
  try {
    const res = await api.get<Department[]>('/api/v1/departments')
    departments.value = res || []
  } catch (err: unknown) {
    console.error('Failed to load departments:', err)
  }
}

async function fetchEmployees() {
  loading.value = true
  error.value = null
  try {
    const params: Record<string, unknown> = {
      page: 1,
      page_size: 50
    }
    if (searchQuery.value) params.search = searchQuery.value
    if (selectedDepartment.value) params.department_id = selectedDepartment.value
    if (selectedStatus.value) params.status = selectedStatus.value

    const res = await api.get<PaginatedResponse<Employee>>('/api/v1/employees', params)
    employees.value = res?.data || []
  } catch (err: unknown) {
    const errorObj = err as { data?: { error?: { message?: string } }, message?: string }
    error.value = errorObj?.data?.error?.message || errorObj?.message || 'Failed to load employees from API Gateway.'
  } finally {
    loading.value = false
  }
}

async function handleCreateEmployee() {
  if (!newEmployee.first_name || !newEmployee.last_name || !newEmployee.email || !newEmployee.job_title) {
    formError.value = 'Please fill in all required fields.'
    return
  }

  submitting.value = true
  formError.value = null
  try {
    await api.post<Employee>('/api/v1/employees', { ...newEmployee })
    isAddModalOpen.value = false
    // Reset form
    Object.assign(newEmployee, {
      first_name: '',
      last_name: '',
      email: '',
      phone: '',
      job_title: '',
      department_id: departments.value[0]?.id || '',
      status: 'ACTIVE',
      location: 'Headquarters (Building A)'
    })
    await fetchEmployees()
  } catch (err: unknown) {
    const errorObj = err as { data?: { error?: { message?: string } }, message?: string }
    formError.value = errorObj?.data?.error?.message || errorObj?.message || 'Failed to create employee.'
  } finally {
    submitting.value = false
  }
}

// Watch filters with debounce
let debounceTimer: ReturnType<typeof setTimeout> | null = null
watch([searchQuery, selectedDepartment, selectedStatus], () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    fetchEmployees()
  }, 250)
})

onMounted(() => {
  fetchDepartments()
  fetchEmployees()
})
</script>

<template>
  <div class="space-y-6 max-w-7xl mx-auto">
    <!-- Header -->
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
          Live Directory & Org Structure · Synced with Employee Microservice (:8082) & PostgreSQL
        </p>
      </div>

      <button
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white text-xs font-semibold shadow-lg shadow-blue-500/20 hover:scale-105 transition-all"
        @click="isAddModalOpen = true"
      >
        <UIcon
          name="i-lucide-user-plus"
          class="w-4 h-4"
        />
        <span>+ Add Employee</span>
      </button>
    </div>

    <!-- Filters Bar -->
    <div class="p-4 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl flex flex-col sm:flex-row items-center justify-between gap-3">
      <!-- Search Input -->
      <div class="relative w-full sm:w-80">
        <UIcon
          name="i-lucide-search"
          class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
        />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by name, email, role..."
          class="w-full pl-9 pr-4 py-2 text-xs rounded-xl bg-slate-950/80 border border-slate-800 text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
        >
      </div>

      <!-- Department & Status Selects -->
      <div class="flex items-center gap-2.5 w-full sm:w-auto">
        <select
          v-model="selectedDepartment"
          class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-blue-500"
        >
          <option value="">
            All Departments
          </option>
          <option
            v-for="dept in departments"
            :key="dept.id"
            :value="dept.id"
          >
            {{ dept.name }}
          </option>
        </select>

        <select
          v-model="selectedStatus"
          class="px-3 py-2 rounded-xl bg-slate-950/80 border border-slate-800 text-xs text-slate-200 focus:outline-none focus:border-blue-500"
        >
          <option value="">
            All Statuses
          </option>
          <option value="ACTIVE">
            Active
          </option>
          <option value="ON_LEAVE">
            On Leave
          </option>
          <option value="PROBATION">
            Probation
          </option>
        </select>

        <button
          class="p-2 rounded-xl bg-slate-800/60 hover:bg-slate-800 text-slate-300 transition-colors"
          title="Refresh"
          @click="fetchEmployees"
        >
          <UIcon
            name="i-lucide-refresh-cw"
            class="w-4 h-4"
            :class="{ 'animate-spin': loading }"
          />
        </button>
      </div>
    </div>

    <!-- Error Alert -->
    <div
      v-if="error"
      class="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs flex items-center justify-between"
    >
      <div class="flex items-center gap-2">
        <UIcon
          name="i-lucide-alert-triangle"
          class="w-4 h-4 text-rose-400"
        />
        <span>{{ error }}</span>
      </div>
      <button
        class="underline hover:text-white"
        @click="fetchEmployees"
      >
        Retry
      </button>
    </div>

    <!-- Loading Skeleton -->
    <div
      v-if="loading && employees.length === 0"
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
    >
      <div
        v-for="i in 6"
        :key="i"
        class="p-5 rounded-2xl bg-slate-900/40 border border-slate-800 animate-pulse space-y-3"
      >
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-slate-800" />
          <div class="space-y-1.5 flex-1">
            <div class="w-24 h-3.5 bg-slate-800 rounded" />
            <div class="w-36 h-2.5 bg-slate-800/60 rounded" />
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="employees.length === 0 && !loading"
      class="p-12 text-center rounded-3xl bg-slate-900/40 border border-slate-800/80 space-y-3"
    >
      <UIcon
        name="i-lucide-users"
        class="w-10 h-10 text-slate-600 mx-auto"
      />
      <h3 class="text-sm font-bold text-slate-300">
        No employees found
      </h3>
      <p class="text-xs text-slate-500 max-w-sm mx-auto">
        No employee records match your search criteria. Try adjusting filters or click "+ Add Employee" to create one.
      </p>
    </div>

    <!-- Cards Grid -->
    <div
      v-else
      class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
    >
      <div
        v-for="emp in employees"
        :key="emp.id"
        class="p-5 rounded-2xl bg-slate-900/60 border border-slate-800/80 backdrop-blur-xl hover:border-blue-500/40 transition-all group relative"
      >
        <div class="flex items-start justify-between gap-3 mb-3">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center font-bold text-white text-sm shadow-md">
              {{ emp.first_name[0] }}{{ emp.last_name[0] }}
            </div>
            <div>
              <h3 class="text-sm font-bold text-white group-hover:text-blue-300 transition-colors">
                {{ emp.full_name }}
              </h3>
              <p class="text-xs text-slate-400">
                {{ emp.job_title }}
              </p>
            </div>
          </div>
          <span
            class="text-[10px] font-semibold px-2 py-0.5 rounded-full border"
            :class="emp.status === 'ACTIVE' ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30' : 'bg-amber-500/10 text-amber-300 border-amber-500/30'"
          >
            {{ emp.status }}
          </span>
        </div>

        <div class="space-y-1.5 text-xs text-slate-400 pt-2 border-t border-slate-800/60">
          <div class="flex items-center justify-between">
            <span class="font-mono text-[11px] text-blue-400">{{ emp.department_name || 'Unassigned' }}</span>
            <span class="text-[11px] text-slate-500">{{ emp.location }}</span>
          </div>
          <div class="flex items-center gap-1.5 text-slate-300 truncate">
            <UIcon
              name="i-lucide-mail"
              class="w-3.5 h-3.5 text-slate-500 shrink-0"
            />
            <span class="text-[11px]">{{ emp.email }}</span>
          </div>
          <div
            v-if="emp.phone"
            class="flex items-center gap-1.5 text-slate-400"
          >
            <UIcon
              name="i-lucide-phone"
              class="w-3.5 h-3.5 text-slate-500 shrink-0"
            />
            <span class="text-[11px] font-mono">{{ emp.phone }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Employee Modal -->
    <div
      v-if="isAddModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-3xl bg-slate-900 border border-slate-800 shadow-2xl p-6 space-y-5 animate-scale-up">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <UIcon
              name="i-lucide-user-plus"
              class="w-5 h-5 text-blue-400"
            />
            Add New Employee Profile
          </h3>
          <button
            class="text-slate-400 hover:text-white"
            @click="isAddModalOpen = false"
          >
            <UIcon
              name="i-lucide-x"
              class="w-5 h-5"
            />
          </button>
        </div>

        <div
          v-if="formError"
          class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs"
        >
          {{ formError }}
        </div>

        <form
          class="space-y-4 text-xs"
          @submit.prevent="handleCreateEmployee"
        >
          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">First Name *</label>
              <input
                v-model="newEmployee.first_name"
                type="text"
                required
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Last Name *</label>
              <input
                v-model="newEmployee.last_name"
                type="text"
                required
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Email Address *</label>
              <input
                v-model="newEmployee.email"
                type="email"
                required
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Phone Number</label>
              <input
                v-model="newEmployee.phone"
                type="text"
                placeholder="+84 901 ..."
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Job Title *</label>
              <input
                v-model="newEmployee.job_title"
                type="text"
                required
                placeholder="Software Engineer..."
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Department</label>
              <select
                v-model="newEmployee.department_id"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="">
                  Select department
                </option>
                <option
                  v-for="d in departments"
                  :key="d.id"
                  :value="d.id"
                >
                  {{ d.name }}
                </option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Location</label>
              <input
                v-model="newEmployee.location"
                type="text"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
            </div>
            <div class="space-y-1">
              <label class="font-semibold text-slate-300">Status</label>
              <select
                v-model="newEmployee.status"
                class="w-full px-3 py-2 rounded-xl bg-slate-950 border border-slate-800 text-white focus:outline-none focus:border-blue-500"
              >
                <option value="ACTIVE">
                  ACTIVE
                </option>
                <option value="PROBATION">
                  PROBATION
                </option>
                <option value="ON_LEAVE">
                  ON_LEAVE
                </option>
              </select>
            </div>
          </div>

          <div class="pt-3 flex items-center justify-end gap-3 border-t border-slate-800">
            <button
              type="button"
              class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 font-semibold"
              @click="isAddModalOpen = false"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="submitting"
              class="px-5 py-2 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white font-semibold flex items-center gap-2 shadow-lg shadow-blue-500/25 disabled:opacity-50"
            >
              <UIcon
                v-if="submitting"
                name="i-lucide-loader-2"
                class="w-4 h-4 animate-spin"
              />
              <span>{{ submitting ? 'Saving...' : 'Save Employee' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
