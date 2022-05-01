package video

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/gateway"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/grpckit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/runkit"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newGatewayCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gateway",
		Short: "starts video gateway server",
		RunE:  runGateway,
	}
}

type GatewayArgs struct {
	HTTPAddr                     string `long:"http_addr" env:"HTTP_ADDR" default:":8080"`
	GRPCAddr                     string `long:"grpc_addr" env:"GRPC_ADDR" default:":8081"`
	grpckit.GrpcClientConnConfig `group:"grpc" namespace:"grpc" env-namespace:"GRPC"`
	runkit.GracefulConfig        `group:"graceful" namespace:"graceful" env-namespace:"GRACEFUL"`
	logkit.LoggerConfig          `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
}

func runGateway(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args GatewayArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("failed to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		_ = logger.Sync()
	}()

	ctx = logger.WithContext(ctx)

	logger.Info("listen to HTTP addr", zap.String("http_addr", args.HTTPAddr))
	lis, err := net.Listen("tcp", args.HTTPAddr)
	if err != nil {
		logger.Fatal("failed to listen HTTP addr", zap.Error(err))
	}
	defer func() {
		if err := lis.Close(); err != nil {
			logger.Fatal("failed to close HTTP listener", zap.Error(err))
		}
	}()

	conn := grpckit.NewGrpcClientConn(ctx, &args.GrpcClientConnConfig)
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Fatal("failed to close gRPC client connection", zap.Error(err))
		}
	}()

	return runkit.GracefulRun(serveHTTP(lis, conn.ClientConn, logger), &args.GracefulConfig)
}

func serveHTTP(lis net.Listener, conn *grpc.ClientConn, logger *logkit.Logger) runkit.GracefulRunFunc {
	mux := runtime.NewServeMux()

	// register additional routes
	handler := gateway.NewHandler(pb.NewVideoClient(conn), logger)
	if err := mux.HandlePath("POST", "/v1/videos", handler.HandleUploadVideo); err != nil {
		logger.Fatal("failed to register additional routes", zap.Error(err))
	}

	httpServer := &http.Server{
		Handler: mux,
	}

	return func(ctx context.Context) error {
		if err := pb.RegisterVideoHandler(ctx, mux, conn); err != nil {
			logger.Fatal("failed to register handler to HTTP server", zap.Error(err))
		}

		go func() {
			if err := httpServer.Serve(lis); err != nil {
				logger.Fatal("failed to run HTTP server", zap.Error(err))
			}
		}()

		<-ctx.Done()

		if err := httpServer.Shutdown(context.Background()); err != nil {
			logger.Fatal("failed to shutdown HTTP server", zap.Error(err))
		}

		return nil
	}
}
