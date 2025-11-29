package cli

import (
	"bytes"
	"strings"
	"testing"

	"git.tyss.io/cj3636/dman/internal/config"
)

func TestConfigLintCommand(t *testing.T) {
	cfg = &config.Config{
		AuthToken:     "t",
		ServerURL:     "http://localhost:3626",
		StorageDriver: "disk",
		Users: map[string]config.User{
			"alice": {Home: "/home/alice/", Track: []string{"docs/", "!docs/config.yaml"}},
			"bob":   {Home: "/home/bob/"},
		},
	}
	defer func() { cfg = nil }()

	buf := &bytes.Buffer{}
	configLintCmd.SetOut(buf)
	if err := configLintCmd.RunE(configLintCmd, nil); err != nil {
		t.Fatalf("lint command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "configuration OK") {
		t.Fatalf("expected success output, got: %s", out)
	}
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Fatalf("expected both users in output, got: %s", out)
	}
	if !strings.Contains(out, "exclusion") {
		t.Fatalf("expected exclusion summary, got: %s", out)
	}
}
