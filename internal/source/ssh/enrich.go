package ssh

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gossh "golang.org/x/crypto/ssh"
	"github.com/zx06/procpane/internal/model"
)

const enrichMarker = "__procpane_pid__:"

func pidsForCommEnrich(listeners []model.Listener) []int32 {
	var pids []int32
	seen := make(map[int32]struct{})
	for i := range listeners {
		if listeners[i].PID <= 0 {
			continue
		}
		if _, ok := seen[listeners[i].PID]; ok {
			continue
		}
		seen[listeners[i].PID] = struct{}{}
		pids = append(pids, listeners[i].PID)
	}
	return pids
}

func applyCommNames(listeners []model.Listener, names map[int32]string) {
	for i := range listeners {
		if n, ok := names[listeners[i].PID]; ok && n != "" {
			listeners[i].Name = n
		}
	}
}

// enrichRemote fills process names via batch ps comm. Scanner names are provisional
// (BusyBox netstat truncates program fields); ps comm overwrites when available.
func enrichRemote(ctx context.Context, client *gossh.Client, _ string, listeners []model.Listener) {
	pids := pidsForCommEnrich(listeners)
	if len(pids) == 0 {
		return
	}
	names := batchRemoteComm(ctx, client, pids)
	applyCommNames(listeners, names)
}

func batchRemoteComm(ctx context.Context, client *gossh.Client, pids []int32) map[int32]string {
	out := make(map[int32]string, len(pids))
	var b strings.Builder
	for _, pid := range pids {
		fmt.Fprintf(&b, "echo '%s%d'; ps -p %d -o comm= 2>/dev/null || true; echo\n", enrichMarker, pid, pid)
	}
	raw, err := runRemoteEnrich(ctx, client, b.String())
	if err != nil {
		return out
	}
	return parseBatchCommOutput(raw)
}

func parseBatchCommOutput(raw string) map[int32]string {
	out := make(map[int32]string)
	var cur int32
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, enrichMarker) {
			n, _ := strconv.ParseInt(strings.TrimPrefix(line, enrichMarker), 10, 32)
			cur = int32(n)
			continue
		}
		if cur > 0 && line != "" {
			out[cur] = line
			cur = 0
		}
	}
	return out
}
