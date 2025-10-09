package config

import (
	"errors"
	"git.tyss.io/cj3636/dman/pkg/model"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type User struct {
	Home    string   `yaml:"home" json:"home"`
	Include []string `yaml:"include" json:"include"`
}

type Config struct {
	AuthToken     string          `yaml:"auth_token" json:"auth_token"`
	ServerURL     string          `yaml:"server_url" json:"server_url"`
	GlobalInclude []string        `yaml:"include" json:"include"`
	Users         map[string]User `yaml:"users" json:"users"`
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
