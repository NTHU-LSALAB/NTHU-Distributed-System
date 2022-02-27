package rediskit

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRediskit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Rediskit")
}

var _ = Describe("Rediskit", func() {
	var ctx context.Context
	var redisClient *RedisClient
	var redisConf RedisConfig

	BeforeEach(func() {
		ctx = context.Background()
		redisConf.Addr = "localhost:6379"
		if addr := os.Getenv("REDIS_ADDR"); addr != "" {
			redisConf.Addr = addr
		}
	})

	AfterEach(func() {
		Expect(redisClient.Close()).NotTo(HaveOccurred())
	})

	Describe("NewRedisClient", func() {
		JustBeforeEach(func() {
			redisClient = NewRedisClient(ctx, &redisConf)
		})

		Context("success", func() {
			It("returns new RedisClient", func() {
				Expect(redisClient).NotTo(BeNil())
			})
		})
	})
})
