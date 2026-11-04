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
)

func main() {
	svc := gui.NewService()

	app := application.New(application.Options{
		Name:        "procpane",
		Description: "Local LISTEN port panel",
		Icon:        procpaneicons.App256,
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

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
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
	menu.Add("刷新端口列表").OnClick(func(_ *application.Context) {
		app.Event.Emit("procpane:refresh", nil)
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
