package logkit

import (
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	"go.uber.org/zap"
)

// SaramaLogger implements saramakit.Logger
type SaramaLogger struct {
	*Logger
}

var _ saramakit.Logger = (*SaramaLogger)(nil)

func (l *SaramaLogger) With(key, value string) saramakit.Logger {
	return &SaramaLogger{
		Logger: l.Logger.With(zap.String(key, value)),
	}
}

func (l *SaramaLogger) Error(msg string, err error) {
	l.Logger.Error(msg, zap.Error(err))
}

func NewSaramaLogger(logger *Logger) *SaramaLogger {
	return &SaramaLogger{
		Logger: logger,
	}
}
