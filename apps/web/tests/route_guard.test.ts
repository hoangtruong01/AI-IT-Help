import { describe, it, expect } from 'vitest'
import { isRouteAllowedForRole, safeLocalRedirect } from '../app/utils/route-permissions'

describe('Route Guard RBAC Matrix', () => {
  const roles = ['ROLE_EMPLOYEE', 'ROLE_AGENT', 'ROLE_MANAGER', 'ROLE_ADMIN']

  describe('Public / General Authenticated Routes', () => {
    const generalRoutes = ['/', '/helpdesk', '/workflows', '/knowledge', '/ai', '/employees']

    generalRoutes.forEach((route) => {
      roles.forEach((role) => {
        it(`allows ${role} to access ${route}`, () => {
          expect(isRouteAllowedForRole(route, role)).toBe(true)
        })
      })
    })
  })

  describe('Admin & Manager Restricted Routes (/audit, /reports, /monitoring, /changes)', () => {
    const restrictedRoutes = ['/audit', '/reports', '/monitoring', '/changes']

    restrictedRoutes.forEach((route) => {
      it(`allows ROLE_ADMIN to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_ADMIN')).toBe(true)
      })

      it(`allows ROLE_MANAGER to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_MANAGER')).toBe(true)
      })

      it(`denies ROLE_AGENT to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_AGENT')).toBe(false)
      })

      it(`denies ROLE_EMPLOYEE to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_EMPLOYEE')).toBe(false)
      })
    })
  })

  describe('Operator Restricted Routes (/problems, /assets)', () => {
    const operatorRoutes = ['/problems', '/assets']

    operatorRoutes.forEach((route) => {
      it(`allows ROLE_ADMIN to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_ADMIN')).toBe(true)
      })

      it(`allows ROLE_MANAGER to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_MANAGER')).toBe(true)
      })

      it(`allows ROLE_AGENT to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_AGENT')).toBe(true)
      })

      it(`denies ROLE_EMPLOYEE to access ${route}`, () => {
        expect(isRouteAllowedForRole(route, 'ROLE_EMPLOYEE')).toBe(false)
      })
    })
  })

  describe('Nested Sub-route matching', () => {
    it('denies nested restricted route for employee', () => {
      expect(isRouteAllowedForRole('/audit/logs/123', 'ROLE_EMPLOYEE')).toBe(false)
      expect(isRouteAllowedForRole('/monitoring/metrics', 'ROLE_EMPLOYEE')).toBe(false)
      expect(isRouteAllowedForRole('/assets/hardware/new', 'ROLE_EMPLOYEE')).toBe(false)
    })

    it('allows nested restricted route for admin', () => {
      expect(isRouteAllowedForRole('/audit/logs/123', 'ROLE_ADMIN')).toBe(true)
      expect(isRouteAllowedForRole('/monitoring/metrics', 'ROLE_ADMIN')).toBe(true)
      expect(isRouteAllowedForRole('/assets/hardware/new', 'ROLE_ADMIN')).toBe(true)
    })
  })
})

describe('post-login redirect safety', () => {
  it('keeps local return URLs', () => {
    expect(safeLocalRedirect('/helpdesk?page=2')).toBe('/helpdesk?page=2')
  })

  it('rejects absolute, protocol-relative, and backslash redirects', () => {
    expect(safeLocalRedirect('https://evil.example')).toBe('/')
    expect(safeLocalRedirect('//evil.example')).toBe('/')
    expect(safeLocalRedirect('/\\evil.example')).toBe('/')
  })
})
