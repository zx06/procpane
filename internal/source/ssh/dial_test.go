package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os/exec"
	"testing"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

func startTestSSHServer(t *testing.T) (addr string, cleanup func()) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		t.Fatal(err)
	}
	cfg := &gossh.ServerConfig{NoClientAuth: true}
	cfg.AddHostKey(signer)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				_, chans, reqs, err := gossh.NewServerConn(c, cfg)
				if err != nil {
					c.Close()
					return
				}
				go gossh.DiscardRequests(reqs)
				for newCh := range chans {
					if newCh.ChannelType() != "session" {
						newCh.Reject(gossh.UnknownChannelType, "")
						continue
					}
					ch, inReqs, err := newCh.Accept()
					if err != nil {
						continue
					}
					go gossh.DiscardRequests(inReqs)
					ch.Close()
				}
			}(conn)
		}
	}()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	return net.JoinHostPort("127.0.0.1", portStr), func() {
		ln.Close()
		<-done
	}
}

func TestDialProxyCommand(t *testing.T) {
	if _, err := exec.LookPath("nc"); err != nil {
		t.Skip("nc not installed")
	}
	serverAddr, stop := startTestSSHServer(t)
	defer stop()
	_, portStr, err := net.SplitHostPort(serverAddr)
	if err != nil {
		t.Fatal(err)
	}

	entry := HostEntry{
		Alias:        "proxy-test",
		Hostname:     "127.0.0.1",
		Port:         portStr,
		ProxyCommand: fmt.Sprintf(`nc -w 5 127.0.0.1 %s`, portStr),
		ConnectTimeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := dial(ctx, entry)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	client.Close()
}

func TestExpandProxyCommandTokens(t *testing.T) {
	entry := HostEntry{
		Alias:    "nas",
		Hostname: "10.0.0.1",
		Port:     "2222",
	}
	got := expandProxyCommand("echo %h %p %r %n", entry, "alice")
	want := "echo 10.0.0.1 2222 alice nas"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := expandProxyCommand("pct %%h", entry, "u"); got != "pct %h" {
		t.Fatalf("escaped %%: %q", got)
	}
}
