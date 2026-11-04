package ssh

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
)

// HostEntry is one OpenSSH config Host alias with resolved connection fields.
type HostEntry struct {
	Alias           string        `json:"alias"`
	Hostname        string        `json:"hostname"`
	Port            string        `json:"port,omitempty"`
	User            string        `json:"user,omitempty"`
	Identity        string        `json:"identity,omitempty"` // first IdentityFile (compat)
	IdentityFiles   []string      `json:"identityFiles,omitempty"`
	IdentitiesOnly  bool          `json:"identitiesOnly,omitempty"`
	ProxyCommand    string        `json:"proxyCommand,omitempty"`
	ProxyJump       []string      `json:"proxyJump,omitempty"`
	ConnectTimeout  time.Duration `json:"connectTimeout,omitempty"`
}

var hostLineRe = regexp.MustCompile(`(?mi)^\s*Host\s+(.+)$`)
var includeLineRe = regexp.MustCompile(`(?mi)^\s*Include\s+(.+)$`)

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
	userCfg := filepath.Join(home, ".ssh", "config")
	if b, err := os.ReadFile(userCfg); err == nil {
		expanded, err := expandIncludes(string(b), filepath.Join(home, ".ssh"), false, 0)
		if err != nil {
			return "", err
		}
		parts = append(parts, expanded)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	for _, p := range []string{"/etc/ssh/ssh_config"} {
		if b, err := os.ReadFile(p); err == nil {
			expanded, err := expandIncludes(string(b), filepath.Dir(p), true, 0)
			if err != nil {
				return "", err
			}
			parts = append(parts, expanded)
		}
	}
	return strings.Join(parts, "\n"), nil
}

func expandIncludes(text, baseDir string, system bool, depth int) (string, error) {
	if depth > 5 {
		return "", ssh_config.ErrDepthExceeded
	}
	var out strings.Builder
	for _, line := range strings.Split(text, "\n") {
		m := includeLineRe.FindStringSubmatch(line)
		if m == nil {
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		fields := strings.Fields(stripInlineComment(m[1]))
		for _, tok := range fields {
			globPath := resolveIncludePath(tok, baseDir, system)
			matches, err := filepath.Glob(globPath)
			if err != nil {
				return "", err
			}
			for _, mf := range matches {
				b, err := os.ReadFile(mf)
				if err != nil {
					continue
				}
				nested, err := expandIncludes(string(b), filepath.Dir(mf), system, depth+1)
				if err != nil {
					return "", err
				}
				out.WriteString(nested)
				if !strings.HasSuffix(nested, "\n") {
					out.WriteByte('\n')
				}
			}
		}
	}
	return out.String(), nil
}

func stripInlineComment(s string) string {
	if i := strings.IndexByte(s, '#'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

func resolveIncludePath(tok, baseDir string, system bool) string {
	if filepath.IsAbs(tok) {
		return tok
	}
	if system {
		return filepath.Join("/etc/ssh", tok)
	}
	return filepath.Join(baseDir, tok)
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

// ResolveHostEntry resolves OpenSSH keywords for alias via kevinburke/ssh_config
// (including Include files on Get/GetAll).
func ResolveHostEntry(alias string) HostEntry {
	return resolveHostEntryWith(alias, ssh_config.DefaultUserSettings)
}

func resolveHostEntryWith(alias string, us *ssh_config.UserSettings) HostEntry {
	identityFiles := us.GetAll(alias, "IdentityFile")
	entry := HostEntry{
		Alias:          alias,
		Hostname:       us.Get(alias, "Hostname"),
		Port:           us.Get(alias, "Port"),
		User:           us.Get(alias, "User"),
		IdentityFiles:  expandIdentityPaths(identityFiles),
		IdentitiesOnly: parseIdentitiesOnly(us.Get(alias, "IdentitiesOnly")),
		ProxyCommand:   us.Get(alias, "ProxyCommand"),
		ProxyJump:      parseProxyJumpValues(us.GetAll(alias, "ProxyJump")),
		ConnectTimeout: parseConnectTimeout(us.Get(alias, "ConnectTimeout")),
	}
	if len(entry.IdentityFiles) > 0 {
		entry.Identity = entry.IdentityFiles[0]
	} else if id := us.Get(alias, "IdentityFile"); id != "" {
		entry.Identity = expandHome(id)
	}
	if entry.Hostname == "" {
		entry.Hostname = alias
	}
	return entry
}

func expandIdentityPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		out = append(out, expandHome(p))
	}
	return out
}
