import { useAuthStore } from '~/stores/auth'
import { isRouteAllowedForRole } from '~/utils/route-permissions'

/**
 * Global authentication and RBAC authorization route middleware.
 * Intercepts all page transitions to enforce session validity and role-based access.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  // The access token intentionally exists only in browser memory. Do not rotate
  // the HttpOnly refresh cookie in an internal SSR subrequest because its
  // Set-Cookie response would not reach the browser. All data loading currently
  // happens onMounted, so client-side guarding runs before protected API calls.
  if (import.meta.server) return

  const authStore = useAuthStore()

  // 1. Skip auth check for public routes
  const publicRoutes = ['/login', '/register', '/forgot-password']
  if (publicRoutes.includes(to.path)) {
    // If already logged in and visiting login, redirect to home
    if (authStore.isAuthenticated) {
      return navigateTo('/')
    }
    return
  }

  // 2. If not authenticated, attempt silent session initialization via HttpOnly cookie
  if (!authStore.isAuthenticated && !authStore.isInitialized) {
    await authStore.initSession()
  }

  // 3. If still unauthenticated, redirect to login page with return URL
  if (!authStore.isAuthenticated) {
    return navigateTo({
      path: '/login',
      query: { redirect: to.fullPath !== '/' ? to.fullPath : undefined }
    })
  }

  // 4. Enforce Role-Based Access Control (RBAC) Guard
  const userRole = authStore.role
  if (!isRouteAllowedForRole(to.path, userRole)) {
    console.warn(`Access denied to ${to.path} for role ${userRole}`)
    return navigateTo('/')
  }
})
