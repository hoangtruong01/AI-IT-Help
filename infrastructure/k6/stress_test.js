import http from 'k6/http';
import { check, sleep } from 'k6';

// Stress & Spike Test: Rapid spike up to 800 VUs
export const options = {
  stages: [
    { duration: '3s', target: 100 },
    { duration: '10s', target: 800 }, // Spike
    { duration: '5s', target: 100 },
    { duration: '2s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],
  },
};

const BASE_URL = __ENV.GATEWAY_URL || 'http://localhost:8080';

export default function () {
  const res = http.get(`${BASE_URL}/health`);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(0.05);
}
