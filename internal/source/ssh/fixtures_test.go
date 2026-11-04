package ssh

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kevinburke/ssh_config"
)

func testdataPath(t *testing.T, elem ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(append([]string{filepath.Dir(file), "testdata"}, elem...)...)
}

func userSettingsForFile(t *testing.T, path string) *ssh_config.UserSettings {
	t.Helper()
	us := &ssh_config.UserSettings{}
	us.ConfigFinder(func() string { return path })
	return us
}

func installUserSSHConfig(t *testing.T, configPath string) {
	t.Helper()
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// --- kevinburke/ssh_config testdata (MIT) ---

func TestFixtureKevinburkeConfig1Aliases(t *testing.T) {
	data, err := os.ReadFile(testdataPath(t, "kevinburke-ssh_config", "config1"))
	if err != nil {
		t.Fatal(err)
	}
	got := parseHostAliases(string(data))
	for _, want := range []string{"localhost", "127.0.0.1", "wap", "wopr", "dhcp-??"} {
		if _, ok := sliceContains(got, want); !ok {
			t.Errorf("missing alias %q in %v", want, got)
		}
	}
	for _, a := range got {
		if a == "*" {
			t.Errorf("parseHostAliases should skip %q", a)
		}
	}
}

func TestFixtureKevinburkeConfig1ProxyCommand(t *testing.T) {
	path := testdataPath(t, "kevinburke-ssh_config", "config1")
	us := userSettingsForFile(t, path)
	entry := resolveHostEntryWith("wopr", us)
	if entry.User != "root" {
		t.Fatalf("User: %q", entry.User)
	}
	if !strings.Contains(entry.ProxyCommand, "proxy1") || !strings.Contains(entry.ProxyCommand, "proxy2") {
		t.Fatalf("ProxyCommand: %q", entry.ProxyCommand)
	}
}

func TestFixtureKevinburkeConfig3HostStarMerge(t *testing.T) {
	// Expected ports from kevinburke/ssh_config config_test.go TestGetWildcard.
	path := testdataPath(t, "kevinburke-ssh_config", "config3")
	us := userSettingsForFile(t, path)
	cases := []struct {
		alias string
		port  string
	}{
		{"bastion.stage.i.us.example.net", "22"},
		{"bastion.net", "25"},
		{"10.2.3.4", "23"},
		{"101.2.3.4", "25"},
		{"20.20.20.4", "24"},
		{"20.20.20.20", "25"},
	}
	for _, tc := range cases {
		entry := resolveHostEntryWith(tc.alias, us)
		if entry.Port != tc.port {
			t.Errorf("alias %q: port got %q want %q", tc.alias, entry.Port, tc.port)
		}
	}
}

func TestFixtureKevinburkeNegatedHost(t *testing.T) {
	path := testdataPath(t, "kevinburke-ssh_config", "negated")
	us := userSettingsForFile(t, path)
	if p := resolveHostEntryWith("bbc.example.com", us).Port; p != "1234" {
		t.Fatalf("wildcard port: %q", p)
	}
	if p := resolveHostEntryWith("foo.dialup.example.com", us).Port; p != "5678" {
		t.Fatalf("negated match should fall through to Host *: got %q", p)
	}
}

func TestFixtureKevinburkeMultiIdentityFile(t *testing.T) {
	path := testdataPath(t, "kevinburke-ssh_config", "identities")
	us := userSettingsForFile(t, path)
	entry := resolveHostEntryWith("has2identity", us)
	if len(entry.IdentityFiles) != 2 || entry.IdentityFiles[0] != "f1" || entry.IdentityFiles[1] != "f2" {
		t.Fatalf("IdentityFiles: %v", entry.IdentityFiles)
	}
}

func TestFixtureKevinburkeInclude(t *testing.T) {
	// Same layout as kevinburke/ssh_config TestInclude (writes included target under ~/.ssh).
	if testing.Short() {
		t.Skip("skipping fs write in short mode")
	}
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	includeTarget := []byte(`
Host kevinburke.ssh_config.test.example.com
    Port 4567
`)
	if err := os.WriteFile(filepath.Join(sshDir, "kevinburke-ssh-config-test-file"), includeTarget, 0o644); err != nil {
		t.Fatal(err)
	}
	mainCfg, err := os.ReadFile(testdataPath(t, "kevinburke-ssh_config", "include"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "config"), mainCfg, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(merged, "Port 4567") {
		t.Fatalf("Include not expanded in merged config")
	}
	aliases := parseHostAliases(merged)
	found := false
	for _, a := range aliases {
		if a == "kevinburke.ssh_config.test.example.com" {
			found = true
		}
	}
	if !found {
		t.Fatalf("included host alias not listed: %v", aliases)
	}

}

func TestFixtureKevinburkeConnectTimeoutDefaultPath(t *testing.T) {
	path := testdataPath(t, "kevinburke-ssh_config", "extraspace")
	us := userSettingsForFile(t, path)
	if p := resolveHostEntryWith("test.test", us).Port; p != "1234" {
		t.Fatalf("Port: %q", p)
	}
}

// --- cocotyty/sshconf test_cases ---

func TestFixtureCocotytyIncludeAliases(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh", "included")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sub1.conf", "sub2.conf"} {
		src := testdataPath(t, "cocotyty-sshconf", "basic", "included", name)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sshDir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	main, err := os.ReadFile(testdataPath(t, "cocotyty-sshconf", "basic", "include_main.conf"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".ssh"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".ssh", "config"), main, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	merged, err := readMergedConfig()
	if err != nil {
		t.Fatal(err)
	}
	aliases := parseHostAliases(merged)
	for _, want := range []string{"included1", "included2"} {
		ok := false
		for _, a := range aliases {
			if a == want {
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("missing include alias %q in %v", want, aliases)
		}
	}
}

func TestFixtureCocotytyEnterpriseProxyCommand(t *testing.T) {
	path := testdataPath(t, "cocotyty-sshconf", "enterprise.conf")
	us := userSettingsForFile(t, path)
	entry := resolveHostEntryWith("prod-db", us)
	if !strings.Contains(entry.ProxyCommand, "bastion") {
		t.Fatalf("ProxyCommand: %q", entry.ProxyCommand)
	}
}

func TestFixtureCocotytySimpleHostStarUser(t *testing.T) {
	path := testdataPath(t, "cocotyty-sshconf", "basic", "simple.conf")
	us := userSettingsForFile(t, path)
	if u := resolveHostEntryWith("web", us).User; u != "webuser" {
		t.Fatalf("web user: %q", u)
	}
	if u := resolveHostEntryWith("unknown-host", us).User; u != "defaultuser" {
		t.Fatalf("Host * merge user: %q", u)
	}
}

// --- OpenSSH regress excerpts (Public Domain) ---

func TestFixtureOpenSSHProxyJumpChain(t *testing.T) {
	path := testdataPath(t, "openssh-regress", "proxyjump-chain.conf")
	us := userSettingsForFile(t, path)
	entry := resolveHostEntryWith("viajump", us)
	if entry.Hostname != "final.example.com" {
		t.Fatalf("Hostname: %q", entry.Hostname)
	}
	if len(entry.ProxyJump) != 2 || entry.ProxyJump[0] != "jump1" || entry.ProxyJump[1] != "jump2" {
		t.Fatalf("ProxyJump: %v", entry.ProxyJump)
	}
}

func TestFixtureOpenSSHConnectTimeout(t *testing.T) {
	path := testdataPath(t, "openssh-regress", "connecttimeout-host.conf")
	us := userSettingsForFile(t, path)
	entry := resolveHostEntryWith("timeout-host", us)
	if entry.ConnectTimeout != 42*time.Second {
		t.Fatalf("ConnectTimeout: %v", entry.ConnectTimeout)
	}
}

func TestFixtureOpenSSHIdentitiesOnly(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "user_a")
	if err := os.WriteFile(keyPath, []byte("not-a-real-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(testdataPath(t, "openssh-regress", "identities-only-agent-restrict.conf"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := strings.ReplaceAll(string(raw), "REPLACED_USER_A_KEY", keyPath)
	if err := os.WriteFile(filepath.Join(sshDir, "config"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("SSH_AUTH_SOCK", "")

	us := userSettingsForFile(t, filepath.Join(sshDir, "config"))
	entry := resolveHostEntryWith("host_a", us)
	if !entry.IdentitiesOnly {
		t.Fatal("expected IdentitiesOnly yes")
	}
	if len(entry.IdentityFiles) != 1 || entry.IdentityFiles[0] != keyPath {
		t.Fatalf("IdentityFiles: %v", entry.IdentityFiles)
	}
	if _, err := authMethods(entry); err == nil {
		t.Fatal("expected auth error for non-parseable key file without defaults")
	}
}
