package main

import (
	"log"
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
	procpanefrontend "github.com/zx06/procpane/frontend"
	"github.com/zx06/procpane/internal/gui"
)

func main() {
	svc := gui.NewService()

	app := application.New(application.Options{
		Name:        "procpane",
		Description: "Local LISTEN port panel",
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
		Title:           "procpane — 本地监听端口",
		Width:           920,
		Height:          620,
		Name:            "main",
		URL:             "/",
		Hidden:          false,
		HideOnEscape:    true,
		HideOnFocusLost: false,
		BackgroundColour: application.NewRGB(11, 13, 18),
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBarHiddenInset,
		},
		Windows: application.WindowsWindow{
			HiddenOnTaskbar: false,
		},
	})

	window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		window.Hide()
		e.Cancel()
	})

	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(icons.SystrayMacTemplate)
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
