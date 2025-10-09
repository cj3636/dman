package config

import "testing"

func TestUsersListGlobalIncludeFallback(t *testing.T) {
	c := &Config{GlobalInclude: []string{"a", "b"}, Users: map[string]User{"u": {Home: "/home/u/"}}}
	us := c.UsersList()
	if len(us) != 1 {
		t.Fatalf("expected 1 user")
	}
	if len(us[0].Include) != 2 || us[0].Include[0] != "a" {
		t.Fatalf("global include not applied: %#v", us[0].Include)
	}
}

func TestUsersListDefaultIncludeFallback(t *testing.T) {
	c := &Config{Users: map[string]User{"u": {Home: "/home/u/"}}}
	us := c.UsersList()
	if len(us) != 1 {
		t.Fatalf("expected 1 user")
	}
	if len(us[0].Include) != len(DefaultInclude) {
		t.Fatalf("expected default include length %d got %d", len(DefaultInclude), len(us[0].Include))
	}
}
