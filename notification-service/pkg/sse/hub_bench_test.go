package sse

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
)

// benchHub is the minimal interface implemented by both Hub and SyncMapHub.
type benchHub interface {
	Subscribe(userUUID string) (<-chan Event, func())
	Publish(userUUID string, ev Event)
}

// parseSubscribersEnv reads HUB_BENCH_SUBSCRIBERS environment variable to
// allow overriding the number of pre-created subscribers in heavy benchmarks.
func parseSubscribersEnv(defaultValue int) int {
	if v := os.Getenv("HUB_BENCH_SUBSCRIBERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultValue
}

// parseUsersEnv reads HUB_BENCH_USERS environment variable to allow
// overriding the number of distinct users in many-user benchmarks.
func parseUsersEnv(defaultValue int) int {
	if v := os.Getenv("HUB_BENCH_USERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultValue
}

// --- Subscribe/Unsubscribe churn benchmarks ---

func benchmarkSubUnsub(b *testing.B, newHub func() benchHub) {
	h := newHub()
	const user = "user-1"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, unsubscribe := h.Subscribe(user)
			unsubscribe()
		}
	})
}

func BenchmarkMutexHub_SubUnsub(b *testing.B) {
	benchmarkSubUnsub(b, func() benchHub { return NewHub() })
}

func BenchmarkSyncMapHub_SubUnsub(b *testing.B) {
	benchmarkSubUnsub(b, func() benchHub { return NewSyncMapHub() })
}

// --- Publish with many subscribers (single user, steady-state) ---

func benchmarkPublishSteadyState(b *testing.B, newHub func() benchHub, defaultSubscribers int) {
	h := newHub()
	const user = "user-steady"

	subs := parseSubscribersEnv(defaultSubscribers)
	for i := 0; i < subs; i++ {
		// We ignore the channel and unsubscribe here; this benchmark focuses
		// on Publish cost with a fixed subscriber set.
		_, _ = h.Subscribe(user)
	}

	ev := Event{Type: "bench"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Publish(user, ev)
		}
	})
}

// 默认预创建 1e5 个订阅；实际要压到“单用户百万连接”时可以：
// HUB_BENCH_SUBSCRIBERS=1000000 go test ./pkg/sse -run=^$ -bench=BenchmarkMutexHub_PublishSteady -benchmem

func BenchmarkMutexHub_PublishSteady(b *testing.B) {
	benchmarkPublishSteadyState(b, func() benchHub { return NewHub() }, 100_000)
}

func BenchmarkSyncMapHub_PublishSteady(b *testing.B) {
	benchmarkPublishSteadyState(b, func() benchHub { return NewSyncMapHub() }, 100_000)
}

// --- Churn: subscribe + publish + unsubscribe in a tight loop ---

func benchmarkChurn(b *testing.B, newHub func() benchHub) {
	h := newHub()
	const user = "user-churn"
	ev := Event{Type: "bench"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, unsubscribe := h.Subscribe(user)
			h.Publish(user, ev)
			unsubscribe()
		}
	})
}

func BenchmarkMutexHub_Churn(b *testing.B) {
	benchmarkChurn(b, func() benchHub { return NewHub() })
}

func BenchmarkSyncMapHub_Churn(b *testing.B) {
	benchmarkChurn(b, func() benchHub { return NewSyncMapHub() })
}

// --- Many users, single connection per user ---

// benchmarkManyUsersSingleConnPublish 模拟“10 万用户各自 1 条连接”的场景：
// 预先为 N 个 userUUID 建立订阅，每次 Publish 只对其中一个用户推送事件。
func benchmarkManyUsersSingleConnPublish(b *testing.B, newHub func() benchHub, defaultUsers int) {
	h := newHub()

	users := parseUsersEnv(defaultUsers)
	userIDs := make([]string, users)
	for i := 0; i < users; i++ {
		uid := fmt.Sprintf("user-%d", i)
		userIDs[i] = uid
		_, _ = h.Subscribe(uid)
	}

	ev := Event{Type: "bench"}

	var idx uint64

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 简单轮询不同用户，避免 rand 带来的额外开销。
			i := atomic.AddUint64(&idx, 1)
			uid := userIDs[int(i)%users]
			h.Publish(uid, ev)
		}
	})
}

// 默认 1e5 用户，每个用户 1 条连接。
// HUB_BENCH_USERS=100000 go test ./pkg/sse -run=^$ -bench=BenchmarkMutexHub_ManyUsersSingleConn -benchmem

func BenchmarkMutexHub_ManyUsersSingleConn(b *testing.B) {
	benchmarkManyUsersSingleConnPublish(b, func() benchHub { return NewHub() }, 100_000)
}

func BenchmarkSyncMapHub_ManyUsersSingleConn(b *testing.B) {
	benchmarkManyUsersSingleConnPublish(b, func() benchHub { return NewSyncMapHub() }, 100_000)
}
