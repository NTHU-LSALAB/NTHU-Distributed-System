package pgkit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-pg/pg/v10"
	"go.uber.org/zap"
)

type PGConfig struct {
	URL string `long:"url" env:"URL" description:"the URL of PostgreSQL" required:"true"`
}

type PGClient struct {
	*pg.DB
	closeFunc func()
}

func (c *PGClient) Close() error {
	if c.closeFunc != nil {
		c.closeFunc()
	}
	return c.DB.Close()
}

func NewPGClient(ctx context.Context, conf *PGConfig) *PGClient {
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
	}

	logger := logkit.FromContext(ctx).With(zap.String("url", conf.URL))
	opts, err := pg.ParseURL(conf.URL)
	if err != nil {
		logger.Fatal("failed to parse PostgreSQL url", zap.Error(err))
	}

	db := pg.Connect(opts).WithContext(ctx)
	if err := db.Ping(ctx); err != nil {
		logger.Fatal("failed to ping PostgreSQL", zap.Error(err))
	}

	logger.Info("create PostgreSQL client successfully")

	return &PGClient{
		DB: db,
	}
}
