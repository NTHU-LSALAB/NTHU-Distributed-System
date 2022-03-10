package kafkakit

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type KafkaConsumerConfig struct {
	Addrs []string `long:"addrs" env:"ADDRS" env-delim:"," description:"the addresses of Kafka servers" required:"true"`
	Topic string   `long:"topic" env:"TOPIC" description:"the topic for the Kafka consumer group to consume" required:"true"`
	Group string   `long:"group" env:"GROUP" description:"the ID of the Kafka consumer group" required:"true"`
}

type KafkaConsumer struct {
	sarama.ConsumerGroup

	topic string
}

func (kc *KafkaConsumer) Consume(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
	for {
		if err := kc.ConsumerGroup.Consume(ctx, []string{kc.topic}, handler); err != nil {
			return err
		}
	}
}

func (kc *KafkaConsumer) Close() error {
	return kc.ConsumerGroup.Close()
}

func NewKafkaConsumer(ctx context.Context, conf *KafkaConsumerConfig) *KafkaConsumer {
	logger := logkit.FromContext(ctx).With(
		zap.Strings("addrs", conf.Addrs),
		zap.String("topic", conf.Topic),
		zap.String("group", conf.Group),
	)

	config := sarama.NewConfig()

	cg, err := sarama.NewConsumerGroup(conf.Addrs, conf.Group, config)
	if err != nil {
		logger.Fatal("failed to create Kafka consumer group", zap.Error(err))
	}

	logger.Info("create Kafka consumer successfully")

	return &KafkaConsumer{
		ConsumerGroup: cg,
		topic:         conf.Topic,
	}
}
