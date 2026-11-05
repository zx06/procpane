package main

import (
	"log"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	procpaneicons "github.com/zx06/procpane/assets/icons"
	procpanefrontend "github.com/zx06/procpane/frontend"
	"github.com/zx06/procpane/internal/gui"
)

// Default tray popover size (compact menu-bar utility, not a full app window).
const (
	windowWidth  = 440
	windowHeight = 520
	settingsW    = 420
	settingsH    = 480
)

func main() {
	svc := gui.NewService()

	app := application.New(application.Options{
		Name:        "procpane",
		Description: "Local LISTEN port panel",
		Icon: procpaneicons.App256,
		Server: application.ServerOptions{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(procpanefrontend.Assets),
		},
		Mac: application.MacOptions{
			ActivationPolicy: application.ActivationPolicyAccessory,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
	})

	tray := app.SystemTray.New()
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "procpane",
		Width:            windowWidth,
		Height:           windowHeight,
		Name:             "main",
		URL:              "/",
		Hidden:           true,
		Frameless:        true,
		AlwaysOnTop:      true,
		HideOnEscape:     true,
		HideOnFocusLost:  true,
		BackgroundColour: application.NewRGB(242, 244, 248),
		Mac: application.MacWindow{
			Appearance: application.DefaultAppearance,
		},
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: true,
			Theme:           application.SystemDefault,
		},
	})

	settingsWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "procpane — 设置",
		Width:            settingsW,
		Height:           settingsH,
		Name:             "settings",
		URL:              "/settings.html",
		Hidden:           true,
		Frameless:        false,
		AlwaysOnTop:      false,
		HideOnEscape:     true,
		BackgroundColour: application.NewRGB(242, 244, 248),
		Mac: application.MacWindow{
			Appearance: application.DefaultAppearance,
		},
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: false,
			Theme:           application.SystemDefault,
		},
	})

	svc.BindSettingsWindow(func() {
		settingsWindow.Show()
	})
	svc.BindEventEmitter(func(name string, data ...any) {
		app.Event.Emit(name, data...)
	})
	if err := svc.SyncAutostart(); err != nil {
		log.Printf("autostart sync: %v", err)
	}

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		settingsWindow.Hide()
		e.Cancel()
	})

	switch runtime.GOOS {
	case "darwin":
		tray.SetTemplateIcon(procpaneicons.TrayMacTemplate)
	case "windows":
		tray.SetIcon(procpaneicons.WindowsICO)
	default:
		tray.SetIcon(procpaneicons.TrayDark)
	}

	menu := app.NewMenu()
	menu.Add("打开面板").OnClick(func(_ *application.Context) {
		window.Show()
	})
	menu.Add("刷新当前源").OnClick(func(_ *application.Context) {
		svc.RefreshCurrentSource()
	})
	menu.Add("设置…").OnClick(func(_ *application.Context) {
		settingsWindow.Show()
	})
	menu.AddSeparator()
	menu.Add("退出 procpane").OnClick(func(_ *application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)
	tray.SetTooltip("procpane")
	tray.AttachWindow(window).WindowOffset(4)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
