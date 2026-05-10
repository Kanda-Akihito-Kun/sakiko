package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sakiko.local/sakiko-core/logx"

	"go.uber.org/zap"
)

func main() {
	logger, err := logx.Configure(logx.Config{
		Name:        "sakiko-wails",
		Level:       envOrDefault("SAKIKO_LOG_LEVEL", "info"),
		Development: envBool("SAKIKO_LOG_DEVELOPMENT"),
	})
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "configure logger failed: %v\n", err)
		os.Exit(1)
	}
	logger = installDesktopNotificationLogging(logger)
	logx.ReplaceGlobals(logger)
	defer func() {
		_ = logx.Sync()
	}()

	logger.Info("starting wails application", zap.String("runtime_mode", runtimeMode()))
	app := newApplication()

	if err := app.Run(); err != nil {
		logger.Error("wails application exited with error", zap.Error(err))
		_ = logx.Sync()
		os.Exit(1)
	}
	logger.Info("wails application stopped", zap.String("runtime_mode", runtimeMode()))
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

func envIntOrDefault(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
