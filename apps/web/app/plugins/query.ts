import { VueQueryPlugin, QueryClient } from '@tanstack/vue-query'

export default defineNuxtPlugin((nuxtApp) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 60 * 5, // 5 minutes
        retry: 1,
        refetchOnWindowFocus: false
      }
    }
  })

  nuxtApp.vueApp.use(VueQueryPlugin, { queryClient })

  return {
    provide: {
      queryClient
    }
  }
})
