package mongokit

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mongokit", func() {
	Describe("NewMongoClient", func() {
		var (
			mongoclient *MongoClient
			ctx context.Context
		)

		JustBeforeEach(func() {
			mongoclient = NewMongoClient(ctx, &MongoConfig{
				URL:      "mongodb://mongo:27017",
				Database: "video",
			})
		})

		When("success", func() {
			It("returns new Mongokit without error", func() {
				Expect(mongoClient.Close()).NotTo(HaveOccurred())
			})
		})
	})
})
