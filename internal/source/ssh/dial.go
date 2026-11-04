package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

const defaultConnectTimeout = 15 * time.Second

func effectiveUsername(entry HostEntry) string {
	if entry.User != "" {
		return entry.User
	}
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return "root"
}

func expandProxyCommand(cmd string, entry HostEntry, username string) string {
	port := entry.Port
	if port == "" {
		port = "22"
	}
	firstID := ""
	if len(entry.IdentityFiles) > 0 {
		firstID = entry.IdentityFiles[0]
	} else if entry.Identity != "" {
		firstID = expandHome(entry.Identity)
	}
	const esc = "\x00PERCENT\x00"
	r := strings.NewReplacer(
		"%%", esc,
		"%h", entry.Hostname,
		"%p", port,
		"%r", username,
		"%n", entry.Alias,
		"%L", localShortHostname(),
		"%l", localFQDN(),
		"%i", firstID,
	)
	s := r.Replace(cmd)
	return strings.ReplaceAll(s, esc, "%")
}

func localShortHostname() string {
	if h, err := os.Hostname(); err == nil {
		if i := strings.IndexByte(h, '.'); i > 0 {
			return h[:i]
		}
		return h
	}
	return "localhost"
}

func localFQDN() string {
	if h, err := os.Hostname(); err == nil {
		return h
	}
	return "localhost"
}

func parseConnectTimeout(raw string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultConnectTimeout
	}
	sec, err := strconv.Atoi(raw)
	if err != nil || sec <= 0 {
		return defaultConnectTimeout
	}
	return time.Duration(sec) * time.Second
}

func parseIdentitiesOnly(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "yes", "true", "1":
		return true
	default:
		return false
	}
}

func parseProxyJumpValues(vals []string) []string {
	var out []string
	for _, v := range vals {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

// chainConn closes underlying SSH clients when the forwarded connection closes.
type chainConn struct {
	net.Conn
	cleanup func()
}

func (c *chainConn) Close() error {
	err := c.Conn.Close()
	if c.cleanup != nil {
		c.cleanup()
	}
	return err
}

func dialTransport(ctx context.Context, entry HostEntry) (net.Conn, error) {
	username := effectiveUsername(entry)
	timeout := entry.ConnectTimeout
	if timeout <= 0 {
		timeout = defaultConnectTimeout
	}
	port := entry.Port
	if port == "" {
		port = "22"
	}
	addr := net.JoinHostPort(entry.Hostname, port)

	if strings.TrimSpace(entry.ProxyCommand) != "" {
		cmd := expandProxyCommand(entry.ProxyCommand, entry, username)
		return newProxyCmdConn(ctx, cmd)
	}
	if len(entry.ProxyJump) > 0 {
		return dialProxyJump(ctx, entry, timeout)
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// dialProxyJump opens target hostname:port via OpenSSH ProxyJump hosts.
// Each hop uses its own ssh_config (ProxyCommand, nested ProxyJump, etc.).
// Limitation: Match blocks and URI-style -J are not supported; ConnectTimeout applies per hop from that HostEntry.
func dialProxyJump(ctx context.Context, target HostEntry, timeout time.Duration) (net.Conn, error) {
	jumps := target.ProxyJump
	if len(jumps) == 0 {
		return nil, fmt.Errorf("proxyjump: empty chain for %q", target.Alias)
	}

	var clients []*gossh.Client
	closeAll := func() {
		for i := len(clients) - 1; i >= 0; i-- {
			_ = clients[i].Close()
		}
	}

	first := ResolveHostEntry(jumps[0])
	if first.ConnectTimeout <= 0 {
		first.ConnectTimeout = timeout
	}
	client, err := dialSSHClient(ctx, first)
	if err != nil {
		return nil, fmt.Errorf("proxyjump %q: %w", jumps[0], err)
	}
	clients = append(clients, client)

	for _, alias := range jumps[1:] {
		hop := ResolveHostEntry(alias)
		if hop.ConnectTimeout <= 0 {
			hop.ConnectTimeout = timeout
		}
		hopPort := hop.Port
		if hopPort == "" {
			hopPort = "22"
		}
		hopAddr := net.JoinHostPort(hop.Hostname, hopPort)
		fwd, err := client.Dial("tcp", hopAddr)
		if err != nil {
			closeAll()
			return nil, fmt.Errorf("proxyjump forward to %q: %w", alias, err)
		}
		client, err = handshakeSSHClient(ctx, fwd, hopAddr, hop)
		if err != nil {
			fwd.Close()
			closeAll()
			return nil, fmt.Errorf("proxyjump handshake %q: %w", alias, err)
		}
		clients = append(clients, client)
	}

	targetPort := target.Port
	if targetPort == "" {
		targetPort = "22"
	}
	finalAddr := net.JoinHostPort(target.Hostname, targetPort)
	conn, err := client.Dial("tcp", finalAddr)
	if err != nil {
		closeAll()
		return nil, fmt.Errorf("proxyjump to target %q: %w", target.Alias, err)
	}
	return &chainConn{Conn: conn, cleanup: closeAll}, nil
}

func dialSSHClient(ctx context.Context, entry HostEntry) (*gossh.Client, error) {
	port := entry.Port
	if port == "" {
		port = "22"
	}
	addr := net.JoinHostPort(entry.Hostname, port)
	conn, err := dialTransport(ctx, entry)
	if err != nil {
		return nil, err
	}
	return handshakeSSHClient(ctx, conn, addr, entry)
}

func handshakeSSHClient(ctx context.Context, conn net.Conn, addr string, entry HostEntry) (*gossh.Client, error) {
	timeout := entry.ConnectTimeout
	if timeout <= 0 {
		timeout = defaultConnectTimeout
	}
	authMethods, err := authMethods(entry)
	if err != nil {
		conn.Close()
		return nil, err
	}
	cfg := &gossh.ClientConfig{
		User:            effectiveUsername(entry),
		Auth:            authMethods,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}
	_ = ctx
	c, chans, reqs, err := gossh.NewClientConn(conn, addr, cfg)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return gossh.NewClient(c, chans, reqs), nil
}
