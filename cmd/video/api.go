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
	GRPCAddr              string `long:"grpc_addr" env:"GRPC_ADDR" default:":8081"`
	runkit.GracefulConfig `group:"graceful" namespace:"graceful" env-namespace:"GRACEFUL"`
	logkit.LoggerConfig   `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	mongokit.MongoConfig  `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`
}

func runAPI(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args APIArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("fail to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatal("fail to sync logger", err.Error())
		}
	}()

	ctx = logger.WithContext(ctx)

	mongoClient := mongokit.NewMongoClient(ctx, &args.MongoConfig)
	defer func() {
		if err := mongoClient.Close(); err != nil {
			logger.Fatal("fail to close mongo client", zap.Error(err))
		}
	}()

	videoDAO := dao.NewVideoMongoDAO(mongoClient.Database().Collection("videos"))
	svc := service.NewService(videoDAO)

	return runkit.GracefulRun(serveGRPC(args.GRPCAddr, svc, logger), &args.GracefulConfig)
}

func serveGRPC(addr string, svc pb.VideoServer, logger *logkit.Logger) runkit.GracefulRunFunc {
	logger.Info("listen to gRPC addr", zap.String("grpc_addr", addr))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatal("fail to listen gRPC addr", zap.Error(err))
	}
	defer func() {
		if err := lis.Close(); err != nil {
			logger.Fatal("fail to close gRPC listener", zap.Error(err))
		}
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterVideoServer(grpcServer, svc)

	return func(ctx context.Context) error {
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				logger.Error("fail to run gRPC server", zap.Error(err))
			}
		}()

		<-ctx.Done()

		grpcServer.GracefulStop()

		return nil
	}
}
