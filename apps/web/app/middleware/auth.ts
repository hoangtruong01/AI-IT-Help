/**
 * Auth middleware skeleton.
 * Redirects unauthenticated users to login page.
 */
export default defineNuxtRouteMiddleware((to) => {
  const _authStore = useAuthStore()

  // Skip auth check for public routes
  const publicRoutes = ['/login', '/register', '/forgot-password']
  if (publicRoutes.includes(to.path)) {
    return
  }

  // TODO: Enable when auth is implemented
  // if (!authStore.isAuthenticated) {
  //   return navigateTo('/login')
  // }
})
