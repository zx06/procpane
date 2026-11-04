package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/client"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/source"
)

// Source lists published container ports via Docker Engine API (A1).
type Source struct {
	Host string // API endpoint; empty → ResolveHost()
}

var _ source.HostSource = (*Source)(nil)

// New returns a Docker HostSource using the default endpoint.
func New() *Source {
	return &Source{}
}

func (s *Source) apiHost() string {
	if s.Host != "" {
		return s.Host
	}
	return ResolveHost()
}

// ListListeners returns one row per published host port (ContainerList + Ports).
func (s *Source) ListListeners(ctx context.Context) ([]model.Listener, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithHost(s.apiHost()),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var out []model.Listener
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
			out = append(out, model.Listener{
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
			})
		}
	}
	return out, nil
}

// Ping checks engine connectivity.
func Ping(ctx context.Context, host string) error {
	if host == "" {
		host = ResolveHost()
	}
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
