package logx

import (
	"strings"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Name          string
	Level         string
	Development   bool
	Encoding      string
	DisableCaller bool
}

var global atomic.Pointer[zap.Logger]

func init() {
	global.Store(zap.NewNop())
}

func New(cfg Config) (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if raw := strings.TrimSpace(cfg.Level); raw != "" {
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			return nil, err
		}
	}

	encoding := strings.TrimSpace(cfg.Encoding)
	if encoding == "" {
		encoding = "console"
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logger, err := zap.Config{
		Level:             level,
		Development:       cfg.Development,
		Encoding:          encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableCaller:     cfg.DisableCaller,
		DisableStacktrace: !cfg.Development,
	}.Build()
	if err != nil {
		return nil, err
	}

	if cfg.Name != "" {
		logger = logger.Named(cfg.Name)
	}

	return logger, nil
}

func Configure(cfg Config) (*zap.Logger, error) {
	logger, err := New(cfg)
	if err != nil {
		return nil, err
	}
	ReplaceGlobals(logger)
	return logger, nil
}

func ReplaceGlobals(logger *zap.Logger) {
	if logger == nil {
		logger = zap.NewNop()
	}
	global.Store(logger)
	_ = zap.ReplaceGlobals(logger)
}

func L() *zap.Logger {
	logger := global.Load()
	if logger == nil {
		return zap.NewNop()
	}
	return logger
}

func Named(name string) *zap.Logger {
	if strings.TrimSpace(name) == "" {
		return L()
	}
	return L().Named(name)
}

func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

func Sync() error {
	return L().Sync()
}
