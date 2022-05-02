package video

import (
	"context"
	"log"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/dao"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/pb"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/modules/video/stream"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/kafkakit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/mongokit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/runkit"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newStreamCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stream",
		Short: "starts video stream server",
		RunE:  runStream,
	}
}

type StreamArgs struct {
	runkit.GracefulConfig        `group:"graceful" namespace:"graceful" env-namespace:"GRACEFUL"`
	logkit.LoggerConfig          `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	mongokit.MongoConfig         `group:"mongo" namespace:"mongo" env-namespace:"MONGO"`
	kafkakit.KafkaProducerConfig `group:"kafka_producer" namespace:"kafka_producer" env-namespace:"KAFKA_PRODUCER"`
	kafkakit.KafkaConsumerConfig `group:"kafka_consumer" namespace:"kafka_consumer" env-namespace:"KAFKA_CONSUMER"`
}

func runStream(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args StreamArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("failed to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		_ = logger.Sync()
	}()

	ctx = logger.WithContext(ctx)

	mongoClient := mongokit.NewMongoClient(ctx, &args.MongoConfig)
	defer func() {
		if err := mongoClient.Close(); err != nil {
			logger.Fatal("failed to close mongo client", zap.Error(err))
		}
	}()

	producer := kafkakit.NewKafkaProducer(ctx, &args.KafkaProducerConfig)
	defer func() {
		if err := producer.Close(); err != nil {
			logger.Fatal("failed to close Kafka producer", zap.Error(err))
		}
	}()

	consumer := kafkakit.NewKafkaConsumer(ctx, &args.KafkaConsumerConfig)
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Fatal("failed to close Kafka consumer", zap.Error(err))
		}
	}()

	videoDAO := dao.NewMongoVideoDAO(mongoClient.Database().Collection("videos"))

	svc := stream.NewStream(videoDAO, producer)

	return runkit.GracefulRun(serveConsumer(consumer, svc, logger), &args.GracefulConfig)
}

func serveConsumer(consumer *kafkakit.KafkaConsumer, svc pb.VideoStreamServer, logger *logkit.Logger) runkit.GracefulRunFunc {
	handlers := pb.NewVideoStreamHandlers(svc, logkit.NewSaramaLogger(logger))

	return func(ctx context.Context) error {
		if err := consumer.Consume(ctx, handlers.HandleVideoCreatedHandler); err != nil {
			return err
		}

		return nil
	}
}
