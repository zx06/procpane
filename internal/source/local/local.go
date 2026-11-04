package local

import (
	"context"

	"github.com/zx06/procpane/internal/enrich"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/scan"
	"github.com/zx06/procpane/internal/source"
)

// Source implements HostSource for the local machine.
type Source struct {
	ProtoFilter string // tcp, udp, or all
}

var _ source.HostSource = (*Source)(nil)

// New returns a local HostSource listing all protocols.
func New() *Source {
	return &Source{ProtoFilter: "all"}
}

// ListListeners scans LISTEN sockets and enriches process metadata.
func (s *Source) ListListeners(ctx context.Context) ([]model.Listener, error) {
	raw, err := scan.ListRawListeners(ctx, s.ProtoFilter)
	if err != nil {
		return nil, err
	}
	return enrich.EnrichListeners(raw), nil
}
