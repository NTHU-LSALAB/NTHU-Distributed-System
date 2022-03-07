package kafkakit

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type KafkaWriterConfig struct {
	// Wait for Justin to help
	Addr  string `long:"kaddr" env:"KAFKA_ADDR" description:"the address of kakfa server" required:"true"`
	Topic string `long:"ktopic" env:"KAFKA_TOPIC" description:"the topic of changing resolution" required:"true"`
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
	writer := &kafka.Writer{
		Addr:     kafka.TCP(conf.Addr),
		Topic:    conf.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaWriter{
		Writer: writer,
	}
}

func (kw *KafkaWriter) WriteMessage(ctx context.Context, messageValue string) error {
	msg := kafka.Message{
		Value: []byte(messageValue),
	}
	err := kw.Writer.WriteMessages(ctx, msg)
	return err
}
