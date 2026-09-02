import http from 'k6/http';
import { check, fail, sleep, group } from 'k6';

// k6 Load & Performance Testing Script for EOMP
// Tests core operational endpoints under concurrent load against strict SLA thresholds.
export const options = {
  stages: [
    { duration: '30s', target: 20 },  // Ramp-up to 20 virtual users
    { duration: '1m', target: 50 },   // Steady-state load at 50 virtual users
    { duration: '30s', target: 100 },  // Peak spike to 100 virtual users
    { duration: '30s', target: 0 },    // Ramp-down
  ],
  thresholds: {
    // Every functional check must pass; a fast error response is still a failure.
    checks: ['rate==1'],
    // 95% of requests must complete below 200ms, 99% below 500ms (P0 SLA requirement)
    http_req_duration: ['p(95)<200', 'p(99)<500'],
    // Less than 1% of requests may fail
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = (__ENV.BASE_URL || '').replace(/\/$/, '');

function login(email, password, label) {
  const response = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    {
      headers: { 'Content-Type': 'application/json', 'X-Forwarded-For': '198.51.100.1' },
      tags: { endpoint: `login_${label}` },
    },
  );

  if (!check(response, { [`${label} login status is 200`]: (r) => r.status === 200 })) {
    fail(`${label} login failed with HTTP ${response.status}`);
  }

  const token = response.json('access_token');
  if (!token) {
    fail(`${label} login response did not contain access_token`);
  }
  return token;
}

export function setup() {
  const employeeEmail = __ENV.LOADTEST_EMPLOYEE_EMAIL;
  const employeePassword = __ENV.LOADTEST_EMPLOYEE_PASSWORD;
  const managerEmail = __ENV.LOADTEST_MANAGER_EMAIL;
  const managerPassword = __ENV.LOADTEST_MANAGER_PASSWORD;

  if (!BASE_URL || !employeeEmail || !employeePassword || !managerEmail || !managerPassword) {
    fail('BASE_URL and LOADTEST_EMPLOYEE_*/LOADTEST_MANAGER_* credentials are required');
  }

  return {
    employeeToken: login(employeeEmail, employeePassword, 'employee'),
    managerToken: login(managerEmail, managerPassword, 'manager'),
  };
}

export default function (data) {
  // The Gate D runner has one trusted proxy address and presents a stable,
  // unique client address per VU. This exercises the production per-client
  // limiter instead of accidentally load-testing a single proxy bucket.
  const clientAddress = `198.51.${Math.floor(__VU / 254)}.${(__VU % 254) + 1}`;
  const employeeHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.employeeToken}`,
    'X-Forwarded-For': clientAddress,
  };
  const managerHeaders = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${data.managerToken}`,
    'X-Forwarded-For': clientAddress,
  };

  group('1. Health & Readiness Probes', function () {
    const res = http.get(`${BASE_URL}/health`, {
      headers: { 'X-Forwarded-For': clientAddress },
      tags: { endpoint: 'health' },
    });
    check(res, {
      'health status is 200': (r) => r.status === 200,
      'health response time < 50ms': (r) => r.timings.duration < 50,
    });
  });

  sleep(0.5);

  group('2. Helpdesk Ticket Listing & Search', function () {
    const res = http.get(`${BASE_URL}/api/v1/tickets?page=1&page_size=20&status=OPEN`, {
      headers: employeeHeaders,
      tags: { endpoint: 'tickets' },
    });
    check(res, {
      'ticket list status is 200': (r) => r.status === 200,
    });
  });

  sleep(0.5);

  group('3. Reporting Overview', function () {
    const res = http.get(`${BASE_URL}/api/v1/reports/overview?range=30d`, {
      headers: managerHeaders,
      tags: { endpoint: 'reports_overview' },
    });
    check(res, {
      'report overview status is 200': (r) => r.status === 200,
    });
  });

  sleep(1);
}
