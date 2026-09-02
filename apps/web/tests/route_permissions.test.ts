import { describe, it, expect } from 'vitest'
import { isRouteAllowedForRole, safeLocalRedirect } from '../app/utils/route-permissions'

// Fast route-policy contracts executed in Vitest. They do not launch
// a browser and must not be counted as Playwright/browser E2E evidence.
describe('Route Permission Contracts', () => {
  describe('Journey 1: Employee Authentication & Access Boundary', () => {
    const employeeRole = 'ROLE_EMPLOYEE'

    it('permits Employee to access Self-Service Helpdesk, Workflows, Knowledge Base, and AI Assistant', () => {
      const allowedPaths = ['/', '/helpdesk', '/workflows', '/knowledge', '/ai', '/employees']
      allowedPaths.forEach((path) => {
        expect(isRouteAllowedForRole(path, employeeRole)).toBe(true)
      })
    })

    it('strictly blocks Employee from administrative and operator routes', () => {
      const blockedPaths = [
        '/audit',
        '/audit/security',
        '/reports',
        '/reports/export',
        '/monitoring',
        '/monitoring/alerts',
        '/changes',
        '/problems',
        '/assets'
      ]
      blockedPaths.forEach((path) => {
        expect(isRouteAllowedForRole(path, employeeRole)).toBe(false)
      })
    })

    it('sanitizes post-login redirect query parameter against open redirect attacks', () => {
      expect(safeLocalRedirect('/helpdesk')).toBe('/helpdesk')
      expect(safeLocalRedirect('/knowledge/article-123')).toBe('/knowledge/article-123')
      // Malicious targets must default to root
      expect(safeLocalRedirect('https://attacker.evil/steal')).toBe('/')
      expect(safeLocalRedirect('//evil.example/login')).toBe('/')
      expect(safeLocalRedirect('/\\evil.example')).toBe('/')
      expect(safeLocalRedirect('javascript:alert(1)')).toBe('/')
    })
  })

  describe('Journey 2: IT Agent Operational Access', () => {
    const agentRole = 'ROLE_AGENT'

    it('grants Agent access to Problems and Assets management', () => {
      expect(isRouteAllowedForRole('/problems', agentRole)).toBe(true)
      expect(isRouteAllowedForRole('/assets', agentRole)).toBe(true)
      expect(isRouteAllowedForRole('/helpdesk', agentRole)).toBe(true)
    })

    it('denies Agent from executive reporting, compliance audit, and monitoring telemetry', () => {
      expect(isRouteAllowedForRole('/audit', agentRole)).toBe(false)
      expect(isRouteAllowedForRole('/reports', agentRole)).toBe(false)
      expect(isRouteAllowedForRole('/monitoring', agentRole)).toBe(false)
      expect(isRouteAllowedForRole('/changes', agentRole)).toBe(false)
    })
  })

  describe('Journey 3: Department Manager & SecOps Admin Access', () => {
    it('grants Manager access to Reports, Monitoring, and Change Management', () => {
      const managerRole = 'ROLE_MANAGER'
      expect(isRouteAllowedForRole('/reports', managerRole)).toBe(true)
      expect(isRouteAllowedForRole('/monitoring', managerRole)).toBe(true)
      expect(isRouteAllowedForRole('/changes', managerRole)).toBe(true)
      expect(isRouteAllowedForRole('/audit', managerRole)).toBe(true)
    })

    it('grants Admin complete access to all system routes', () => {
      const adminRole = 'ROLE_ADMIN'
      const allRoutes = [
        '/',
        '/helpdesk',
        '/workflows',
        '/knowledge',
        '/ai',
        '/employees',
        '/problems',
        '/assets',
        '/audit',
        '/reports',
        '/monitoring',
        '/changes'
      ]
      allRoutes.forEach((route) => {
        expect(isRouteAllowedForRole(route, adminRole)).toBe(true)
      })
    })
  })

  describe('Journey 4: Optimistic Locking & Conflict State Handling', () => {
    interface TicketState {
      id: string
      version: number
      status: string
    }

    it('models optimistic locking conflict detection and client-side retry state', () => {
      const serverTicket: TicketState = { id: 'TK-1001', version: 2, status: 'IN_PROGRESS' }
      const clientPayload = { id: 'TK-1001', expectedVersion: 1, status: 'RESOLVED' }

      // Conflict check
      const hasConflict = serverTicket.version !== clientPayload.expectedVersion
      expect(hasConflict).toBe(true)

      // UI state transition upon 409
      const uiState = {
        showConflictBanner: hasConflict,
        latestServerVersion: serverTicket.version,
        canRetry: true
      }
      expect(uiState.showConflictBanner).toBe(true)
      expect(uiState.latestServerVersion).toBe(2)

      // Client accepts refreshed version and retries
      const retriedPayload = { ...clientPayload, expectedVersion: serverTicket.version }
      const retrySuccess = serverTicket.version === retriedPayload.expectedVersion
      expect(retrySuccess).toBe(true)
    })
  })
})
