/**
 * Auth middleware skeleton.
 * Redirects unauthenticated users to login page.
 */
export default defineNuxtRouteMiddleware((to) => {
  const authStore = useAuthStore()

  // Skip auth check for public routes
  const publicRoutes = ['/login', '/register', '/forgot-password']
  if (publicRoutes.includes(to.path)) {
    return
  }

  if (!authStore.isAuthenticated) {
    return navigateTo('/login')
  }
})
