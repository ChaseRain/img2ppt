package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func New(level, format string) (*Logger, error) {
	var cfg zap.Config

	if format == "console" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	zapLogger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger.Sugar()}, nil
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Infow(msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Debugw(msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (l *Logger) With(keysAndValues ...interface{}) *Logger {
	return &Logger{l.SugaredLogger.With(keysAndValues...)}
}

func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}
