package postgreskit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-pg/pg/v9"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type PGConfig struct {
	URL string `long:"url" env:"URL" description:"the URL of PostgreSQL" required:"true"`
}

type PGClient struct {
	*pg.DB
	Closed atomic.Bool
}

func (c *PGClient) Close() error {
	c.Closed.Store(true)
	return c.DB.Close()
}

func NewPGClient(ctx context.Context, conf *PGConfig) *PGClient {
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
	}

	logger := logkit.FromContext(ctx).With(zap.String("url", conf.URL))
	opts, err := pg.ParseURL(conf.URL)
	if err != nil {
		logger.Fatal("Failed to parse PostgreSQL url", zap.Error(err))
	}

	db := pg.Connect(opts).WithContext(ctx)
	if _, err := db.Exec("SELECT 1"); err != nil {
		logger.Fatal("failed to pin PostgreSQL", zap.Error(err))
	}

	logger.Info("create PostgreSQL client successfully")

	return &PGClient{
		DB: db,
	}
}
