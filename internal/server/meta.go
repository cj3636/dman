package server

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Meta holds mutable operational metadata persisted alongside store data.
type Meta struct {
	LastPublish string            `json:"last_publish"`
	LastInstall string            `json:"last_install"`
	Metrics     map[string]uint64 `json:"metrics"`
	mu          sync.RWMutex      `json:"-"`
	path        string            `json:"-"`
}

func loadMeta(root string) (*Meta, error) {
	p := filepath.Join(root, "_meta.json")
	m := &Meta{Metrics: map[string]uint64{}, path: p}
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return m, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, err
	}
	m.path = p
	if m.Metrics == nil {
		m.Metrics = map[string]uint64{}
	}
	return m, nil
}

func (m *Meta) save() error {
	m.mu.RLock()
	path := m.path
	m.mu.RUnlock()
	if path == "" {
		return errors.New("meta path unset")
	}
	m.mu.RLock()
	b, err := json.MarshalIndent(struct {
		LastPublish string            `json:"last_publish"`
		LastInstall string            `json:"last_install"`
		Metrics     map[string]uint64 `json:"metrics"`
	}{m.LastPublish, m.LastInstall, m.Metrics}, "", "  ")
	m.mu.RUnlock()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".meta-*")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func (m *Meta) recordPublish() {
	m.mu.Lock()
	m.LastPublish = time.Now().UTC().Format(time.RFC3339)
	m.Metrics["publish_requests"]++
	m.mu.Unlock()
	_ = m.save()
}
func (m *Meta) recordInstall() {
	m.mu.Lock()
	m.LastInstall = time.Now().UTC().Format(time.RFC3339)
	m.Metrics["install_requests"]++
	m.mu.Unlock()
	_ = m.save()
}
func (m *Meta) snapshot() (lastPub, lastInst string, metrics map[string]uint64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := map[string]uint64{}
	for k, v := range m.Metrics {
		cp[k] = v
	}
	return m.LastPublish, m.LastInstall, cp
}
