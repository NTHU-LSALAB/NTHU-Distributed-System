package kafkakit

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaWriterConfig struct {
	// Wait for Justin to help
	Brokers []string `long:"brokers" env:"BROKERS" description:"the address of kakfa server" required:"true"`
	Topic   string   `long:"topic" env:"TOPIC" description:"the topic of changing resolution" required:"true"`
}

type Writer interface {
	WriteMessages(context.Context, []kafka.Message) error
}

type KafkaWriter struct {
	*kafka.Writer
	closeFunc func()
}

func (kw *KafkaWriter) Close() error {
	if kw.closeFunc != nil {
		kw.closeFunc()
	}

	return kw.Writer.Close()
}

func NewKafkaWriter(ctx context.Context, conf *KafkaWriterConfig) *KafkaWriter {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  conf.Brokers,
		Topic:    conf.Topic,
		Balancer: &kafka.LeastBytes{},
	})

	return &KafkaWriter{
		Writer: writer,
	}
}

func (kw *KafkaWriter) WriteMessages(ctx context.Context, messages []kafka.Message) error {
	return kw.Writer.WriteMessages(ctx, messages...)
}
