package grpckit

import (
	"context"
	"time"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClientConnConfig struct {
	Timeout    time.Duration `long:"timeout" env:"TIMEOUT" default:"30s"`
	ServerAddr string        `long:"server_addr" env:"SERVER_ADDR" required:"true"`
}

type GrpcClientConn struct {
	*grpc.ClientConn

	closeFunc func()
}

func (c *GrpcClientConn) Close() error {
	if c.closeFunc != nil {
		c.closeFunc()
	}

	return c.ClientConn.Close()
}

func NewGrpcClientConn(ctx context.Context, conf *GrpcClientConnConfig) *GrpcClientConn {
	logger := logkit.FromContext(ctx).With(
		zap.String("server_addr", conf.ServerAddr),
	)

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, conf.Timeout)

	conn, err := grpc.DialContext(ctx, conf.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed to connect to gRPC server", zap.Error(err))
	}

	logger.Info("connect to gRPC server successfully")

	closeFunc := func() {
		cancel()
	}

	return &GrpcClientConn{
		ClientConn: conn,
		closeFunc:  closeFunc,
	}
}
