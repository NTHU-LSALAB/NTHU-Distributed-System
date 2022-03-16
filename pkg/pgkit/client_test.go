package pgkit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PGClient", func() {
	Describe("NewPGClient", func() {
		var (
			ctx      context.Context
			pgConf   *PGConfig
			pgClient *PGClient
		)

		BeforeEach(func() {
			ctx = logkit.NewLogger(&logkit.LoggerConfig{
				Development: true,
			}).WithContext(context.Background())

			pgConf = &PGConfig{
				URL: "postgres://postgres@postgres:5432/postgres?sslmode=disable",
			}
			if url := os.Getenv("POSTGRES_URL"); url != "" {
				pgConf.URL = url
			}
		})

		JustBeforeEach(func() {
			pgClient = NewPGClient(ctx, pgConf)
		})

		AfterEach(func() {
			Expect(pgClient.Close()).NotTo(HaveOccurred())
		})

		When("success", func() {
			It("returns new PGClient without error", func() {
				Expect(pgClient).NotTo(BeNil())
			})
		})
	})
})
