package migrationkit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"go.uber.org/zap"
)

type MigrationConfig struct {
	Source string `long:"source" env:"SOURCE" description:"the migration files source directory" required:"true"`
	URL    string `long:"url" env:"URL" description:"the database url" required:"true"`
}

type Migration struct {
	*migrate.Migrate
}

func (m *Migration) Close() error {
	serr, derr := m.Migrate.Close()
	if serr != nil {
		return serr
	}

	if derr != nil {
		return derr
	}

	return nil
}

func NewMigration(ctx context.Context, conf *MigrationConfig) *Migration {
	if url := os.ExpandEnv(conf.URL); url != "" {
		conf.URL = url
	}

	logger := logkit.FromContext(ctx).With(
		zap.String("source", conf.Source),
		zap.String("url", conf.URL),
	)

	m, err := migrate.New(conf.Source, conf.URL)
	if err != nil {
		logger.Fatal("failed to create migration", zap.Error(err))
	}

	return &Migration{
		Migrate: m,
	}
}
