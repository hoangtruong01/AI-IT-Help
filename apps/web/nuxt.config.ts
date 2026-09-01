// https://nuxt.com/docs/api/configuration/nuxt-config
const webSecurityHeaders = {
  'Content-Security-Policy': 'default-src \'self\'; script-src \'self\' \'unsafe-inline\'; style-src \'self\' \'unsafe-inline\' https://fonts.googleapis.com; img-src \'self\' data: blob: https://images.unsplash.com; font-src \'self\' data: https://fonts.gstatic.com; connect-src \'self\'; frame-ancestors \'none\'; base-uri \'self\'; form-action \'self\'; object-src \'none\'; upgrade-insecure-requests',
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'X-Content-Type-Options': 'nosniff',
  'X-Frame-Options': 'DENY',
  'Permissions-Policy': 'camera=(), microphone=(), geolocation=()'
}

export default defineNuxtConfig({
  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@pinia/nuxt',
    '@vueuse/nuxt'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  // Runtime config - environment variables
  runtimeConfig: {
    // Server-only (not exposed to client). In containers this must point to the
    // internal Gateway service, not the browser-facing relative /api path.
    apiBaseUrl: 'http://localhost:8080',
    allowedOrigins: '',

    // Public (exposed to client)
    public: {
      apiUrl: 'http://localhost:8080'
    }
  },

  routeRules: {
    '/': { prerender: true, headers: webSecurityHeaders },
    '/**': { headers: webSecurityHeaders }
  },

  compatibilityDate: '2026-06-30',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  },

  icon: {
    provider: 'none',
    serverBundle: false,
    clientBundle: {
      scan: true
    }
  }
})
