package dao

import (
	"context"
	"os"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/migrationkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/pgkit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDAO(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test DAO")
}

var (
	pgClient *pgkit.PGClient
)

var _ = BeforeSuite(func() {
	pgConf := &pgkit.PGConfig{
		URL: "postgres://postgres@postgres:5432/postgres?sslmode=disable",
	}
	if url := os.Getenv("POSTGRES_URL"); url != "" {
		pgConf.URL = url
	}

	migrationConf := &migrationkit.MigrationConfig{
		Source: "file://../migration",
		URL:    pgConf.URL,
	}

	ctx := logkit.NewLogger(&logkit.LoggerConfig{
		Development: true,
	}).WithContext(context.Background())

	migration := migrationkit.NewMigration(ctx, migrationConf)
	defer func() {
		Expect(migration.Close()).NotTo(HaveOccurred())
	}()

	Expect(migration.Up()).NotTo(HaveOccurred())

	pgClient = pgkit.NewPGClient(ctx, pgConf)
})

var _ = AfterSuite(func() {
	Expect(pgClient.Close()).NotTo(HaveOccurred())
})
