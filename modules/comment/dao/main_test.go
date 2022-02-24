package dao

import (
	"context"
	"os"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
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
	var pgConf pgkit.PGConfig

	pgConf.URL = "postgres://postgres@postgres:5432/postgres?sslmode=disable"
	if url := os.Getenv("POSTGRES_URL"); url != "" {
		pgConf.URL = url
	}

	ctx := logkit.NewLogger(&logkit.LoggerConfig{
		Development: true,
	}).WithContext(context.Background())

	pgClient = pgkit.NewPGClient(ctx, &pgConf)
})

var _ = AfterSuite(func() {
	Expect(pgClient.Close()).NotTo(HaveOccurred())
})

var pgExec = func(query string, params ...interface{}) {
	_, err := pgClient.Exec(query, params...)
	Expect(err).NotTo(HaveOccurred())
}
