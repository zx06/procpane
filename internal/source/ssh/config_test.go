package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHostAliases(t *testing.T) {
	cfg := `
Host jump
  HostName jump.example.com

Host app prod
  HostName 10.0.0.1

Host *
  ForwardAgent yes
`
	got := parseHostAliases(cfg)
	want := map[string]bool{"jump": true, "app": true, "prod": true}
	for _, a := range got {
		if !want[a] {
			t.Errorf("unexpected alias %q", a)
		}
		delete(want, a)
	}
	if len(want) != 0 {
		t.Fatalf("missing aliases: %v", want)
	}
}

func TestReadMergedConfigHome(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "Host demo\n  HostName demo.local\n"
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	aliases := parseHostAliases(merged)
	if len(aliases) != 1 || aliases[0] != "demo" {
		t.Fatalf("aliases: %v", aliases)
	}
}
