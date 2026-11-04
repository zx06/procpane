package source

import (
	"context"

	"github.com/zx06/procpane/internal/model"
)

// HostSource lists listening sockets on a host (local, docker, ssh, …).
type HostSource interface {
	ListListeners(ctx context.Context) ([]model.Listener, error)
}
