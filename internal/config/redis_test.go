package config

import "testing"

func TestRedisValidateDefaults(t *testing.T) {
	c := &Config{ServerURL: "http://x", StorageDriver: "redis", Users: map[string]User{"u": {Home: "/h/"}}}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if c.Redis.Addr == "" {
		t.Fatalf("expected default addr set")
	}
	if c.Redis.Socket != "" {
		t.Fatalf("socket should be empty")
	}
}

func TestRedisValidateMutualExclusive(t *testing.T) {
	c := &Config{ServerURL: "http://x", StorageDriver: "redis", Users: map[string]User{"u": {Home: "/h/"}}, Redis: Redis{Addr: "127.0.0.1:6379", Socket: "/tmp/redis.sock"}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected error for both addr and socket")
	}
}
