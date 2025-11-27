package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUsersListGlobalTrackFallback(t *testing.T) {
	c := &Config{GlobalTrack: []string{"a", "b"}, Users: map[string]User{"u": {Home: "/home/u/"}}}
	us := c.UsersList()
	if len(us) != 1 {
		t.Fatalf("expected 1 user")
	}
	if len(us[0].Track) != 2 || us[0].Track[0] != "a" {
		t.Fatalf("global track not applied: %#v", us[0].Track)
	}
}

func TestUsersListDefaultTrackFallback(t *testing.T) {
	c := &Config{Users: map[string]User{"u": {Home: "/home/u/"}}}
	us := c.UsersList()
	if len(us) != 1 {
		t.Fatalf("expected 1 user")
	}
	if len(us[0].Track) != len(DefaultTrack) {
		t.Fatalf("expected default track length %d got %d", len(DefaultTrack), len(us[0].Track))
	}
}

func TestValidateInjectsDefaultTrack(t *testing.T) {
	c := &Config{ServerURL: "http://localhost:3626", Users: map[string]User{
		"u": {Home: "/home/u/", Track: nil},
	}}
	if err := c.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if len(c.GlobalTrack) != len(DefaultTrack) {
		t.Fatalf("expected global track to be default, got %v", c.GlobalTrack)
	}
}

func TestValidateRejectsAllExclusions(t *testing.T) {
	c := &Config{ServerURL: "http://localhost:3626", Users: map[string]User{
		"u": {Home: "/home/u/", Track: []string{"!docs/config.yaml"}},
	}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected validation failure for exclusion-only track")
	}
}

func TestValidateCatchesBadPattern(t *testing.T) {
	c := &Config{ServerURL: "http://localhost:3626", Users: map[string]User{
		"u": {Home: "/home/u/", Track: []string{"[oops"}},
	}}
	if err := c.Validate(); err == nil {
		t.Fatalf("expected validation failure for malformed glob")
	}
}

func TestLoadMigratesLegacyInclude(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "dman.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
auth_token: t
server_url: http://localhost:3626
include:
  - docs/
users:
  alice:
    home: /home/alice/
    include:
      - ".bashrc"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.GlobalTrack) != 1 || cfg.GlobalTrack[0] != "docs/" {
		t.Fatalf("legacy include not migrated: %#v", cfg.GlobalTrack)
	}
	if u := cfg.Users["alice"]; len(u.Track) != 1 || u.Track[0] != ".bashrc" {
		t.Fatalf("user legacy include not migrated: %#v", u.Track)
	}
}
