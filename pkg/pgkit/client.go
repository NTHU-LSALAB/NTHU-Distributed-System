package pgkit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/go-pg/pg/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
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
	logger := logkit.FromContext(ctx).With(zap.String("url", conf.URL))
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
		m, err := migrate.New(
			"file://src/modules/migrations",
			url,
		)

		if err != nil {
			logger.Fatal("Failed to create database table", zap.Error(err))
		}

		if err := m.Down(); err != nil {
			logger.Fatal("Failed to create database table", zap.Error(err))
		}
	}

	opts, err := pg.ParseURL(conf.URL)
	if err != nil {
		logger.Fatal("Failed to parse PostgreSQL url", zap.Error(err))
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
