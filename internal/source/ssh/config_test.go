package ssh

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMergedConfigUsesHomeConfig(t *testing.T) {
	installUserSSHConfig(t, testdataPath(t, "cocotyty-sshconf", "basic", "simple.conf"))
	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	aliases := parseHostAliases(merged)
	if len(aliases) < 3 {
		t.Fatalf("expected simple.conf hosts, got %v", aliases)
	}
}

func TestParseHostAliasesSkipsGlobAndNegation(t *testing.T) {
	data, err := os.ReadFile(testdataPath(t, "cocotyty-sshconf", "basic", "wildcards.conf"))
	if err != nil {
		t.Fatal(err)
	}
	got := parseHostAliases(string(data))
	for _, a := range got {
		if a == "*" || strings.HasPrefix(a, "!") {
			t.Fatalf("parseHostAliases should skip %q", a)
		}
	}
	if _, ok := sliceContains(got, "prod.server.example.com"); !ok {
		t.Fatalf("missing prod host in %v", got)
	}
}

func TestMixedIncludePathsTildeAbsoluteAndConfigD(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	colimaDir := filepath.Join(dir, "colima")
	sealosDir := filepath.Join(sshDir, "sealos")
	configD := filepath.Join(sshDir, "config.d")
	for _, d := range []string{sshDir, colimaDir, sealosDir, configD} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(colimaDir, "ssh_config"), []byte(`
Host colima-vm
    Hostname 127.0.0.1
    User colima
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sealosDir, "devbox_config"), []byte(`
Host sealos-devbox
    Hostname devbox.sealos.example
    User sealos
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configD, "extra.conf"), []byte(`
Host configd-extra
    Hostname extra.internal
`), 0o644); err != nil {
		t.Fatal(err)
	}
	mainCfg := strings.Join([]string{
		"Include " + filepath.Join(colimaDir, "ssh_config"),
		"Include ~/.ssh/sealos/devbox_config",
		"Include config.d/*",
		"Host *",
		"  Compression yes",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(mainCfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	aliases := parseHostAliases(merged)
	for _, want := range []string{"colima-vm", "sealos-devbox", "configd-extra"} {
		if _, ok := sliceContains(aliases, want); !ok {
			t.Fatalf("missing alias %q from merged config; got %v", want, aliases)
		}
	}

	hosts, err := ListHosts()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"colima-vm", "sealos-devbox", "configd-extra"} {
		found := false
		for _, h := range hosts {
			if h.Alias == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("ListHosts missing %q; got %d entries", want, len(hosts))
		}
	}
	for _, h := range hosts {
		switch h.Alias {
		case "colima-vm":
			if h.Hostname != "127.0.0.1" || h.User != "colima" {
				t.Fatalf("ListHosts colima-vm: %+v", h)
			}
		case "sealos-devbox":
			if h.Hostname != "devbox.sealos.example" || h.User != "sealos" {
				t.Fatalf("ListHosts sealos-devbox: %+v", h)
			}
		}
	}

	entry := ResolveHostEntry("sealos-devbox")
	if entry.Hostname != "devbox.sealos.example" {
		t.Fatalf("ResolveHostEntry Hostname: got %q want devbox.sealos.example", entry.Hostname)
	}
	if entry.User != "sealos" {
		t.Fatalf("ResolveHostEntry User: got %q want sealos", entry.User)
	}
	entry = ResolveHostEntry("colima-vm")
	if entry.Hostname != "127.0.0.1" || entry.User != "colima" {
		t.Fatalf("absolute Include resolve: %+v", entry)
	}
}

func sliceContains(ss []string, v string) (string, bool) {
	for _, s := range ss {
		if s == v {
			return s, true
		}
	}
	return "", false
}
