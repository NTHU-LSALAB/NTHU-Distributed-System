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

type Reader interface {
	ReadMessages(ctx context.Context) (kafka.Message, error)
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

func (kr *KafkaReader) ReadMessages(ctx context.Context) (kafka.Message, error) {
	m, err := kr.ReadMessage(ctx)

	return m, err
}
