package mongokit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoClient", func() {
	Describe("NewMongoClient", func() {
		var (
			mongoClient *MongoClient
			ctx         context.Context
			mongoConfig *MongoConfig
		)

		BeforeEach(func() {
			ctx = logkit.NewLogger(&logkit.LoggerConfig{
				Development: true,
			}).WithContext(context.Background())

			mongoConfig = &MongoConfig{
				URL:      "mongodb://mongo:27017",
				Database: "nthu_distributed_system",
			}

			if url := os.Getenv("MONGO_URL"); url != "" {
				mongoConfig.URL = url
			}

			if database := os.Getenv("MONGO_DATABASE"); database != "" {
				mongoConfig.Database = database
			}
		})

		JustBeforeEach(func() {
			mongoClient = NewMongoClient(ctx, mongoConfig)
		})

		AfterEach(func() {
			Expect(mongoClient.Close()).NotTo(HaveOccurred())
		})

		When("success", func() {
			It("returns new MongoClient without error", func() {
				Expect(mongoClient).NotTo(BeNil())
			})
		})
	})
})
