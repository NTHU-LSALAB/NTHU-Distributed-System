package kafkakit

import (
	"context"
	"strings"

	"github.com/segmentio/kafka-go"
)

type KafkaReaderConfig struct {
	Addr  string
	Topic string
}

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

func NewKafkaReader(ctx context.Context, conf *KafkaReaderConfig) *KafkaReader {
	brokers := strings.Split(conf.Addr, ",")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   brokers,
		Topic:     conf.Topic,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	return &KafkaReader{
		Reader: reader,
	}
}
