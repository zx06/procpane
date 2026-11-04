package gui

import (
	"context"
	"sync"
	"time"

	"github.com/zx06/procpane/internal/app"
	"github.com/zx06/procpane/internal/model"
)

// ProcpaneService exposes list/kill to the Wails frontend.
type ProcpaneService struct {
	svc *app.Service
	mu  sync.Mutex
}

// NewService returns a Wails-bindable service.
func NewService() *ProcpaneService {
	return &ProcpaneService{svc: app.NewService()}
}

// ListListeners returns enriched LISTEN rows for the GUI.
func (p *ProcpaneService) ListListeners(proto string) ([]model.Listener, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if proto == "" {
		proto = "all"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return p.svc.List(ctx, proto)
}

// KillProcess sends TERM or KILL after backend safety checks (GUI confirms in UI).
func (p *ProcpaneService) KillProcess(pid int32, signal string) (model.KillResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if signal == "" {
		signal = "TERM"
	}
	res := p.svc.KillProcess(pid, signal)
	if res.Error != "" && !res.Success {
		return res, nil
	}
	return res, nil
}

// Version returns the app version string for the UI badge.
func (p *ProcpaneService) Version() string {
	return "0.1.0-v1"
}
