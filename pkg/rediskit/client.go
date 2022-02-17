package rediskit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisConfig struct {
	URL      string `long:"url" env:"URL" description:"the URL of Redis" required:"true"`
	Addr     string `long:"addr" env:"ADDR" description:"the Address of Redis" required:"true"`
	Password string `long:"password" env:"PASSWORD" description:"the Password of Redis"`
	DB       int    `long:"db" env:"DB" description:"default db"`
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
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
	}

	logger := logkit.FromContext(ctx).
		With(zap.String("url", conf.URL)).
		With(zap.String("addr", conf.Addr)).
		With(zap.String("password", conf.Password)).
		With(zap.Int("db", conf.DB))

	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		logger.Fatal("failed to ping to Redis", zap.Error(err))
	}

	return &RedisClient{
		Client: client,
	}
}
