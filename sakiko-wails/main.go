package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sakiko.local/sakiko-core/logx"

	"github.com/wailsapp/wails/v3/pkg/application"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	const (
		defaultWindowWidth  = 1120
		defaultWindowHeight = 760
		minWindowWidth      = 1024
		minWindowHeight     = 680
	)

	logger, err := logx.Configure(logx.Config{
		Name:        "sakiko-wails",
		Level:       envOrDefault("SAKIKO_LOG_LEVEL", "info"),
		Development: envBool("SAKIKO_LOG_DEVELOPMENT"),
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "configure logger failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = logx.Sync()
	}()

	logger.Info("starting wails application")
	app := application.New(application.Options{
		Name:        "sakiko",
		Description: "Desktop client powered by sakiko-core",
		Services: []application.Service{
			application.NewService(&SakikoService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

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

	if err := app.Run(); err != nil {
		logger.Error("wails application exited with error", zap.Error(err))
		_ = logx.Sync()
		os.Exit(1)
	}
	logger.Info("wails application stopped")
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envBool(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false
	}
	parsed, err := strconv.ParseBool(value)
	return err == nil && parsed
}
