package docker

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/docker/cli/cli/command"
	cliconfig "github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/client"
)

// ContextSummary is one Docker CLI context for UI/CLI listing.
type ContextSummary struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Host          string `json:"host"`
	SkipTLSVerify bool   `json:"skipTLSVerify,omitempty"`
	Current       bool   `json:"current"`  // Docker CLI current (*)
	Selected      bool   `json:"selected"` // procpane persisted choice
	Error         string `json:"error,omitempty"`
}

func contextStore() (*command.ContextStoreWithDefault, error) {
	cfg := command.DefaultContextStoreConfig()
	return &command.ContextStoreWithDefault{
		Store: store.New(cliconfig.ContextStoreDir(), cfg),
		Resolver: func() (*command.DefaultContext, error) {
			return command.ResolveDefaultContext(&cliflags.ClientOptions{}, cfg)
		},
	}, nil
}

// DockerCLICurrentContext returns the name Docker CLI would use (config + DOCKER_CONTEXT).
func DockerCLICurrentContext() string {
	if os.Getenv("DOCKER_HOST") != "" {
		return command.DefaultContextName
	}
	if ctx := os.Getenv(command.EnvOverrideContext); ctx != "" {
		return ctx
	}
	cfg := cliconfig.LoadDefaultConfigFile(io.Discard)
	if cfg != nil && cfg.CurrentContext != "" {
		return cfg.CurrentContext
	}
	return command.DefaultContextName
}

// ListContexts reads ~/.docker/contexts (and virtual default) without shelling out to docker.
func ListContexts(selected string) ([]ContextSummary, error) {
	cs, err := contextStore()
	if err != nil {
		return nil, err
	}
	metaList, err := cs.List()
	if err != nil {
		return nil, err
	}
	cur := DockerCLICurrentContext()
	out := make([]ContextSummary, 0, len(metaList))
	for _, raw := range metaList {
		row := ContextSummary{
			Name:     raw.Name,
			Current:  raw.Name == cur,
			Selected: selected != "" && raw.Name == selected,
		}
		dc, err := command.GetDockerContext(raw)
		if err != nil {
			row.Error = err.Error()
			out = append(out, row)
			continue
		}
		row.Description = dc.Description
		epMeta, err := docker.EndpointFromContext(raw)
		if err != nil {
			row.Error = err.Error()
		} else {
			row.Host = epMeta.Host
			row.SkipTLSVerify = epMeta.SkipTLSVerify
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == command.DefaultContextName {
			return true
		}
		if out[j].Name == command.DefaultContextName {
			return false
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// EndpointForContext resolves a named Docker context to a dialable endpoint (incl. TLS).
func EndpointForContext(contextName string) (docker.Endpoint, error) {
	if contextName == "" {
		return docker.Endpoint{}, fmt.Errorf("empty context name")
	}
	cs, err := contextStore()
	if err != nil {
		return docker.Endpoint{}, err
	}
	meta, err := cs.GetMetadata(contextName)
	if err != nil {
		return docker.Endpoint{}, err
	}
	epMeta, err := docker.EndpointFromContext(meta)
	if err != nil {
		return docker.Endpoint{}, err
	}
	return docker.WithTLSData(cs, contextName, epMeta)
}

// ResolveDialTarget picks the Engine endpoint host and client options.
// selectedContext empty → legacy ResolveHost (DOCKER_HOST or local socket); DOCKER_HOST wins.
// non-empty → named Docker context from the local context store.
func ResolveDialTarget(selectedContext string) (host string, opts []client.Opt, err error) {
	if selectedContext == "" {
		host = ResolveHost()
		opts = []client.Opt{
			client.FromEnv,
			client.WithHost(host),
			client.WithAPIVersionNegotiation(),
		}
		return host, opts, nil
	}
	ep, err := EndpointForContext(selectedContext)
	if err != nil {
		return "", nil, err
	}
	opts, err = ep.ClientOpts()
	if err != nil {
		return "", nil, err
	}
	return ep.Host, opts, nil
}

// NewClient builds a Docker Engine API client for the given procpane context selection.
func NewClient(selectedContext string) (*client.Client, string, error) {
	host, opts, err := ResolveDialTarget(selectedContext)
	if err != nil {
		return nil, "", err
	}
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, host, fmt.Errorf("docker client: %w", err)
	}
	return cli, host, nil
}
