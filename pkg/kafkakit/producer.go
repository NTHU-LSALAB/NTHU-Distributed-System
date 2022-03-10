package kafkakit

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
)

type Producer interface {
	SendMessages(msgs []*ProducerMessage) error
}

type ProducerMessage struct {
	Key   []byte
	Value []byte
}

type KafkaProducerConfig struct {
	Addrs        []string `long:"addrs" env:"ADDRS" env-delim:"," description:"the addresses of Kafka servers" required:"true"`
	Topic        string   `long:"topic" env:"TOPIC" description:"the topic for the Kafka producer to send" required:"true"`
	RequiredAcks int16    `long:"required_acks" env:"REQUIRED_ACKS" description:"number of replica acks the producer must receive before responding, available values are 0, 1 and -1" default:"-1"`
}

type KafkaProducer struct {
	sarama.SyncProducer

	topic string
}

var _ Producer = (*KafkaProducer)(nil)

func (kp *KafkaProducer) SendMessages(msgs []*ProducerMessage) error {
	smsgs := make([]*sarama.ProducerMessage, 0, len(msgs))
	for _, msg := range msgs {
		smsgs = append(smsgs, &sarama.ProducerMessage{
			Topic: kp.topic,
			Key:   sarama.ByteEncoder(msg.Key),
			Value: sarama.ByteEncoder(msg.Value),
		})
	}

	return kp.SyncProducer.SendMessages(smsgs)
}

func (kp *KafkaProducer) Close() error {
	return kp.SyncProducer.Close()
}

func NewProducer(ctx context.Context, conf *KafkaProducerConfig) *KafkaProducer {
	logger := logkit.FromContext(ctx).With(
		zap.Strings("addrs", conf.Addrs),
		zap.String("topic", conf.Topic),
		zap.Int16("required_acks", conf.RequiredAcks),
	)

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.RequiredAcks(conf.RequiredAcks)

	// If this config is used to create a `SyncProducer`, both must be set
	// to true and you shall not read from the channels since the producer
	// does this internally.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(conf.Addrs, config)
	if err != nil {
		logger.Fatal("failed to create Kafka sync producer", zap.Error(err))
	}

	logger.Info("create Kafka producer successfully")

	return &KafkaProducer{
		SyncProducer: producer,
		topic:        conf.Topic,
	}
}
