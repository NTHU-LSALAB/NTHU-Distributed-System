package rediskit

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisConfig struct {
	Addr     string `long:"addr" env:"ADDR" description:"the address of Redis" required:"true"`
	Password string `long:"password" env:"PASSWORD" description:"the password of Redis"`
	Database int    `long:"database" env:"DATABASE" description:"the database of Redis"`
}

type RedisClient struct {
	*redis.Client
	closeFunc func()
}

func (c *RedisClient) Close() error {
	if c.closeFunc != nil {
		c.closeFunc()
	}

	return c.Client.Close()
}

func NewRedisClient(ctx context.Context, conf *RedisConfig) *RedisClient {
	logger := logkit.FromContext(ctx).With(
		zap.String("addr", conf.Addr),
		zap.Int("database", conf.Database),
	)

	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.Database,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Fatal("failed to ping to Redis", zap.Error(err))
	}

	return &RedisClient{
		Client: client,
	}
}
