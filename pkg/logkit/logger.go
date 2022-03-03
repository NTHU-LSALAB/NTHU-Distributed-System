package logkit

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerLevel zapcore.Level

func (l LoggerLevel) MarshalFlag() (string, error) {
	return zapcore.Level(l).String(), nil
}

func (l *LoggerLevel) UnmarshalFlag(value string) error {
	zv := (*zapcore.Level)(l)
	return zv.UnmarshalText([]byte(value))
}

type LoggerConfig struct {
	Level       LoggerLevel `long:"level" description:"set log level" default:"info" env:"LEVEL"`
	Development bool        `long:"development" description:"enable development mode" env:"DEVELOPMENT"`
}

type Logger struct {
	*zap.Logger
}

func NewNopLogger() *Logger {
	return &Logger{Logger: zap.NewNop()}
}

func NewLogger(conf *LoggerConfig) *Logger {
	var config zap.Config

	if conf.Development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.Level(conf.Level))

	logger, err := config.Build()
	if err != nil {
		log.Fatal("failed to build logger")
	}

	return &Logger{Logger: logger}
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return WithContext(ctx, l)
}

func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
	}
}
