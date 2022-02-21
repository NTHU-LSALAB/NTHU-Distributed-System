package dao

import (
	"context"
	"os"
	"testing"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/mongokit"
	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/rediskit"
	. "github.com/onsi/ginkgo"
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
	var mongoConf mongokit.MongoConfig
	var redisConf rediskit.RedisConfig

	mongoConf.URL = "mongodb://localhost:27017"
	if url := os.Getenv("MONGO_URL"); url != "" {
		mongoConf.URL = url
	}

	mongoConf.Database = "video"
	if database := os.Getenv("MONGO_DATABASE"); database != "" {
		mongoConf.Database = database
	}

	redisConf.Addr = "localhost:6379"
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		redisConf.Addr = addr
	}

	ctx := logkit.NewLogger(&logkit.LoggerConfig{
		Development: true,
	}).WithContext(context.Background())

	redisClient = rediskit.NewRedisClient(ctx, &redisConf)
	mongoClient = mongokit.NewMongoClient(ctx, &mongoConf)
})

var _ = AfterSuite(func() {
	Expect(mongoClient.Close()).NotTo(HaveOccurred())
	Expect(redisClient.Close()).NotTo(HaveOccurred())
})
