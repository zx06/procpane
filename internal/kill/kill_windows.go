//go:build windows

package kill

import (
	"fmt"
	"os"
	"syscall"

	"github.com/zx06/procpane/internal/enrich"
	"github.com/zx06/procpane/internal/model"
)

// SignalName on Windows maps TERM/KILL to process termination.
func SignalName(name string) (string, error) {
	switch name {
	case "TERM", "term", "SIGTERM", "KILL", "kill", "SIGKILL":
		return name, nil
	default:
		return "", fmt.Errorf("unknown signal %q (use TERM or KILL)", name)
	}
}

// Send terminates the process via TerminateProcess semantics (syscall).
func Send(pid int32, signal string) model.KillResult {
	res := model.KillResult{PID: pid, Signal: signal}
	if err := ValidateTargetPID(pid); err != nil {
		res.Error = err.Error()
		return res
	}
	info := enrich.EnrichProcess(pid)
	res.Name = info.Name
	if _, err := SignalName(signal); err != nil {
		res.Error = err.Error()
		return res
	}
	h, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer syscall.CloseHandle(h)
	if err := syscall.TerminateProcess(h, 1); err != nil {
		res.Error = err.Error()
		return res
	}
	res.Success = true
	return res
}

// Avoid unused import on windows build when os is only used indirectly.
var _ = os.Getpid
