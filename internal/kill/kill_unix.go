//go:build !windows

package kill

import (
	"fmt"
	"os"
	"syscall"

	"github.com/zx06/procpane/internal/enrich"
	"github.com/zx06/procpane/internal/model"
)

// SignalName maps CLI signal names to os signals.
func SignalName(name string) (os.Signal, error) {
	switch name {
	case "TERM", "term", "SIGTERM":
		return syscall.SIGTERM, nil
	case "KILL", "kill", "SIGKILL":
		return syscall.SIGKILL, nil
	default:
		return nil, fmt.Errorf("unknown signal %q (use TERM or KILL)", name)
	}
}

// Send sends a signal to pid after safety checks.
func Send(pid int32, signal string) model.KillResult {
	res := model.KillResult{PID: pid, Signal: signal}
	if err := ValidateTargetPID(pid); err != nil {
		res.Error = err.Error()
		return res
	}
	info := enrich.EnrichProcess(pid)
	res.Name = info.Name
	sig, err := SignalName(signal)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	p, err := os.FindProcess(int(pid))
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if err := p.Signal(sig); err != nil {
		res.Error = err.Error()
		return res
	}
	res.Success = true
	return res
}
