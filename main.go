package main

import (
	"embed"
	"runtime"

	"lanfiletransfertool/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	application := app.NewApp()
	var err error
	if runtime.GOOS == "windows" {
		err = wails.Run(&options.App{
			Title:  "LAN File Transfer Tool",
			Width:  400,
			Height: 800,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			OnStartup:     application.Startup,
			OnDomReady:    application.DomReady,
			OnBeforeClose: application.BeforeClose,
			OnShutdown:    application.Shutdown,
			Bind: []interface{}{
				application,
			},
			Windows: &windows.Options{
				WebviewIsTransparent: false,
				WindowIsTranslucent:  false,
				DisableWindowIcon:    false,
			},
		})
	} else if runtime.GOOS == "darwin" {
		err = wails.Run(&options.App{
			Title:  "LAN File Transfer Tool",
			Width:  400,
			Height: 800,
			AssetServer: &assetserver.Options{
				Assets: assets,
			},
			OnStartup:     application.Startup,
			OnDomReady:    application.DomReady,
			OnBeforeClose: application.BeforeClose,
			OnShutdown:    application.Shutdown,
			Bind: []interface{}{
				application,
			},
			Windows: &windows.Options{
				WebviewIsTransparent: false,
				WindowIsTranslucent:  false,
				DisableWindowIcon:    false,
			},
			Mac: &mac.Options{
				TitleBar: &mac.TitleBar{
					TitlebarAppearsTransparent: false,
					HideTitle:                  false,
					HideTitleBar:               false,
					FullSizeContent:            false,
					UseToolbar:                 false,
					HideToolbarSeparator:       true,
				},
				Appearance:           mac.NSAppearanceNameDarkAqua,
				WebviewIsTransparent: false,
				WindowIsTranslucent:  false,
			},
		})
	}

	if err != nil {
		println("Error:", err.Error())
	}
}
