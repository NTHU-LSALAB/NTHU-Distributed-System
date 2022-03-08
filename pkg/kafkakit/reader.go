package kafkakit

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaReaderConfig struct {
	Brokers []string `long:"brokers" env:"BROKERS" description:"the address of kakfa server" required:"true"`
	Topic   string   `long:"topic" env:"TOPIC" description:"the topic of changing resolution" required:"true"`
	GroupID string   `long:"group_id" env:"GROUP_ID" description:"the id of the consumer groups" required:"true"`
}

type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, messages []kafka.Message) error
}

type KafkaReader struct {
	*kafka.Reader
	closeFunc func()
}

func (kr *KafkaReader) Close() error {
	if kr.closeFunc != nil {
		kr.closeFunc()
	}

	return kr.Reader.Close()
}

func NewKafkaReader(ctx context.Context, conf *KafkaReaderConfig) *KafkaReader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  conf.Brokers,
		GroupID:  conf.GroupID,
		Topic:    conf.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &KafkaReader{
		Reader: reader,
	}
}

func (kr *KafkaReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	m, err := kr.Reader.FetchMessage(ctx)

	return m, err
}

func (kr *KafkaReader) CommitMessages(ctx context.Context, messages []kafka.Message) error {
	err := kr.Reader.CommitMessages(ctx, messages...)

	return err
}
