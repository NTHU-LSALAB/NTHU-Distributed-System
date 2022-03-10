package kafkakit

import (
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/justin0u0/protoc-gen-grpc-sarama/pkg/saramakit"
	"go.uber.org/zap"
)

// Logger implements saramakit.Logger
type Logger struct {
	*logkit.Logger
}

var _ saramakit.Logger = (*Logger)(nil)

func (l *Logger) With(key, value string) saramakit.Logger {
	return &Logger{
		Logger: l.Logger.With(zap.String(key, value)),
	}
}

func (l *Logger) Error(msg string, err error) {
	l.Logger.Error(msg, zap.Error(err))
}

func NewLogger(logger *logkit.Logger) *Logger {
	return &Logger{
		Logger: logger,
	}
}
