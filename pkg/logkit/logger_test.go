package logkit

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	Describe("NewLogger", func() {
		var logger *Logger

		JustBeforeEach(func() {
			logger = NewLogger(&LoggerConfig{Development: true})
		})

		When("success", func() {
			It("returns new logger without error", func() {
				Expect(logger).NotTo(BeNil())
			})
		})
	})

	Describe("WithContext", func() {
		var logger *Logger
		var ctx context.Context

		JustBeforeEach(func() {
			logger = NewLogger(&LoggerConfig{Development: true})
			ctx = logger.WithContext(context.Background())
		})

		When("success", func() {
			It("inserts logger into context", func() {
				Expect(FromContext(ctx)).To(Equal(logger))
			})
		})
	})
})
