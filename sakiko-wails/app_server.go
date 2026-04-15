//go:build server

package main

import (
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	defaultServerHost = "localhost"
	defaultServerPort = 8080
)

func runtimeMode() string {
	return "server"
}

func configureApplicationOptions(options *application.Options) {
	options.Server = application.ServerOptions{
		Host:            envOrDefault("SAKIKO_SERVER_HOST", defaultServerHost),
		Port:            envIntOrDefault("SAKIKO_SERVER_PORT", defaultServerPort),
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

func configureApplicationRuntime(_ *application.App) {
}
