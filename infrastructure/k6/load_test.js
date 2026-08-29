import http from 'k6/http';
import { check, group, sleep } from 'k6';

// =============================================================================
// EOMP PHASE 6: ENTERPRISE CONCURRENCY & LOAD BENCHMARK (500 VIRTUAL USERS)
// =============================================================================

export const options = {
  stages: [
    { duration: '5s', target: 50 },   // Warm-up & connection pool ramp-up
    { duration: '15s', target: 200 }, // Sustained production business traffic
    { duration: '15s', target: 500 }, // Peak concurrent load spike
    { duration: '5s', target: 0 },    // Graceful cool-down
  ],
  thresholds: {
    // 95% of all HTTP requests must complete in under 200ms
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    // Less than 1% error rate across cluster
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || 'http://localhost:8080';

export default function () {
  const authHeaders = {
    'Content-Type': 'application/json',
    'X-User-ID': `u-agent-${__VU}`,
    'X-User-Email': `agent.${__VU}@eomp.local`,
    'X-User-Role': 'ROLE_AGENT',
    'X-User-Name': `Support Agent ${__VU}`,
  };

  // Group 1: Gateway Health & SRE Metrics Radar
  group('01_Gateway_Observability', function () {
    const resHealth = http.get(`${BASE_URL}/health`);
    check(resHealth, {
      'health check status 200': (r) => r.status === 200,
    });

    const resMetrics = http.get(`${BASE_URL}/metrics`);
    check(resMetrics, {
      'prometheus RED metrics status 200': (r) => r.status === 200,
    });

    const resOverview = http.get(`${BASE_URL}/api/v1/monitoring/overview`);
    check(resOverview, {
      'monitoring overview status 200': (r) => r.status === 200,
    });
  });

  // Group 2: Helpdesk & Concurrency Ticket Processing
  group('02_Helpdesk_Operations', function () {
    const ticketPayload = JSON.stringify({
      title: `Network Outage Zone ${__VU % 10} - High Packet Loss`,
      description: 'Multiple users reporting intermittent VPN disconnects on Gateway cluster node.',
      category: 'NETWORK',
      priority: 'HIGH',
      requester_id: `u-emp-${__VU}`,
    });

    const resCreate = http.post(`${BASE_URL}/api/v1/tickets`, ticketPayload, { headers: authHeaders });
    check(resCreate, {
      'ticket creation status 200 or 201': (r) => r.status === 200 || r.status === 201,
    });

    const resList = http.get(`${BASE_URL}/api/v1/tickets?page=1&page_size=10`, { headers: authHeaders });
    check(resList, {
      'tickets list status 200': (r) => r.status === 200,
    });
  });

  // Group 3: AI Operations Copilot & RAG Query
  group('03_AI_Copilot_Triage', function () {
    const aiPayload = JSON.stringify({
      title: 'VPN Connection Failure after certificate renewal',
      description: 'User cannot connect to corporate VPN staging gateway with Error code SEC-8801.',
    });

    const resAI = http.post(`${BASE_URL}/api/v1/ai/analyze-ticket`, aiPayload, { headers: authHeaders });
    check(resAI, {
      'ai triage analysis status 200': (r) => r.status === 200,
    });
  });

  // Group 4: CMDB & Asset Inventory
  group('04_Asset_CMDB_Lookup', function () {
    const resAssets = http.get(`${BASE_URL}/api/v1/assets?page=1&page_size=10`, { headers: authHeaders });
    check(resAssets, {
      'assets query status 200': (r) => r.status === 200,
    });
  });

  // Group 5: BI Analytics & Executive Reporting Under Load
  group('05_Reporting_And_Audit', function () {
    const resReports = http.get(`${BASE_URL}/api/v1/reports/overview?range=30d`, { headers: authHeaders });
    check(resReports, {
      'reporting overview status 200': (r) => r.status === 200,
    });

    const adminHeaders = Object.assign({}, authHeaders, { 'X-User-Role': 'ROLE_ADMIN' });
    const resAudit = http.get(`${BASE_URL}/api/v1/audit/logs?page=1&page_size=10`, { headers: adminHeaders });
    check(resAudit, {
      'audit trail query status 200': (r) => r.status === 200,
    });
  });

  sleep(0.05); // Rapid iteration pacing
}
