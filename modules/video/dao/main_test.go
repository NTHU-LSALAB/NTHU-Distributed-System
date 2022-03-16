package dao

import (
	"context"
	"os"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/mongokit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDAO(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test DAO")
}

var (
	mongoClient *mongokit.MongoClient
	redisClient *rediskit.RedisClient
)

var _ = BeforeSuite(func() {
	mongoConf := &mongokit.MongoConfig{
		URL:      "mongodb://mongo:27017",
		Database: "video",
	}

	if url := os.Getenv("MONGO_URL"); url != "" {
		mongoConf.URL = url
	}

	if database := os.Getenv("MONGO_DATABASE"); database != "" {
		mongoConf.Database = database
	}

	redisConf := &rediskit.RedisConfig{
		Addr: "redis:6379",
	}

	ctx := logkit.NewLogger(&logkit.LoggerConfig{
		Development: true,
	}).WithContext(context.Background())

	mongoClient = mongokit.NewMongoClient(ctx, mongoConf)
	redisClient = rediskit.NewRedisClient(ctx, redisConf)
})

var _ = AfterSuite(func() {
	Expect(mongoClient.Close()).NotTo(HaveOccurred())
	Expect(redisClient.Close()).NotTo(HaveOccurred())
})
