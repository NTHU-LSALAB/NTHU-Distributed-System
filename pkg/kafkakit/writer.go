package kafkakit

import (
	"context"
	"os"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	// Wait for Justin to help
	kafkaAddr  string `long:"kaddr" env:"KAFKA_ADDR" description:"the address of kakfa server" required:"true"`
	kafkaTopic string `long:"ktopic" env:"KAFKA_TOPIC" description:"the topic of changing resolution" required:"true"`
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

func NewKafkaWriter(ctx context.Context, conf *KafkaConfig) *KafkaWriter {
	if addr := os.ExpandEnv(conf.kafkaAddr); addr != "" {
		conf.kafkaAddr = addr
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(conf.kafkaAddr),
		Topic:    conf.kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaWriter{
		Writer: writer,
	}
}
