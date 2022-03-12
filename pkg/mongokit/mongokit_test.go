package mongokit

import (
	"context"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mongokit", func() {
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
				Database: "video",
			}
		})

		JustBeforeEach(func() {
			mongoClient = NewMongoClient(ctx, mongoConfig)
		})

		AfterEach(func() {
			Expect(mongoClient.Close()).NotTo(HaveOccurred())
		})

		When("success", func() {
			It("returns new Mongokit without error", func() {
				Expect(mongoClient).NotTo(BeNil())
			})
		})
	})
})
