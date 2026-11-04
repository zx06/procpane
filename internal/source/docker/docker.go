package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/source"
)

// Source lists published container ports via Docker Engine API (A1).
type Source struct {
	Host          string // deprecated: prefer Context; non-empty overrides host only when Context empty
	Context       string // procpane-selected Docker context name; empty → ResolveHost()
}

var _ source.HostSource = (*Source)(nil)

// New returns a Docker HostSource using the default endpoint.
func New() *Source {
	return &Source{}
}

// NewWithContext returns a Source that dials the given procpane context (empty = ResolveHost).
func NewWithContext(contextName string) *Source {
	return &Source{Context: contextName}
}

// ListListeners returns one row per published host port (ContainerList + Ports).
func (s *Source) ListListeners(ctx context.Context) ([]model.Listener, error) {
	cli, _, err := s.client()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	type publishKey struct {
		container string
		proto     model.Protocol
		pubPort   uint32
	}
	best := make(map[publishKey]model.Listener)

	for _, c := range containers {
		name := c.ID[:12]
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		shortID := c.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		image := c.Image

		for _, p := range c.Ports {
			if p.PublicPort == 0 {
				continue
			}
			var mp model.Protocol
			switch strings.ToLower(p.Type) {
			case "tcp":
				mp = model.ProtoTCP
			case "udp":
				mp = model.ProtoUDP
			default:
				continue
			}
			pubHost := p.IP
			if pubHost == "" {
				pubHost = "0.0.0.0"
			}
			pubPort := uint32(p.PublicPort)
			id := fmt.Sprintf("docker-%s-%s-%d", shortID, mp, pubPort)
			row := model.Listener{
				ID:            id,
				Port:          pubPort,
				Proto:         mp,
				PID:           0,
				Name:          name,
				Cmdline:       model.Unknown,
				Cwd:           model.Unknown,
				User:          model.Unknown,
				Started:       model.Unknown,
				Chain:         []model.ChainNode{},
				Source:        "docker",
				ContainerID:   shortID,
				ContainerName: name,
				Image:         image,
				PublishHost:   pubHost,
				PublishPort:   pubPort,
			}
			key := publishKey{container: shortID, proto: mp, pubPort: pubPort}
			if prev, ok := best[key]; !ok || publishHostRank(pubHost) > publishHostRank(prev.PublishHost) {
				best[key] = row
			}
		}
	}
	out := make([]model.Listener, 0, len(best))
	for _, l := range best {
		out = append(out, l)
	}
	return out, nil
}

// publishHostRank prefers dual-stack bindings that represent all interfaces (0.0.0.0 over ::).
func publishHostRank(host string) int {
	switch host {
	case "", "0.0.0.0":
		return 3
	case "::":
		return 1
	default:
		if strings.Contains(host, ":") {
			return 2
		}
		return 2
	}
}

func (s *Source) client() (*client.Client, string, error) {
	if s.Host != "" && s.Context == "" {
		host := s.Host
		cli, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithHost(host),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return nil, host, fmt.Errorf("docker client: %w", err)
		}
		return cli, host, nil
	}
	return NewClient(s.Context)
}

// Ping checks engine connectivity for a resolved host (legacy helper).
func Ping(ctx context.Context, host string) error {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return err
	}
	defer cli.Close()
	_, err = cli.Ping(ctx)
	return err
}

// PingContext probes the Engine for a procpane context selection (empty = ResolveHost).
func PingContext(ctx context.Context, selectedContext string) (host string, err error) {
	cli, host, err := NewClient(selectedContext)
	if err != nil {
		return "", err
	}
	defer cli.Close()
	_, err = cli.Ping(ctx)
	return host, err
}
