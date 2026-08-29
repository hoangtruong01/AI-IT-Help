import http from 'k6/http';
import { check, sleep } from 'k6';

// =============================================================================
// EOMP PHASE 6: PEAK STRESS & SPIKE BENCHMARK (800 VIRTUAL USERS)
// =============================================================================

export const options = {
  stages: [
    { duration: '3s', target: 100 },  // Quick ramp
    { duration: '10s', target: 800 }, // Mega Spike to 800 VUs
    { duration: '5s', target: 200 },  // Sustained high recovery
    { duration: '2s', target: 0 },    // Cool down
  ],
  thresholds: {
    // 99% of requests under extreme spike must complete in < 500ms
    http_req_duration: ['p(99)<500'],
    // Error rate must stay below 2% during 800 VUs spike
    http_req_failed: ['rate<0.02'],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || 'http://localhost:8080';

export default function () {
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-User-Role': 'ROLE_ADMIN',
    },
  };

  const resHealth = http.get(`${BASE_URL}/health`);
  check(resHealth, {
    'health check status 200': (r) => r.status === 200,
  });

  const resMetrics = http.get(`${BASE_URL}/metrics`);
  check(resMetrics, {
    'metrics scrape status 200': (r) => r.status === 200,
  });

  const resMonitoring = http.get(`${BASE_URL}/api/v1/monitoring/overview`);
  check(resMonitoring, {
    'monitoring overview status 200': (r) => r.status === 200,
  });

  sleep(0.02);
}
