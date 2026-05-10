//go:build !server

package main

import "github.com/wailsapp/wails/v3/pkg/application"

const (
	defaultWindowWidth  = 1280
	defaultWindowHeight = 760
	minWindowWidth      = 1200
	minWindowHeight     = 680
)

func runtimeMode() string {
	return "desktop"
}

func configureApplicationOptions(_ *application.Options) {
}

func configureApplicationRuntime(app *application.App) {
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "sakiko",
		Width:     defaultWindowWidth,
		Height:    defaultWindowHeight,
		MinWidth:  minWindowWidth,
		MinHeight: minWindowHeight,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(11, 20, 34),
		URL:              "/",
	})
}
