package dao

import (
	"context"
	"os"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/mongokit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDAO(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test DAO")
}

var (
	mongoClient *mongokit.MongoClient
)

var _ = BeforeSuite(func() {
	var conf mongokit.MongoConfig

	conf.URL = "mongodb://localhost:27017"
	if url := os.Getenv("MONGO_URL"); url != "" {
		conf.URL = url
	}

	conf.Database = "video"
	if database := os.Getenv("MONGO_DATABASE"); database != "" {
		conf.Database = database
	}

	ctx := logkit.NewLogger(&logkit.LoggerConfig{
		Development: true,
	}).WithContext(context.Background())

	mongoClient = mongokit.NewMongoClient(ctx, &conf)
})

var _ = AfterSuite(func() {
	Expect(mongoClient.Close()).NotTo(HaveOccurred())
})
