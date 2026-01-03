package config

import (
	"errors"
	"fmt"
	"git.tyss.io/cj3636/dman/pkg/model"
	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type User struct {
	Home        string   `yaml:"home" json:"home"`
	Track       []string `yaml:"track" json:"track"`
	LegacyTrack []string `yaml:"include,omitempty" json:"-"`
}

// Redis configuration (optional when storage_driver != redis)
type Redis struct {
	Addr            string `yaml:"addr" json:"addr"` // host:port for TCP
	Socket          string `yaml:"socket" json:"socket"`
	Username        string `yaml:"username" json:"username"`
	Password        string `yaml:"password" json:"password"`
	DB              int    `yaml:"db" json:"db"`
	TLS             bool   `yaml:"tls" json:"tls"`
	TLSCA           string `yaml:"tls_ca" json:"tls_ca"`
	TLSInsecureSkip bool   `yaml:"tls_insecure_skip_verify" json:"tls_insecure_skip_verify"`
	TLSServerName   string `yaml:"tls_server_name" json:"tls_server_name"`
}

// MariaDB configuration (optional when storage_driver not in maria/mariadb/mysql)
type Maria struct {
	Addr            string `yaml:"addr" json:"addr"` // host:port
	Socket          string `yaml:"socket" json:"socket"`
	DB              string `yaml:"db" json:"db"`
	User            string `yaml:"user" json:"user"`
	Password        string `yaml:"password" json:"password"`
	TLS             bool   `yaml:"tls" json:"tls"`
	TLSCA           string `yaml:"tls_ca" json:"tls_ca"`
	TLSCert         string `yaml:"tls_cert" json:"tls_cert"`
	TLSKey          string `yaml:"tls_key" json:"tls_key"`
	TLSInsecureSkip bool   `yaml:"tls_insecure_skip_verify" json:"tls_insecure_skip_verify"`
	TLSServerName   string `yaml:"tls_server_name" json:"tls_server_name"`
}

type Config struct {
	AuthToken     string          `yaml:"auth_token" json:"auth_token"`
	ServerURL     string          `yaml:"server_url" json:"server_url"`
	StorageDriver string          `yaml:"storage_driver" json:"storage_driver"`
	GlobalTrack   []string        `yaml:"track" json:"track"`
	LegacyTrack   []string        `yaml:"include,omitempty" json:"-"`
	Users         map[string]User `yaml:"users" json:"users"`
	Redis         Redis           `yaml:"redis" json:"redis"`
	Maria         Maria           `yaml:"db" json:"db"`
	path          string          // loaded from
}

func (c *Config) UserNames() []string {
	var names []string
	for k := range c.Users {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (c *Config) UsersList() []model.UserSpec {
	var out []model.UserSpec
	for name, u := range c.Users {
		inc := effectiveTrack(u.Track, c.GlobalTrack)
		out = append(out, model.UserSpec{Name: name, Home: u.Home, Track: inc})
	}
	return out
}

func (c *Config) expand() {
	home, _ := os.UserHomeDir()
	for k, u := range c.Users {
		h := u.Home
		if strings.HasPrefix(h, "~") {
			if h == "~" || strings.HasPrefix(h, "~/") {
				h = filepath.Join(home, strings.TrimPrefix(h, "~/")) + string(os.PathSeparator)
			}
		}
		h = os.ExpandEnv(h)
		if !strings.HasSuffix(h, string(os.PathSeparator)) {
			h += string(os.PathSeparator)
		}
		u.Home = h
		c.Users[k] = u
	}
}

func (c *Config) validateRedis() error {
	if c.Redis.Socket == "" && c.Redis.Addr == "" {
		c.Redis.Addr = "127.0.0.1:6379" // sensible default
	}
	if c.Redis.Socket != "" && c.Redis.Addr != "" {
		return errors.New("redis.socket and redis.addr are mutually exclusive")
	}
	return nil
}

func (c *Config) validateMaria() error {
	if c.Maria.DB == "" {
		return errors.New("maria_db required")
	}
	if c.Maria.User == "" {
		return errors.New("maria_user required")
	}
	if c.Maria.Socket == "" && c.Maria.Addr == "" {
		c.Maria.Addr = "127.0.0.1:3306" // sensible default
	}
	if c.Maria.Socket != "" && c.Maria.Addr != "" {
		return errors.New("db.socket and db.addr are mutually exclusive")
	}
	return nil
}

func (c *Config) Validate() error {
	if c.ServerURL == "" {
		return errors.New("server_url is required")
	}
	if c.Users == nil || len(c.Users) == 0 {
		return errors.New("at least one user must be configured")
	}
	if c.StorageDriver == "" {
		c.StorageDriver = "disk"
	}
	switch c.StorageDriver {
	case "disk", "redis", "maria", "mariadb", "mysql", "redis-mem":
	default:
		return errors.New("unsupported storage_driver: " + c.StorageDriver)
	}
	if c.StorageDriver == "redis" {
		if err := c.validateRedis(); err != nil {
			return err
		}
	}
	if c.StorageDriver == "maria" || c.StorageDriver == "mariadb" || c.StorageDriver == "mysql" {
		if err := c.validateMaria(); err != nil {
			return err
		}
	}

	// validate tracking patterns (global + per-user effective lists)
	globalTrack := c.GlobalTrack
	if len(globalTrack) == 0 {
		globalTrack = DefaultTrack
	}
	if err := validateTrackList(globalTrack, "global track list"); err != nil {
		return err
	}
	c.GlobalTrack = globalTrack
	for name, u := range c.Users {
		if u.Home == "" {
			return errors.New("user " + name + " requires home path")
		}
		list := effectiveTrack(u.Track, c.GlobalTrack)
		if err := validateTrackList(list, "user "+name+" track list"); err != nil {
			return err
		}
	}
	return nil
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	c.path = path
	if c.Users == nil {
		c.Users = map[string]User{}
	}
	c.migrateLegacyIncludes()
	c.normalizeTracks()
	c.expand()
	return &c, nil
}

func Save(c *Config, path string) error {
	if path == "" {
		path = c.path
	}
	if path == "" {
		return errors.New("no path for config")
	}
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, fs.FileMode(0o600))
}

func (c *Config) migrateLegacyIncludes() {
	// merge legacy include keys into track lists for backwards compatibility
	if len(c.LegacyTrack) > 0 {
		c.GlobalTrack = append(c.GlobalTrack, c.LegacyTrack...)
	}
	c.LegacyTrack = nil
	for name, u := range c.Users {
		if len(u.LegacyTrack) > 0 {
			u.Track = append(u.Track, u.LegacyTrack...)
		}
		u.LegacyTrack = nil
		c.Users[name] = u
	}
}

func (c *Config) normalizeTracks() {
	c.GlobalTrack = normalizeTrackList(c.GlobalTrack)
	for name, u := range c.Users {
		u.Track = normalizeTrackList(u.Track)
		c.Users[name] = u
	}
}

func normalizeTrackList(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range in {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func effectiveTrack(userTrack, globalTrack []string) []string {
	if len(userTrack) > 0 {
		return userTrack
	}
	if len(globalTrack) > 0 {
		return globalTrack
	}
	return DefaultTrack
}

func validateTrackList(list []string, scope string) error {
	hasInclude := false
	for _, raw := range list {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "!") {
			p = strings.TrimPrefix(p, "!")
		} else {
			hasInclude = true
		}
		if ok := doublestar.ValidatePattern(filepath.ToSlash(p)); !ok {
			return fmt.Errorf("invalid %s entry %q: failed glob validation", scope, raw)
		}
	}
	if !hasInclude {
		return fmt.Errorf("%s must contain at least one non-exclusion pattern", scope)
	}
	return nil
}
