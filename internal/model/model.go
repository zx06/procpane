package model

import "time"

// Unknown is shown when enrich cannot read a field.
const Unknown = "?"

// Protocol is tcp or udp.
type Protocol string

const (
	ProtoTCP Protocol = "tcp"
	ProtoUDP Protocol = "udp"
)

// Listener is a single LISTEN socket with enriched process metadata.
type Listener struct {
	ID       string   `json:"id"`
	Port     uint32   `json:"port"`
	Proto    Protocol `json:"proto"`
	PID      int32    `json:"pid"`
	Name     string   `json:"name"`
	CPU      float64  `json:"cpuPercent"`
	MemBytes uint64   `json:"memBytes"`
	MemMB    float64  `json:"memMb"`
	Cmdline  string   `json:"cmdline"`
	Cwd      string   `json:"cwd"`
	User     string   `json:"user"`
	Started  string   `json:"started"`
	Leftover bool     `json:"leftover,omitempty"`
	Chain    []ChainNode `json:"chain"`
}

// ChainNode is one step in the parent process chain (root → leaf).
type ChainNode struct {
	PID  int32  `json:"pid"`
	Name string `json:"name"`
	Note string `json:"note"`
}

// ProcessInfo is detailed metadata for a PID (show / enrich).
type ProcessInfo struct {
	PID      int32       `json:"pid"`
	Name     string      `json:"name"`
	Cmdline  string      `json:"cmdline"`
	Cwd      string      `json:"cwd"`
	User     string      `json:"user"`
	Started  string      `json:"started"`
	CPU      float64     `json:"cpuPercent"`
	MemBytes uint64      `json:"memBytes"`
	MemMB    float64     `json:"memMb"`
	Chain    []ChainNode `json:"chain"`
	Ports    []PortRef   `json:"ports,omitempty"`
}

// PortRef is a listening port owned by a process.
type PortRef struct {
	Port  uint32   `json:"port"`
	Proto Protocol `json:"proto"`
}

// KillRequest is the stable contract for kill operations.
type KillRequest struct {
	PID    int32  `json:"pid"`
	Signal string `json:"signal"` // TERM or KILL
}

// KillResult summarizes an attempted kill.
type KillResult struct {
	PID     int32  `json:"pid"`
	Name    string `json:"name"`
	Signal  string `json:"signal"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// ListResponse is the JSON envelope for list/find.
type ListResponse struct {
	GeneratedAt time.Time  `json:"generatedAt"`
	Listeners   []Listener `json:"listeners"`
}
