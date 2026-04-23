package main

import (
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestBuildDesktopNotificationIncludesErrorField(t *testing.T) {
	entry := zapcore.Entry{
		Level:      zap.WarnLevel,
		Message:    "refresh profile failed",
		LoggerName: "sakiko-wails.service",
		Time:       time.Date(2026, 4, 14, 8, 30, 0, 0, time.UTC),
	}

	notification, ok := buildDesktopNotification(entry, []zap.Field{zap.Error(assertErr("network timeout"))})
	if !ok {
		t.Fatalf("buildDesktopNotification() ok = false")
	}
	if notification.Level != "warning" {
		t.Fatalf("notification.Level = %q", notification.Level)
	}
	if notification.Source != "service" {
		t.Fatalf("notification.Source = %q", notification.Source)
	}
	if notification.Message != "refresh profile failed: network timeout" {
		t.Fatalf("notification.Message = %q", notification.Message)
	}
}

func TestFormatNotificationSource(t *testing.T) {
	tests := map[string]string{
		"":                    "backend",
		"sakiko-wails":        "backend",
		"sakiko-wails.kernel": "kernel",
		"sakiko-core.service": "service",
	}

	for input, want := range tests {
		if got := formatNotificationSource(input); got != want {
			t.Fatalf("formatNotificationSource(%q) = %q, want %q", input, got, want)
		}
	}
}

type assertErr string

func (e assertErr) Error() string {
	return string(e)
}
