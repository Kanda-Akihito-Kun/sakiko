package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func newApplication() *application.App {
	options := application.Options{
		Name:        "Sakiko",
		Description: "Desktop proxy benchmark and speed test tool",
		Services: []application.Service{
			application.NewService(&SakikoService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Flags: map[string]any{
			"sakikoRuntimeMode": runtimeMode(),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	}

	configureApplicationOptions(&options)
	app := application.New(options)
	configureApplicationRuntime(app)
	return app
}
