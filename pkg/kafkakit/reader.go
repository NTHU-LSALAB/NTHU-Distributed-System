package kafkakit

import (
	"context"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

type KafkaReader struct {
	*kafka.Reader
	closeFunc func()
}

func (kw *KafkaReader) Close() error {
	if kw.closeFunc != nil {
		kw.closeFunc()
	}

	return kw.Reader.Close()
}

func NewKafkaReader(ctx context.Context, conf *KafkaConfig) *KafkaReader {
	if addr := os.ExpandEnv(conf.kafkaAddr); addr != "" {
		conf.kafkaAddr = addr
	}

	brokers := strings.Split(conf.kafkaAddr, ",")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     conf.kafkaTopic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	return &KafkaReader{
		Reader: reader,
	}
}
