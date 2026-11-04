package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"sync"
	"time"
)

// proxyCmdConn wraps stdin/stdout of a ProxyCommand child as net.Conn.
type proxyCmdConn struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	once   sync.Once
	err    error
}

func newProxyCmdConn(ctx context.Context, command string) (*proxyCmdConn, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("proxycommand start: %w", err)
	}
	return &proxyCmdConn{cmd: cmd, stdin: stdin, stdout: stdout}, nil
}

func (c *proxyCmdConn) Read(b []byte) (int, error)  { return c.stdout.Read(b) }
func (c *proxyCmdConn) Write(b []byte) (int, error) { return c.stdin.Write(b) }

func (c *proxyCmdConn) Close() error {
	c.once.Do(func() {
		_ = c.stdin.Close()
		_ = c.stdout.Close()
		if c.cmd.Process != nil {
			_ = c.cmd.Process.Kill()
		}
		c.err = c.cmd.Wait()
	})
	return c.err
}

func (c *proxyCmdConn) LocalAddr() net.Addr                { return proxyAddr("local") }
func (c *proxyCmdConn) RemoteAddr() net.Addr               { return proxyAddr("remote") }
func (c *proxyCmdConn) SetDeadline(t time.Time) error      { return nil }
func (c *proxyCmdConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *proxyCmdConn) SetWriteDeadline(t time.Time) error { return nil }

type proxyAddr string

func (p proxyAddr) Network() string { return "proxycommand" }
func (p proxyAddr) String() string  { return "proxycommand:" + string(p) }
