package storage

import (
	"bytes"
	"fmt"
	"git.tyss.io/cj3636/dman/internal/config"
	"os"
	"testing"
)

func benchmarkBackend(b *testing.B, cfg *config.Config) {
	backend, err := NewBackend(cfg, b.TempDir())
	if err != nil {
		b.Fatalf("backend init: %v", err)
	}
	payload := bytes.Repeat([]byte("x"), 1024) // 1KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := fmt.Sprintf("user%d", i%5)
		if err := backend.Save(name, fmt.Sprintf("f%d.txt", i), bytes.NewReader(payload)); err != nil {
			b.Fatalf("save: %v", err)
		}
	}
}

func BenchmarkSaveDisk(b *testing.B) {
	cfg := &config.Config{ServerURL: "http://x", StorageDriver: "disk", Users: map[string]config.User{"u": {Home: "/h/"}}}
	benchmarkBackend(b, cfg)
}

func BenchmarkSaveRedisMem(b *testing.B) {
	cfg := &config.Config{ServerURL: "http://x", StorageDriver: "redis-mem", Users: map[string]config.User{"u": {Home: "/h/"}}}
	benchmarkBackend(b, cfg)
}

func BenchmarkSaveRedisRealOptional(b *testing.B) {
	addr := os.Getenv("DMAN_BENCH_REDIS_ADDR")
	if addr == "" {
		b.Skip("DMAN_BENCH_REDIS_ADDR not set")
	}
	cfg := &config.Config{ServerURL: "http://x", StorageDriver: "redis", Users: map[string]config.User{"u": {Home: "/h/"}}, Redis: config.Redis{Addr: addr}}
	if err := cfg.Validate(); err != nil {
		b.Fatalf("validate: %v", err)
	}
	benchmarkBackend(b, cfg)
}
