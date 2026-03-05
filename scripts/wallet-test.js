import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    constant_rate: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '10s',
      preAllocatedVUs: 10,
      maxVUs: 100,
    },
  },
};

const BASE_URL = 'http://api:8080';
const WALLET_ID = '11111111-1111-1111-1111-111111111110';

export default function () {
  const payload = JSON.stringify({
    walletId: WALLET_ID,
    operationType: 'DEPOSIT',
    amount: 100,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const res = http.post(`${BASE_URL}/api/v1/wallet`, payload, params);

  if (res.status !== 200)
    console.log('status:', res.status, 'body:', res.body);

  check(res, {
    'no 5xx': (r) => r.status < 500,
    'status 200': (r) => r.status === 200,
  });
}
