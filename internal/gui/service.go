package gui

import (
	"context"
	"sync"
	"time"

	"github.com/zx06/procpane/internal/app"
	"github.com/zx06/procpane/internal/autostart"
	"github.com/zx06/procpane/internal/config"
	"github.com/zx06/procpane/internal/model"
	dockersrc "github.com/zx06/procpane/internal/source/docker"
	sshsrc "github.com/zx06/procpane/internal/source/ssh"
	"github.com/zx06/procpane/internal/version"
)

// ProcpaneService exposes list/kill/show and v0.2 host sources to the Wails frontend.
type ProcpaneService struct {
	svc   *app.Service
	store *config.Store

	localMu  sync.Mutex
	dockerMu sync.Mutex
	sshMu    sync.Mutex
	cfgMu    sync.Mutex

	activeSource string
	activeMu     sync.RWMutex

	emitFn         func(name string, data ...any)
	showSettingsFn func()
}

// SSHHostView is one OpenSSH alias for the GUI.
type SSHHostView struct {
	Alias    string `json:"alias"`
	Hostname string `json:"hostname"`
	Port     string `json:"port,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// DockerContextView is one Docker CLI context for Settings.
type DockerContextView struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Host          string `json:"host"`
	SkipTLSVerify bool   `json:"skipTLSVerify,omitempty"`
	Current       bool   `json:"current"`
	Selected      bool   `json:"selected"`
	Error         string `json:"error,omitempty"`
}

// DockerStatus is Engine endpoint info and context list for Settings / Docker tab.
type DockerStatus struct {
	Endpoint        string              `json:"endpoint"`
	Reachable       bool                `json:"reachable"`
	Error           string              `json:"error,omitempty"`
	SelectedContext string              `json:"selectedContext,omitempty"`
	CurrentContext  string              `json:"currentContext,omitempty"`
	Contexts        []DockerContextView `json:"contexts,omitempty"`
}

// SettingsView backs the independent settings window.
type SettingsView struct {
	StatePath          string        `json:"statePath"`
	RefreshIntervalSec int           `json:"refreshIntervalSec"`
	Autostart          bool          `json:"autostart"`
	Docker             DockerStatus  `json:"docker"`
	SSHHosts           []SSHHostView `json:"sshHosts"`
}

// NewService returns a Wails-bindable service.
func NewService() *ProcpaneService {
	store, err := config.DefaultStore()
	if err != nil {
		store = &config.Store{}
	}
	return &ProcpaneService{svc: app.NewService(), store: store, activeSource: "local"}
}

// BindSettingsWindow wires the independent settings window opener (set from cmd/procpane-app).
func (p *ProcpaneService) BindSettingsWindow(show func()) {
	p.showSettingsFn = show
}

// BindEventEmitter wires Wails application.Event.Emit (set from cmd/procpane-app).
func (p *ProcpaneService) BindEventEmitter(emit func(name string, data ...any)) {
	p.emitFn = emit
}

func (p *ProcpaneService) emit(name string, data ...any) {
	if p.emitFn != nil {
		p.emitFn(name, data...)
	}
}

func (p *ProcpaneService) emitSettingsChanged(payload map[string]any) {
	p.emit("procpane:settings-changed", payload)
}

// SetActiveSource records which tab is visible for tray “刷新当前源”.
func (p *ProcpaneService) SetActiveSource(tab string) {
	if tab == "" {
		tab = "local"
	}
	p.activeMu.Lock()
	p.activeSource = tab
	p.activeMu.Unlock()
}

// RefreshCurrentSource emits a refresh event for the active tab (tray menu).
func (p *ProcpaneService) RefreshCurrentSource() {
	p.activeMu.RLock()
	src := p.activeSource
	p.activeMu.RUnlock()
	if src == "" {
		src = "local"
	}
	p.emit("procpane:refresh", map[string]string{"source": src})
}

// OpenSettings shows the settings window (tray menu or SSH empty-state CTA).
func (p *ProcpaneService) OpenSettings() {
	if p.showSettingsFn != nil {
		p.showSettingsFn()
	}
}

// ListListeners returns enriched local LISTEN rows for the GUI.
func (p *ProcpaneService) ListListeners(proto string) ([]model.Listener, error) {
	p.localMu.Lock()
	defer p.localMu.Unlock()
	if proto == "" {
		proto = "all"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return p.svc.List(ctx, proto)
}

// ListDockerListeners returns published container ports (A1).
func (p *ProcpaneService) ListDockerListeners() ([]model.Listener, error) {
	p.dockerMu.Lock()
	defer p.dockerMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	selected, _ := p.store.DockerContext()
	src := dockersrc.NewWithContext(selected)
	return src.ListListeners(ctx)
}

// DockerStatus probes the configured Engine endpoint.
func (p *ProcpaneService) DockerStatus() DockerStatus {
	p.dockerMu.Lock()
	defer p.dockerMu.Unlock()
	return p.dockerStatusLocked()
}

// ListSSHHosts returns ssh_config aliases with enable flags.
func (p *ProcpaneService) ListSSHHosts() ([]SSHHostView, error) {
	p.cfgMu.Lock()
	defer p.cfgMu.Unlock()
	st, err := p.store.Load()
	if err != nil {
		return nil, err
	}
	hosts, err := sshsrc.ListHosts()
	if err != nil {
		return nil, err
	}
	out := make([]SSHHostView, 0, len(hosts))
	for _, h := range hosts {
		out = append(out, SSHHostView{
			Alias:    h.Alias,
			Hostname: h.Hostname,
			Port:     h.Port,
			Enabled:  st.SSH.Enabled[h.Alias],
		})
	}
	return out, nil
}

// ListSSHListeners lists remote LISTEN for an enabled host alias.
func (p *ProcpaneService) ListSSHListeners(alias string) ([]model.Listener, error) {
	p.sshMu.Lock()
	defer p.sshMu.Unlock()
	on, err := p.store.IsSSHEnabled(alias)
	if err != nil {
		return nil, err
	}
	if !on {
		return nil, config.ErrHostNotEnabled
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	return sshsrc.New(alias).ListListeners(ctx)
}

// GetSettings returns data for the settings window.
func (p *ProcpaneService) GetSettings() (SettingsView, error) {
	p.cfgMu.Lock()
	defer p.cfgMu.Unlock()
	st, err := p.store.Load()
	if err != nil {
		return SettingsView{}, err
	}
	sshHosts, err := sshsrc.ListHosts()
	if err != nil {
		return SettingsView{}, err
	}
	hostViews := make([]SSHHostView, 0, len(sshHosts))
	for _, h := range sshHosts {
		hostViews = append(hostViews, SSHHostView{
			Alias:    h.Alias,
			Hostname: h.Hostname,
			Port:     h.Port,
			Enabled:  st.SSH.Enabled[h.Alias],
		})
	}
	refreshSec, _ := p.store.RefreshIntervalSec()
	autostartOn, _ := p.store.Autostart()
	return SettingsView{
		StatePath:          p.store.Path(),
		RefreshIntervalSec: refreshSec,
		Autostart:          autostartOn,
		Docker:             p.dockerStatusLocked(),
		SSHHosts:           hostViews,
	}, nil
}

func (p *ProcpaneService) dockerStatusLocked() DockerStatus {
	selected, _ := p.store.DockerContext()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	host, err := dockersrc.PingContext(ctx, selected)
	st := DockerStatus{
		Endpoint:        host,
		Reachable:       err == nil,
		SelectedContext: selected,
		CurrentContext:  dockersrc.DockerCLICurrentContext(),
	}
	if err != nil {
		st.Error = err.Error()
	}
	ctxs, listErr := dockersrc.ListContexts(selected)
	if listErr != nil {
		if st.Error == "" {
			st.Error = listErr.Error()
		}
	} else {
		st.Contexts = make([]DockerContextView, len(ctxs))
		for i, c := range ctxs {
			st.Contexts[i] = DockerContextView(c)
		}
	}
	return st
}

// SetDockerContext persists the Docker context name (empty clears to DOCKER_HOST behavior).
func (p *ProcpaneService) SetDockerContext(name string) error {
	p.cfgMu.Lock()
	err := p.store.SetDockerContext(name)
	p.cfgMu.Unlock()
	if err == nil {
		p.emitSettingsChanged(map[string]any{"dockerContext": true})
	}
	return err
}

// SetSSHEnabled toggles one SSH host alias in persisted state.
func (p *ProcpaneService) SetSSHEnabled(alias string, enabled bool) error {
	p.cfgMu.Lock()
	err := p.store.SetSSHEnabled(alias, enabled)
	p.cfgMu.Unlock()
	if err == nil {
		p.emitSettingsChanged(map[string]any{"sshHosts": true})
	}
	return err
}

// SetRefreshIntervalSec persists the GUI auto-refresh interval for all tabs.
func (p *ProcpaneService) SetRefreshIntervalSec(sec int) error {
	p.cfgMu.Lock()
	err := p.store.SetRefreshIntervalSec(sec)
	p.cfgMu.Unlock()
	if err == nil {
		p.emitSettingsChanged(map[string]any{"refreshIntervalSec": sec})
	}
	return err
}

// SyncAutostart applies persisted autostart to the OS (call once at app launch).
func (p *ProcpaneService) SyncAutostart() error {
	p.cfgMu.Lock()
	on, err := p.store.Autostart()
	p.cfgMu.Unlock()
	if err != nil {
		return err
	}
	return autostart.SetEnabled(on)
}

// SetAutostart toggles login autostart and updates OS registration immediately.
func (p *ProcpaneService) SetAutostart(on bool) error {
	p.cfgMu.Lock()
	prev, err := p.store.Autostart()
	if err != nil {
		p.cfgMu.Unlock()
		return err
	}
	if err := p.store.SetAutostart(on); err != nil {
		p.cfgMu.Unlock()
		return err
	}
	p.cfgMu.Unlock()

	if err := autostart.SetEnabled(on); err != nil {
		p.cfgMu.Lock()
		_ = p.store.SetAutostart(prev)
		p.cfgMu.Unlock()
		return err
	}
	p.emitSettingsChanged(map[string]any{"autostart": on})
	return nil
}

// KillProcess sends TERM or KILL after backend safety checks (local only).
func (p *ProcpaneService) KillProcess(pid int32, signal string) (model.KillResult, error) {
	p.localMu.Lock()
	defer p.localMu.Unlock()
	if signal == "" {
		signal = "TERM"
	}
	res := p.svc.KillProcess(pid, signal)
	if res.Error != "" && !res.Success {
		return res, nil
	}
	return res, nil
}

// ShowProcess returns enriched metadata for a PID (parent chain drill-down).
func (p *ProcpaneService) ShowProcess(pid int32) (model.ProcessInfo, error) {
	p.localMu.Lock()
	defer p.localMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return p.svc.Show(ctx, pid), nil
}

// Version returns the app version string for the UI badge.
func (p *ProcpaneService) Version() string {
	return version.Version
}
