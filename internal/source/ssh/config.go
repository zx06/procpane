package ssh

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kevinburke/ssh_config"
)

// HostEntry is one OpenSSH config Host alias with resolved connection fields.
type HostEntry struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	Port     string `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Identity string `json:"identity,omitempty"`
}

var hostLineRe = regexp.MustCompile(`(?mi)^\s*Host\s+(.+)$`)

// ListHosts parses ~/.ssh/config (+ system defaults) and returns Host aliases.
func ListHosts() ([]HostEntry, error) {
	merged, err := readMergedConfig()
	if err != nil {
		return nil, err
	}
	aliases := parseHostAliases(merged)
	seen := make(map[string]struct{})
	var out []HostEntry
	for _, alias := range aliases {
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		out = append(out, ResolveHostEntry(alias))
	}
	return out, nil
}

func readMergedConfig() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var parts []string
	if b, err := os.ReadFile(filepath.Join(home, ".ssh", "config")); err == nil {
		parts = append(parts, string(b))
	} else if !os.IsNotExist(err) {
		return "", err
	}
	for _, p := range []string{"/etc/ssh/ssh_config"} {
		if b, err := os.ReadFile(p); err == nil {
			parts = append(parts, string(b))
		}
	}
	return strings.Join(parts, "\n"), nil
}

func parseHostAliases(configText string) []string {
	var out []string
	for _, m := range hostLineRe.FindAllStringSubmatch(configText, -1) {
		if len(m) < 2 {
			continue
		}
		for _, tok := range strings.Fields(m[1]) {
			if tok == "*" || strings.HasPrefix(tok, "!") {
				continue
			}
			out = append(out, tok)
		}
	}
	return out
}

// ResolveHost returns connection target for alias (Hostname, Port, User, IdentityFile).
func ResolveHost(alias string) (HostEntry, error) {
	if _, err := readMergedConfig(); err != nil {
		return HostEntry{}, err
	}
	return ResolveHostEntry(alias), nil
}

func ResolveHostEntry(alias string) HostEntry {
	entry := HostEntry{
		Alias:    alias,
		Hostname: ssh_config.Get(alias, "Hostname"),
		Port:     ssh_config.Get(alias, "Port"),
		User:     ssh_config.Get(alias, "User"),
		Identity: ssh_config.Get(alias, "IdentityFile"),
	}
	if entry.Hostname == "" {
		entry.Hostname = alias
	}
	return entry
}
