package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const desktopNotificationEventName = "sakiko:desktop-notification"

type DesktopNotification struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"`
	Timestamp string `json:"timestamp"`
}

var desktopNotificationApp atomic.Pointer[application.App]

func installDesktopNotificationLogging(logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return nil
	}

	return logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &desktopNotificationCore{next: core}
	}))
}

func setDesktopNotificationApp(app *application.App) {
	desktopNotificationApp.Store(app)
}

func clearDesktopNotificationApp() {
	desktopNotificationApp.Store(nil)
}

func emitDesktopNotification(notification DesktopNotification) bool {
	app := desktopNotificationApp.Load()
	if app == nil {
		return false
	}
	return app.Event.Emit(desktopNotificationEventName, notification)
}

type desktopNotificationCore struct {
	next          zapcore.Core
	contextFields []zap.Field
}

func (c *desktopNotificationCore) Enabled(level zapcore.Level) bool {
	if c == nil || c.next == nil {
		return false
	}
	return c.next.Enabled(level)
}

func (c *desktopNotificationCore) With(fields []zap.Field) zapcore.Core {
	if c == nil {
		return nil
	}

	combined := make([]zap.Field, 0, len(c.contextFields)+len(fields))
	combined = append(combined, c.contextFields...)
	combined = append(combined, fields...)

	return &desktopNotificationCore{
		next:          c.next.With(fields),
		contextFields: combined,
	}
}

func (c *desktopNotificationCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *desktopNotificationCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	err := c.next.Write(entry, fields)

	if entry.Level < zap.WarnLevel {
		return err
	}

	combined := make([]zap.Field, 0, len(c.contextFields)+len(fields))
	combined = append(combined, c.contextFields...)
	combined = append(combined, fields...)

	notification, ok := buildDesktopNotification(entry, combined)
	if ok {
		emitDesktopNotification(notification)
	}
	return err
}

func (c *desktopNotificationCore) Sync() error {
	if c == nil || c.next == nil {
		return nil
	}
	return c.next.Sync()
}

func buildDesktopNotification(entry zapcore.Entry, fields []zap.Field) (DesktopNotification, bool) {
	level, ok := desktopNotificationLevel(entry.Level)
	if !ok {
		return DesktopNotification{}, false
	}

	fieldMap := encodeNotificationFields(fields)
	message := strings.TrimSpace(entry.Message)
	if errMessage := strings.TrimSpace(stringField(fieldMap, "error")); errMessage != "" && !strings.Contains(message, errMessage) {
		if message == "" {
			message = errMessage
		} else {
			message = fmt.Sprintf("%s: %s", message, errMessage)
		}
	}
	if message == "" {
		return DesktopNotification{}, false
	}

	timestamp := entry.Time
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	return DesktopNotification{
		Level:     level,
		Message:   message,
		Source:    formatNotificationSource(entry.LoggerName),
		Timestamp: timestamp.Format(time.RFC3339),
	}, true
}

func desktopNotificationLevel(level zapcore.Level) (string, bool) {
	switch {
	case level >= zap.ErrorLevel:
		return "error", true
	case level >= zap.WarnLevel:
		return "warning", true
	default:
		return "", false
	}
}

func formatNotificationSource(loggerName string) string {
	source := strings.TrimSpace(loggerName)
	if source == "sakiko-wails" || source == "sakiko-core" {
		return "backend"
	}
	source = strings.TrimPrefix(source, "sakiko-wails.")
	source = strings.TrimPrefix(source, "sakiko-core.")
	source = strings.Trim(source, ".")
	if source == "" {
		return "backend"
	}
	return source
}

func encodeNotificationFields(fields []zap.Field) map[string]any {
	if len(fields) == 0 {
		return nil
	}

	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(encoder)
	}
	return encoder.Fields
}

func stringField(fields map[string]any, key string) string {
	if len(fields) == 0 {
		return ""
	}

	value, exists := fields[key]
	if !exists || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case error:
		return typed.Error()
	default:
		return fmt.Sprintf("%v", typed)
	}
}
