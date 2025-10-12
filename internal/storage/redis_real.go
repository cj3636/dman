package storage

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.tyss.io/cj3636/dman/internal/config"
	redis "github.com/redis/go-redis/v9"
)

// redisBackend implements a real Redis-backed storage. Files are stored as chunked binary values:
//
//	base key holds JSON manifest {"chunks":N,"v":1}; each chunk stored at base:chunk:i (256KB default).
//
// Supports legacy single-value objects created before chunking. Includes simple exponential backoff retries.
// NOTE: Future enhancements: configurable chunk size, metadata hash side-car, pipeline/multi optimizations.
// For large deployments, consider streaming and chunking; current approach buffers entire file in memory.
type redisBackend struct {
	client *redis.Client
}

func NewRedisBackend(cfg *config.Config) (Backend, error) {
	opts := &redis.Options{}
	if cfg.Redis.Socket != "" {
		// unix socket
		opts.Network = "unix"
		opts.Addr = cfg.Redis.Socket
	} else {
		if cfg.Redis.Addr == "" {
			cfg.Redis.Addr = "127.0.0.1:6379"
		}
		opts.Addr = cfg.Redis.Addr
	}
	if cfg.Redis.Username != "" {
		opts.Username = cfg.Redis.Username
	}
	if cfg.Redis.Password != "" {
		opts.Password = cfg.Redis.Password
	}
	opts.DB = cfg.Redis.DB
	if cfg.Redis.TLS {
		to := &tls.Config{InsecureSkipVerify: cfg.Redis.TLSInsecureSkip}
		if cfg.Redis.TLSServerName != "" {
			to.ServerName = cfg.Redis.TLSServerName
		}
		if cfg.Redis.TLSCA != "" {
			pem, err := os.ReadFile(cfg.Redis.TLSCA)
			if err != nil {
				return nil, err
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(pem) {
				return nil, errors.New("failed to append redis ca")
			}
			to.RootCAs = pool
		}
		opts.TLSConfig = to
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &redisBackend{client: client}, nil
}

func (r *redisBackend) sanitize(user, rel string) (string, error) {
	if user == "" {
		return "", errors.New("empty user")
	}
	rel = filepath.ToSlash(strings.TrimPrefix(rel, "./"))
	if rel == "" {
		return "", errors.New("empty path")
	}
	if len(rel) > MaxPathLen {
		return "", errors.New("path too long")
	}
	if strings.HasPrefix(rel, "/") {
		return "", errors.New("absolute path disallowed")
	}
	if strings.Contains(rel, "..") {
		return "", errors.New("path traversal disallowed")
	}
	return user + "/" + rel, nil
}

// Chunked redis backend constants
const (
	redisChunkSize  = 256 * 1024 // 256KB per chunk
	redisMaxRetries = 3
)

type redisManifest struct {
	Chunks int `json:"chunks"`
	V      int `json:"v"`
}

func (r *redisBackend) retry(ctx context.Context, op string, fn func() error) error {
	var err error
	backoff := 50 * time.Millisecond
	for attempt := 0; attempt < redisMaxRetries; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return fmt.Errorf("redis %s failed after retries: %w", op, err)
}

func (r *redisBackend) chunkKey(base string, idx int) string {
	return fmt.Sprintf("%s:chunk:%d", base, idx)
}

// Save now streams into chunk keys and writes a manifest at base key.
func (r *redisBackend) Save(user, rel string, rd io.Reader) error {
	base, err := r.sanitize(user, rel)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// If existing manifest, delete old chunks first (best effort)
	oldManifestRaw, _ := r.client.Get(ctx, base).Bytes()
	if len(oldManifestRaw) > 0 && strings.HasPrefix(string(oldManifestRaw), "{") {
		var om redisManifest
		if json.Unmarshal(oldManifestRaw, &om) == nil && om.Chunks > 0 {
			for i := 0; i < om.Chunks; i++ {
				_ = r.client.Del(ctx, r.chunkKey(base, i)).Err()
			}
		}
	}
	buf := make([]byte, redisChunkSize)
	idx := 0
	for {
		n, readErr := io.ReadFull(rd, buf)
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			if n > 0 {
				chunkCopy := make([]byte, n)
				copy(chunkCopy, buf[:n])
				key := r.chunkKey(base, idx)
				if err := r.retry(ctx, "set-chunk", func() error { return r.client.Set(ctx, key, chunkCopy, 0).Err() }); err != nil {
					return err
				}
				idx++
			}
			break
		}
		if readErr != nil {
			return readErr
		}
		chunkCopy := make([]byte, n)
		copy(chunkCopy, buf[:n])
		key := r.chunkKey(base, idx)
		if err := r.retry(ctx, "set-chunk", func() error { return r.client.Set(ctx, key, chunkCopy, 0).Err() }); err != nil {
			return err
		}
		idx++
	}
	manifestBytes, _ := json.Marshal(redisManifest{Chunks: idx, V: 1})
	if err := r.retry(ctx, "set-manifest", func() error { return r.client.Set(ctx, base, manifestBytes, 0).Err() }); err != nil {
		return err
	}
	return nil
}

// Open reconstructs file from chunks (supports legacy single-value storage).
func (r *redisBackend) Open(user, rel string) (*os.File, error) {
	base, err := r.sanitize(user, rel)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	raw, err := r.client.Get(ctx, base).Bytes()
	if err != nil {
		return nil, err
	}
	var mf redisManifest
	chunkMode := false
	if len(raw) > 0 && raw[0] == '{' && json.Unmarshal(raw, &mf) == nil && mf.Chunks >= 0 {
		chunkMode = true
	}
	f, err := os.CreateTemp("", ".dman-redis-*.")
	if err != nil {
		return nil, err
	}
	if !chunkMode { // legacy single value
		if _, err := f.Write(raw); err != nil {
			f.Close()
			return nil, err
		}
	} else {
		for i := 0; i < mf.Chunks; i++ {
			ck := r.chunkKey(base, i)
			b, err := r.client.Get(ctx, ck).Bytes()
			if err != nil {
				f.Close()
				return nil, err
			}
			if _, err := f.Write(b); err != nil {
				f.Close()
				return nil, err
			}
		}
	}
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// List returns only base manifest/value keys (filters out chunk suffixes).
func (r *redisBackend) List() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	var cursor uint64
	var out []string
	for {
		keys, cur, err := r.client.Scan(ctx, cursor, "*", 512).Result()
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			if strings.Contains(k, ":chunk:") {
				continue
			}
			out = append(out, k)
		}
		cursor = cur
		if cursor == 0 {
			break
		}
	}
	return out, nil
}

// Delete removes manifest/value and any chunks.
func (r *redisBackend) Delete(user, rel string) error {
	base, err := r.sanitize(user, rel)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	raw, _ := r.client.Get(ctx, base).Bytes()
	if len(raw) > 0 && raw[0] == '{' {
		var mf redisManifest
		if json.Unmarshal(raw, &mf) == nil && mf.Chunks > 0 {
			for i := 0; i < mf.Chunks; i++ {
				_ = r.client.Del(ctx, r.chunkKey(base, i)).Err()
			}
		}
	}
	return r.client.Del(ctx, base).Err()
}
