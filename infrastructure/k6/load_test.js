import http from 'k6/http';
import { check, sleep } from 'k6';

// K6 Load Test Configuration: Simulating up to 500 Virtual Users (VUs)
export const options = {
  stages: [
    { duration: '5s', target: 50 },   // Warm-up ramp-up
    { duration: '15s', target: 200 }, // Normal production load
    { duration: '15s', target: 500 }, // Peak traffic stress load
    { duration: '5s', target: 0 },    // Cool-down
  ],
  thresholds: {
    // 95% of requests must complete in under 200ms
    http_req_duration: ['p(95)<200'],
    // Less than 1% error rate allowed under load
    http_req_failed: ['rate<0.01'],
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

  // 1. Gateway Health & Prometheus Metrics (Phase 8 RED Monitoring)
  const resHealth = http.get(`${BASE_URL}/health`);
  check(resHealth, {
    'health check status 200': (r) => r.status === 200,
  });

  const resMetrics = http.get(`${BASE_URL}/metrics`);
  check(resMetrics, {
    'prometheus metrics status 200': (r) => r.status === 200,
  });

  // 2. Monitoring Overview (SRE Console)
  const resOverview = http.get(`${BASE_URL}/api/v1/monitoring/overview`);
  check(resOverview, {
    'monitoring overview status 200': (r) => r.status === 200,
  });

  // 3. Reporting & BI Overview (Phase 9 Executive KPIs)
  const resReports = http.get(`${BASE_URL}/api/v1/reports/overview?range=30d`, params);
  check(resReports, {
    'reporting overview status 200': (r) => r.status === 200,
  });

  // 4. Audit Trail Stream (Phase 10 Security & Compliance)
  const resAudit = http.get(`${BASE_URL}/api/v1/audit/logs?page=1&page_size=10`, params);
  check(resAudit, {
    'audit logs stream status 200': (r) => r.status === 200,
  });

  sleep(0.1);
}
