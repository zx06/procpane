package ssh

import (
	"context"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

const poolIdleTTL = 2 * time.Minute

type pooledClient struct {
	client   *gossh.Client
	lastUsed time.Time
}

var sshPool = struct {
	sync.Mutex
	m map[string]*pooledClient
}{m: make(map[string]*pooledClient)}

func acquireSSHClient(ctx context.Context, entry HostEntry) (*gossh.Client, error) {
	alias := entry.Alias
	sshPool.Lock()
	if pc := sshPool.m[alias]; pc != nil && pc.client != nil {
		if time.Since(pc.lastUsed) < poolIdleTTL && clientAlive(pc.client) {
			pc.lastUsed = time.Now()
			c := pc.client
			sshPool.Unlock()
			return c, nil
		}
		if pc.client != nil {
			_ = pc.client.Close()
		}
		delete(sshPool.m, alias)
	}
	sshPool.Unlock()

	client, err := dial(ctx, entry)
	if err != nil {
		return nil, err
	}
	sshPool.Lock()
	sshPool.m[alias] = &pooledClient{client: client, lastUsed: time.Now()}
	sshPool.Unlock()
	return client, nil
}

func touchSSHClient(alias string) {
	sshPool.Lock()
	if pc := sshPool.m[alias]; pc != nil {
		pc.lastUsed = time.Now()
	}
	sshPool.Unlock()
}

func dropSSHClient(alias string) {
	sshPool.Lock()
	if pc := sshPool.m[alias]; pc != nil {
		if pc.client != nil {
			_ = pc.client.Close()
		}
		delete(sshPool.m, alias)
	}
	sshPool.Unlock()
}

func clientAlive(client *gossh.Client) bool {
	sess, err := client.NewSession()
	if err != nil {
		return false
	}
	_ = sess.Close()
	return true
}
