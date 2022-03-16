package migrationkit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/pgkit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migration", func() {
	Describe("NewMigration", func() {
		var (
			ctx           context.Context
			migration     *Migration
			migrationConf *MigrationConfig
		)

		BeforeEach(func() {
			pgConf := &pgkit.PGConfig{
				URL: "postgres://postgres@postgres:5432/postgres?sslmode=disable",
			}
			if url := os.Getenv("POSTGRES_URL"); url != "" {
				pgConf.URL = url
			}

			migrationConf = &MigrationConfig{
				Source: "file://.",
				URL:    pgConf.URL,
			}

			ctx = logkit.NewLogger(&logkit.LoggerConfig{
				Development: true,
			}).WithContext(context.Background())
		})

		JustBeforeEach(func() {
			migration = NewMigration(ctx, migrationConf)
		})

		AfterEach(func() {
			Expect(migration.Close()).NotTo(HaveOccurred())
		})

		When("success", func() {
			It("returns new Migration without error", func() {
				Expect(migration).NotTo(BeNil())
			})
		})
	})
})
