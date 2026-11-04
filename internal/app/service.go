package app

import (
	"context"
	"strconv"
	"strings"

	"github.com/zx06/procpane/internal/enrich"
	"github.com/zx06/procpane/internal/kill"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/source/local"
)

// Service is shared by CLI and GUI.
type Service struct {
	Source *local.Source
}

// NewService returns a service with default local source.
func NewService() *Service {
	return &Service{Source: local.New()}
}

// List returns enriched listeners, optionally filtered by proto string.
func (s *Service) List(ctx context.Context, proto string) ([]model.Listener, error) {
	src := *s.Source
	src.ProtoFilter = proto
	return src.ListListeners(ctx)
}

// Find matches port from user input (:8080, 8080, host:8080).
func (s *Service) Find(ctx context.Context, spec string, proto string) ([]model.Listener, error) {
	port, err := parsePortSpec(spec)
	if err != nil {
		return nil, err
	}
	all, err := s.List(ctx, proto)
	if err != nil {
		return nil, err
	}
	var matched []model.Listener
	for _, l := range all {
		if l.Port == port {
			matched = append(matched, l)
		}
	}
	return matched, nil
}

func parsePortSpec(spec string) (uint32, error) {
	s := strings.TrimSpace(spec)
	s = strings.TrimPrefix(s, ":")
	if idx := strings.LastIndex(s, ":"); idx >= 0 {
		s = s[idx+1:]
	}
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

// Show returns process detail for pid.
func (s *Service) Show(_ context.Context, pid int32) model.ProcessInfo {
	return enrich.EnrichProcess(pid)
}

// KillProcess sends signal after validation.
func (s *Service) KillProcess(pid int32, signal string) model.KillResult {
	return kill.Send(pid, signal)
}
