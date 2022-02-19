package video

import (
	"context"
	"log"
	"net"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/service"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/mongokit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/runkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/storagekit"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newAPICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "starts Video API server",
		RunE:  runAPI,
	}
}

type APIArgs struct {
	GRPCAddr               string `long:"grpc_addr" env:"GRPC_ADDR" default:":8081"`
	runkit.GracefulConfig  `group:"graceful" namespace:"graceful" env-namespace:"GRACEFUL"`
	logkit.LoggerConfig    `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	mongokit.MongoConfig   `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`
	storagekit.MinIOConfig `group:"minio" namespace:"minio" env-namespace:"MINIO"`
}

func runAPI(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args APIArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("failed to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatal("failed to sync logger", err.Error())
		}
	}()

	ctx = logger.WithContext(ctx)

	mongoClient := mongokit.NewMongoClient(ctx, &args.MongoConfig)
	defer func() {
		if err := mongoClient.Close(); err != nil {
			logger.Fatal("failed to close mongo client", zap.Error(err))
		}
	}()

	videoDAO := dao.NewVideoMongoDAO(mongoClient.Database().Collection("videos"))
	storage := storagekit.NewMinIOClient(ctx, &args.MinIOConfig)
	svc := service.NewService(videoDAO, storage)

	logger.Info("listen to gRPC addr", zap.String("grpc_addr", args.GRPCAddr))
	lis, err := net.Listen("tcp", args.GRPCAddr)
	if err != nil {
		logger.Fatal("failed to listen gRPC addr", zap.Error(err))
	}
	defer func() {
		if err := lis.Close(); err != nil {
			logger.Fatal("failed to close gRPC listener", zap.Error(err))
		}
	}()

	return runkit.GracefulRun(serveGRPC(lis, svc, logger), &args.GracefulConfig)
}

func serveGRPC(lis net.Listener, svc pb.VideoServer, logger *logkit.Logger) runkit.GracefulRunFunc {
	grpcServer := grpc.NewServer()
	pb.RegisterVideoServer(grpcServer, svc)

	return func(ctx context.Context) error {
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				logger.Error("failed to run gRPC server", zap.Error(err))
			}
		}()

		<-ctx.Done()

		grpcServer.GracefulStop()

		return nil
	}
}
