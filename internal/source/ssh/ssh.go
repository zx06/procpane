package ssh

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	gossh "golang.org/x/crypto/ssh"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/source"
)

// Source implements HostSource for one OpenSSH Host alias.
type Source struct {
	HostAlias string
}

var _ source.HostSource = (*Source)(nil)

// New returns an SSH HostSource for the given config alias.
func New(alias string) *Source {
	return &Source{HostAlias: alias}
}

// ListListeners runs ss → lsof → netstat on the remote host (best-effort enrich).
func (s *Source) ListListeners(ctx context.Context) ([]model.Listener, error) {
	entry, err := ResolveHost(s.HostAlias)
	if err != nil {
		return nil, err
	}
	client, err := dial(ctx, entry)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	raw, tool, err := collectRemote(ctx, client)
	if err != nil {
		return nil, err
	}
	listeners := rawToListeners(s.HostAlias, raw)
	enrichRemote(ctx, client, tool, listeners)
	return listeners, nil
}

func dial(ctx context.Context, entry HostEntry) (*gossh.Client, error) {
	port := entry.Port
	if port == "" {
		port = "22"
	}
	addr := net.JoinHostPort(entry.Hostname, port)
	conn, err := dialTransport(ctx, entry)
	if err != nil {
		return nil, fmt.Errorf("ssh dial %s: %w", entry.Alias, err)
	}
	client, err := handshakeSSHClient(ctx, conn, addr, entry)
	if err != nil {
		return nil, fmt.Errorf("ssh handshake %s: %w", entry.Alias, err)
	}
	return client, nil
}

func authMethods(entry HostEntry) ([]gossh.AuthMethod, error) {
	var methods []gossh.AuthMethod
	addKey := func(path string) {
		path = expandHome(path)
		if path == "" {
			return
		}
		if key, err := loadPrivateKey(path); err == nil {
			methods = append(methods, gossh.PublicKeys(key))
		}
	}
	for _, id := range entry.IdentityFiles {
		addKey(id)
	}
	if entry.Identity != "" && len(entry.IdentityFiles) == 0 {
		addKey(entry.Identity)
	}
	if !entry.IdentitiesOnly {
		if sock := os.Getenv("SSH_AUTH_SOCK"); sock != "" {
			if signers, err := signersFromAgent(sock); err == nil && len(signers) > 0 {
				methods = append(methods, gossh.PublicKeys(signers...))
			}
		}
		home, _ := os.UserHomeDir()
		for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
			addKey(filepath.Join(home, ".ssh", name))
		}
	}
	if len(methods) == 0 {
		return nil, fmt.Errorf("no ssh keys or agent available for %q (password auth unsupported)", entry.Alias)
	}
	return methods, nil
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}

func loadPrivateKey(path string) (gossh.Signer, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(b)
}

func collectRemote(ctx context.Context, client *gossh.Client) ([]rawListen, string, error) {
	type attempt struct {
		tool string
		cmd  string
	}
	attempts := []attempt{
		{"ss", "ss -H -ltnp 2>/dev/null; ss -H -lunp 2>/dev/null"},
		{"lsof", "lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null; lsof -nP -iUDP 2>/dev/null"},
		{"netstat", "netstat -ltnp 2>/dev/null; netstat -lunp 2>/dev/null"},
	}
	for _, a := range attempts {
		out, err := runRemote(ctx, client, a.cmd)
		if err != nil {
			continue
		}
		raw := parseListenOutput(a.tool, out)
		if len(raw) > 0 {
			return raw, a.tool, nil
		}
	}
	return nil, "", fmt.Errorf("remote listen scan failed (ss/lsof/netstat)")
}

func runRemote(ctx context.Context, client *gossh.Client, cmd string) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	var buf bytes.Buffer
	sess.Stdout = &buf
	sess.Stderr = &buf
	done := make(chan error, 1)
	go func() { done <- sess.Run(cmd) }()
	select {
	case <-ctx.Done():
		_ = sess.Close()
		return "", ctx.Err()
	case err := <-done:
		if err != nil && buf.Len() == 0 {
			return "", err
		}
		return buf.String(), nil
	}
}

func enrichRemote(ctx context.Context, client *gossh.Client, tool string, listeners []model.Listener) {
	for i := range listeners {
		if listeners[i].PID <= 0 {
			continue
		}
		cmd := fmt.Sprintf("ps -p %d -o comm= 2>/dev/null || true", listeners[i].PID)
		out, err := runRemote(ctx, client, cmd)
		if err != nil {
			continue
		}
		name := strings.TrimSpace(out)
		if name != "" {
			listeners[i].Name = name
		}
		cmdlineCmd := fmt.Sprintf("tr '\\0' ' ' < /proc/%d/cmdline 2>/dev/null || ps -p %d -o args= 2>/dev/null", listeners[i].PID, listeners[i].PID)
		if out, err := runRemote(ctx, client, cmdlineCmd); err == nil {
			if c := strings.TrimSpace(out); c != "" {
				listeners[i].Cmdline = c
			}
		}
		_ = tool
	}
}
