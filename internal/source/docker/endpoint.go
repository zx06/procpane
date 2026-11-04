package docker

import (
	"os"
	"strings"
)

// ResolveHost returns the Docker API endpoint (from DOCKER_HOST or default unix socket).
func ResolveHost() string {
	if h := strings.TrimSpace(os.Getenv("DOCKER_HOST")); h != "" {
		return h
	}
	return "unix:///var/run/docker.sock"
}
