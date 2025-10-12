package config

import (
	"errors"
	"git.tyss.io/cj3636/dman/pkg/model"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type User struct {
	Home    string   `yaml:"home" json:"home"`
	Include []string `yaml:"include" json:"include"`
}

// Redis configuration (optional when storage_driver != redis)
type Redis struct {
	Addr            string `yaml:"redis_addr" json:"redis_addr"`     // host:port for TCP
	Socket          string `yaml:"redis_socket" json:"redis_socket"` // unix domain socket path (mutually exclusive with Addr)
	Username        string `yaml:"redis_username" json:"redis_username"`
	Password        string `yaml:"redis_password" json:"redis_password"`
	DB              int    `yaml:"redis_db" json:"redis_db"`
	TLS             bool   `yaml:"redis_tls" json:"redis_tls"`
	TLSCA           string `yaml:"redis_tls_ca" json:"redis_tls_ca"`
	TLSInsecureSkip bool   `yaml:"redis_tls_insecure_skip_verify" json:"redis_tls_insecure_skip_verify"`
	TLSServerName   string `yaml:"redis_tls_server_name" json:"redis_tls_server_name"`
}

// MariaDB configuration (optional when storage_driver not in maria/mariadb/mysql)
type Maria struct {
	Addr            string `yaml:"maria_addr" json:"maria_addr"`     // host:port
	Socket          string `yaml:"maria_socket" json:"maria_socket"` // unix socket path
	DB              string `yaml:"maria_db" json:"maria_db"`
	User            string `yaml:"maria_user" json:"maria_user"`
	Password        string `yaml:"maria_password" json:"maria_password"`
	TLS             bool   `yaml:"maria_tls" json:"maria_tls"`
	TLSCA           string `yaml:"maria_tls_ca" json:"maria_tls_ca"`
	TLSCert         string `yaml:"maria_tls_cert" json:"maria_tls_cert"`
	TLSKey          string `yaml:"maria_tls_key" json:"maria_tls_key"`
	TLSInsecureSkip bool   `yaml:"maria_tls_insecure_skip_verify" json:"maria_tls_insecure_skip_verify"`
	TLSServerName   string `yaml:"maria_tls_server_name" json:"maria_tls_server_name"`
}

type Config struct {
	AuthToken     string          `yaml:"auth_token" json:"auth_token"`
	ServerURL     string          `yaml:"server_url" json:"server_url"`
	StorageDriver string          `yaml:"storage_driver" json:"storage_driver"`
	GlobalInclude []string        `yaml:"include" json:"include"`
	Users         map[string]User `yaml:"users" json:"users"`
	Redis         Redis           `yaml:",inline" json:",inline"`
	Maria         Maria           `yaml:",inline" json:",inline"`
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
		inc := u.Include
		if len(inc) == 0 { // fallback to global include, then defaults
			if len(c.GlobalInclude) > 0 {
				inc = c.GlobalInclude
			} else {
				inc = DefaultInclude
			}
		}
		out = append(out, model.UserSpec{Name: name, Home: u.Home, Include: inc})
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
		return errors.New("redis_socket and redis_addr are mutually exclusive")
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
		return errors.New("maria_socket and maria_addr are mutually exclusive")
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
