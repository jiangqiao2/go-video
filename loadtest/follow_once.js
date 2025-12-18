import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// Gateway base URL. Defaults to你的云上 NodePort 地址。
// 可通过环境变量 GATEWAY 覆盖，例如：
//   GATEWAY=http://117.50.33.177:30080 k6 run follow_once.js
const GATEWAY = __ENV.GATEWAY || 'http://117.50.33.177:30080';

// 被所有粉丝关注的“明星用户”UUID
const TARGET_UUID = __ENV.TARGET_UUID;

if (!TARGET_UUID) {
  throw new Error('TARGET_UUID env var is required (star user UUID)');
}

// 从 tokens.txt 加载粉丝 access_token（每行一个）
const TOKENS = new SharedArray('tokens', function () {
  const text = open('./tokens.txt');
  return text
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0);
});

if (!TOKENS.length) {
  throw new Error('No tokens loaded from tokens.txt. Run prepare_fans.ps1 or run_follow_loadtest.sh first.');
}

// 压测场景：每个 token（粉丝账号）只执行一次 follow/toggle，不重复关注。
// 比如 tokens.txt 有 10000 行，则总共会发 10000 次请求，用于“不丢事件”验收。
export const options = {
  scenarios: {
    follow_once: {
      executor: 'shared-iterations',
      iterations: TOKENS.length, // 每个 token 一次
      vus: Math.min(TOKENS.length, Number(__ENV.VUS || 300)),
      // 防止某些请求超时拖太久，可通过 MAX_DURATION 覆盖。
      maxDuration: __ENV.MAX_DURATION || '5m',
    },
  },
};

export default function () {
  const idx = __ITER; // 0 .. TOKENS.length-1
  if (idx >= TOKENS.length) {
    return;
  }

  const token = TOKENS[idx];
  const url = `${GATEWAY}/api/user/v1/inner/relation/follow/toggle`;
  const payload = JSON.stringify({ target_uuid: TARGET_UUID, action: 'follow' });

  const res = http.post(url, payload, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    timeout: '5s',
  });

  check(res, {
    'status is expected': (r) =>
      r.status === 200 ||
      r.status === 400 ||
      r.status === 401 ||
      r.status === 409 ||
      r.status === 429,
  });

  // 略微 sleep，避免单个 VU 自旋过快；总请求数由 iterations 控制。
  sleep(0.01);
}

