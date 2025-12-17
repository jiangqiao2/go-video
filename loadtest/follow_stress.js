import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// Gateway base URL. Defaults to your provided address with NodePort 30080.
// You can override it with env var GATEWAY if needed.
const GATEWAY = __ENV.GATEWAY || 'http://117.50.33.177:30080';

// UUID of the "star" user that all fans will follow
const TARGET_UUID = __ENV.TARGET_UUID;

if (!TARGET_UUID) {
  throw new Error('TARGET_UUID env var is required (star user UUID)');
}

// Load fan access tokens from tokens.txt in the same folder (one per line)
const TOKENS = new SharedArray('tokens', function () {
  const text = open('./tokens.txt');
  return text
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
});

if (!TOKENS.length) {
  throw new Error('No tokens loaded from tokens.txt. Run prepare_fans.ps1 first.');
}

export const options = {
  scenarios: {
    follow_storm: {
      executor: 'constant-arrival-rate',

      // Target QPS (follow requests per second).
      // 当前配置：1500 QPS，适合作为下一档压测。
      rate: 1500,
      timeUnit: '1s',

      // Total duration of the test.
      duration: '5m',

      // Client-side concurrency capacity.
      preAllocatedVUs: 200,
      maxVUs: 1000,
    },
  },
};

export default function () {
  const idx = __ITER % TOKENS.length;
  const token = TOKENS[idx];

  const url = `${GATEWAY}/api/user/v1/inner/relation/follow`;
  const payload = JSON.stringify({ target_uuid: TARGET_UUID });

  const res = http.post(url, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    timeout: '5s',
  });

  // Consider 200 as success; also tolerate some business errors such as
  // duplicate follow or validation issues for the purpose of stress testing.
  check(res, {
    'status is expected': (r) =>
      r.status === 200 ||
      r.status === 400 ||
      r.status === 401 ||
      r.status === 409 ||
      r.status === 429,
  });

  // Small sleep to avoid a single VU spinning too aggressively;
  // overall RPS is controlled by the "rate" above.
  sleep(0.01);
}
