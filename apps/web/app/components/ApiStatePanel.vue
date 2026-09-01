<script setup lang="ts">
import type { ApiViewState } from '~/utils/api-view-state'

const props = withDefaults(defineProps<{
  state: ApiViewState
  resource?: string
}>(), {
  resource: 'data'
})

defineEmits<{ retry: [] }>()

const content = computed(() => {
  if (props.state === 'forbidden') {
    return {
      icon: 'i-lucide-shield-x',
      title: 'Access forbidden',
      detail: `Your role is not allowed to view this ${props.resource}.`,
      color: 'rose'
    }
  }
  if (props.state === 'unavailable') {
    return {
      icon: 'i-lucide-cloud-off',
      title: 'Backend unavailable',
      detail: `The ${props.resource} service could not be reached. Existing data has not been interpreted as zero.`,
      color: 'amber'
    }
  }
  return {
    icon: 'i-lucide-inbox',
    title: 'No records found',
    detail: `The ${props.resource} request succeeded, but no records match the current filters.`,
    color: 'slate'
  }
})
</script>

<template>
  <div
    v-if="state === 'forbidden' || state === 'unavailable' || state === 'empty'"
    :data-api-state="state"
    class="p-8 rounded-3xl border text-center space-y-3"
    :class="content.color === 'rose'
      ? 'bg-rose-500/10 border-rose-500/30'
      : content.color === 'amber'
        ? 'bg-amber-500/10 border-amber-500/30'
        : 'bg-slate-900/60 border-slate-800'"
  >
    <UIcon
      :name="content.icon"
      class="w-9 h-9 mx-auto"
      :class="content.color === 'rose' ? 'text-rose-400' : content.color === 'amber' ? 'text-amber-400' : 'text-slate-500'"
    />
    <h3 class="text-sm font-bold text-white">
      {{ content.title }}
    </h3>
    <p class="text-xs text-slate-400 max-w-lg mx-auto">
      {{ content.detail }}
    </p>
    <button
      v-if="state === 'unavailable'"
      class="text-xs font-semibold text-amber-300 underline hover:text-white"
      @click="$emit('retry')"
    >
      Retry
    </button>
  </div>
</template>
