package comment

import (
	"context"
	"log"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/migrationkit"
	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newMigrationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migration",
		Short: "runs the comment module migration job",
		RunE:  runMigration,
	}
}

type MigrationArgs struct {
	logkit.LoggerConfig          `group:"logger" namespace:"logger" env-namespace:"LOGGER"`
	migrationkit.MigrationConfig `group:"migration" namespace:"migration" env-namespace:"MIGRATION"`
}

func runMigration(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	var args MigrationArgs
	if _, err := flags.NewParser(&args, flags.Default).Parse(); err != nil {
		log.Fatal("failed to parse flag", err.Error())
	}

	logger := logkit.NewLogger(&args.LoggerConfig)
	defer func() {
		_ = logger.Sync()
	}()

	ctx = logger.WithContext(ctx)

	migration := migrationkit.NewMigration(ctx, &args.MigrationConfig)
	if err := migration.Up(); err != nil {
		logger.Fatal("failed to run migration", zap.Error(err))
	}

	logger.Info("run migration job successfully, terminating ...")

	return nil
}
