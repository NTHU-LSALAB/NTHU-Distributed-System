package comment

import (
	"context"
	"log"
	"net"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/comment/service"
	videopb "github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/grpckit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/otelkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/pgkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/runkit"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newAPICommand() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "starts comment API server",
		RunE:  runAPI,
	}
}

type APIArgs struct {
	GRPCAddr                             string                       `long:"grpc_addr" env:"GRPC_ADDR" default:":8081"`
	VideoClientConnConfig                grpckit.GrpcClientConnConfig `group:"video" namespace:"video" env-namespace:"VIDEO"`
	runkit.GracefulConfig                `group:"graceful" namespace:"graceful" env-namespace:"GRACEFUL"`
	logkit.LoggerConfig                  `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	pgkit.PGConfig                       `group:"postgres" namespace:"postgres" env-namespace:"POSTGRES"`
	rediskit.RedisConfig                 `group:"redis" namespace:"redis" env-namespace:"REDIS"`
	otelkit.PrometheusServiceMeterConfig `group:"meter" namespace:"meter" env-namespace:"METER"`
}

func runAPI(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args APIArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("failed to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		_ = logger.Sync()
	}()

	ctx = logger.WithContext(ctx)

	pgClient := pgkit.NewPGClient(ctx, &args.PGConfig)
	defer func() {
		if err := pgClient.Close(); err != nil {
			logger.Fatal("failed to close pg client", zap.Error(err))
		}
	}()

	redisClient := rediskit.NewRedisClient(ctx, &args.RedisConfig)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Fatal("failed to close redis client", zap.Error(err))
		}
	}()

	videoClientConn := grpckit.NewGrpcClientConn(ctx, &args.VideoClientConnConfig)
	defer func() {
		if err := videoClientConn.Close(); err != nil {
			logger.Fatal("failed to close video gRPC client", zap.Error(err))
		}
	}()

	pgCommentDAO := dao.NewPGCommentDAO(pgClient)
	commentDAO := dao.NewRedisCommentDAO(redisClient, pgCommentDAO)
	videoClient := videopb.NewVideoClient(videoClientConn)

	svc := service.NewService(commentDAO, videoClient)

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

	meter := otelkit.NewPrometheusServiceMeter(ctx, &args.PrometheusServiceMeterConfig)
	defer func() {
		if err := meter.Close(); err != nil {
			logger.Fatal("failed to close meter", zap.Error(err))
		}
	}()

	return runkit.GracefulRun(serveGRPC(lis, svc, logger, grpc.UnaryInterceptor(meter.UnaryServerInterceptor())), &args.GracefulConfig)
}

func serveGRPC(lis net.Listener, svc pb.CommentServer, logger *logkit.Logger, opt ...grpc.ServerOption) runkit.GracefulRunFunc {
	grpcServer := grpc.NewServer(opt...)
	pb.RegisterCommentServer(grpcServer, svc)

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
