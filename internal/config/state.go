package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// State holds lightweight procpane preferences (SSH host enable flags).
type State struct {
	SSH SSHState `yaml:"ssh"`
}

// SSHState tracks which OpenSSH config Host aliases are enabled.
type SSHState struct {
	Enabled map[string]bool `yaml:"enabled"`
}

// Store loads and saves state under XDG config (e.g. ~/.config/procpane/state.yaml).
type Store struct {
	path string
	mu   sync.Mutex
}

// DefaultStore returns a Store using XDG_CONFIG_HOME or ~/.config/procpane/state.yaml.
func DefaultStore() (*Store, error) {
	dir, err := configDir()
	if err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "state.yaml")}, nil
}

func configDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "procpane"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "procpane"), nil
}

// Path returns the on-disk state file path.
func (s *Store) Path() string {
	return s.path
}

// Load reads state; missing file yields empty defaults.
func (s *Store) Load() (State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return State{SSH: SSHState{Enabled: map[string]bool{}}}, nil
		}
		return State{}, err
	}
	var st State
	if err := yaml.Unmarshal(data, &st); err != nil {
		return State{}, err
	}
	if st.SSH.Enabled == nil {
		st.SSH.Enabled = map[string]bool{}
	}
	return st, nil
}

// Save writes state, creating the config directory if needed.
func (s *Store) Save(st State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st.SSH.Enabled == nil {
		st.SSH.Enabled = map[string]bool{}
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// SetSSHEnabled toggles one host alias and persists.
func (s *Store) SetSSHEnabled(alias string, on bool) error {
	st, err := s.Load()
	if err != nil {
		return err
	}
	if on {
		st.SSH.Enabled[alias] = true
	} else {
		delete(st.SSH.Enabled, alias)
	}
	return s.Save(st)
}

// IsSSHEnabled reports whether alias is enabled (default false).
func (s *Store) IsSSHEnabled(alias string) (bool, error) {
	st, err := s.Load()
	if err != nil {
		return false, err
	}
	return st.SSH.Enabled[alias], nil
}

// ConfigSummary is a read-only view for CLI/GUI.
type ConfigSummary struct {
	StatePath     string            `json:"statePath"`
	SSHEnabled    map[string]bool   `json:"sshEnabled"`
	DockerHost    string            `json:"dockerHost"`
	DockerReachable bool            `json:"dockerReachable"`
	DockerError   string            `json:"dockerError,omitempty"`
}

// ShowSummary builds config + docker probe info.
func ShowSummary(store *Store, dockerHost string, dockerOK bool, dockerErr error) (ConfigSummary, error) {
	st, err := store.Load()
	if err != nil {
		return ConfigSummary{}, err
	}
	sum := ConfigSummary{
		StatePath:   store.Path(),
		SSHEnabled:  copyMap(st.SSH.Enabled),
		DockerHost:  dockerHost,
		DockerReachable: dockerOK,
	}
	if dockerErr != nil {
		sum.DockerError = dockerErr.Error()
	}
	return sum, nil
}

func copyMap(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// FormatEnabledHosts returns sorted enabled aliases.
func FormatEnabledHosts(enabled map[string]bool) []string {
	var out []string
	for k, v := range enabled {
		if v {
			out = append(out, k)
		}
	}
	if len(out) == 0 {
		return nil
	}
	// simple sort
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// ErrHostNotEnabled is returned when listing a disabled SSH host.
var ErrHostNotEnabled = fmt.Errorf("ssh host is not enabled; run: procpane config enable <host>")
