package migrationkit

import (
	"context"
	"errors"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type MigrationConfig struct {
	// Source is the migration files source directory,
	// currently we only accept file system as source.
	//
	// Register more source types by importing other source packages.
	//
	// File System: https://github.com/golang-migrate/migrate/tree/master/source/file
	Source string `long:"source" env:"SOURCE" description:"the migration files source directory" required:"true"`
	// URL is the migration database URL,
	// currently we only accept Postgres as database.
	//
	// Register more database types by importing other database packages.
	//
	// Postgres: https://github.com/golang-migrate/migrate/tree/master/database/postgres
	URL string `long:"url" env:"URL" description:"the database url" required:"true"`
}

type Migration struct {
	*migrate.Migrate
	logger *logkit.Logger
}

func (m *Migration) Up() error {
	if err := m.Migrate.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("no change to migrate")

			return nil
		}

		return err
	}

	return nil
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

	logger.Info("create migration successfully")

	return &Migration{
		Migrate: m,
		logger:  logger,
	}
}
