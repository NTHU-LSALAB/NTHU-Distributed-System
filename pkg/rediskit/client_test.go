package rediskit

import (
	"context"
	"os"

	"github.com/NTHU-LSALAB/NTHU-Distributed-System/pkg/logkit"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RedisClient", func() {
	Describe("NewRedisClient", func() {
		var (
			ctx         context.Context
			redisClient *RedisClient
			redisConf   RedisConfig
		)

		BeforeEach(func() {
			ctx = logkit.WithContext(context.Background(), logkit.NewNopLogger())

			redisConf.Addr = "localhost:6379"
			if addr := os.Getenv("REDIS_ADDR"); addr != "" {
				redisConf.Addr = addr
			}
		})

		AfterEach(func() {
			Expect(redisClient.Close()).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			redisClient = NewRedisClient(ctx, &redisConf)
		})

		When("success", func() {
			It("returns redis client without error", func() {
				Expect(redisClient).NotTo(BeNil())
			})
		})
	})
})
