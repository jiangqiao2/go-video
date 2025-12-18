import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

// Gateway base URL. 默认指向云上的 Kong NodePort；可通过 GATEWAY 覆盖，例如：
//   GATEWAY=http://117.50.33.177:30080 k6 run follow_once.js
const GATEWAY = __ENV.GATEWAY || 'http://117.50.33.177:30080';

// 被所有粉丝关注的“明星用户” UUID
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

const FAN_COUNT = TOKENS.length;

// 压测场景：每个 token（粉丝账号）只执行一次 follow/toggle，不重复关注。
// 这里使用 per-vu-iterations，每个 VU 对应 tokens.txt 中的一行，保证“一人一次”。
export const options = {
  scenarios: {
    follow_once: {
      executor: 'per-vu-iterations',
      vus: FAN_COUNT,                      // 每个 VU 绑定一个 token
      iterations: 1,                       // 每个 VU 只跑一次
      maxDuration: __ENV.MAX_DURATION || '5m',
    },
  },
};

export default function () {
  const idx = __VU - 1; // __VU 从 1 开始，这里映射到 0..FAN_COUNT-1
  if (idx < 0 || idx >= FAN_COUNT) {
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

  // 略微 sleep，避免 VU 自旋过快；总请求数由 vus * iterations 决定
  sleep(0.01);
}

