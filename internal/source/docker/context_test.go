package docker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/context/docker"
	"github.com/docker/cli/cli/context/store"
)

func TestListContextsFromStore(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)

	cfg := command.DefaultContextStoreConfig()
	raw := store.New(filepath.Join(dir, "contexts"), cfg)
	err := raw.CreateOrUpdate(store.Metadata{
		Name: "remote",
		Metadata: map[string]any{
			"Description": "fixture remote engine",
		},
		Endpoints: map[string]any{
			docker.DockerEndpoint: docker.EndpointMeta{
				Host:          "tcp://127.0.0.1:2375",
				SkipTLSVerify: true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	list, err := ListContexts("remote")
	if err != nil {
		t.Fatal(err)
	}
	var foundDefault, foundRemote bool
	for _, c := range list {
		if c.Name == command.DefaultContextName {
			foundDefault = true
		}
		if c.Name == "remote" {
			foundRemote = true
			if c.Host != "tcp://127.0.0.1:2375" {
				t.Fatalf("host = %q", c.Host)
			}
			if c.Description != "fixture remote engine" {
				t.Fatalf("description = %q", c.Description)
			}
			if !c.Selected {
				t.Fatal("expected selected")
			}
		}
	}
	if !foundDefault || !foundRemote {
		t.Fatalf("contexts: %+v", list)
	}
}

func TestResolveHostEmptySelection(t *testing.T) {
	t.Setenv("DOCKER_HOST", "tcp://example:2375")
	host, _, err := ResolveDialTarget("")
	if err != nil {
		t.Fatal(err)
	}
	if host != "tcp://example:2375" {
		t.Fatalf("host = %q", host)
	}
}

func TestEndpointForContextDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DOCKER_CONFIG", dir)
	t.Setenv("DOCKER_HOST", "unix:///tmp/test.sock")

	ep, err := EndpointForContext(command.DefaultContextName)
	if err != nil {
		t.Fatal(err)
	}
	if ep.Host != "unix:///tmp/test.sock" {
		t.Fatalf("host = %q", ep.Host)
	}
}

func TestMain(m *testing.M) {
	// Avoid reading real ~/.docker in parallel tests.
	if os.Getenv("DOCKER_CONFIG") == "" {
		_ = os.Setenv("DOCKER_CONFIG", filepath.Join(os.TempDir(), "procpane-docker-test-unused"))
	}
	os.Exit(m.Run())
}
