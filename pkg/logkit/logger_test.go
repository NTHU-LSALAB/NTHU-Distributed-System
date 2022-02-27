package logkit

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLoggerkit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Loggerkit")
}

var _ = Describe("Loggerkit", func() {
	Describe("NewLogger", func() {
		var logger *Logger

		JustBeforeEach(func() {
			logger = NewLogger(&LoggerConfig{Development: true})
		})

		Context("success", func() {
			It("returns new Logger", func() {
				Expect(logger).NotTo(BeNil())
			})
		})

	})

	Describe("WithContext", func() {
		var logger *Logger
		ctx := context.Background()

		JustBeforeEach(func() {
			ctx = WithContext(ctx, logger)
		})

		Context("success", func() {
			It("Return looger and context", func() {
				Expect(ctx).NotTo(BeNil())
			})
		})
	})

	Describe("Loggerlevel", func() {
		var (
			level       string
			getLevel    string
			loggerlevel LoggerLevel
			err         error
		)

		BeforeEach(func() { level = "info" })

		When("MarshalFlag", func() {
			JustBeforeEach(func() {
				getLevel, err = loggerlevel.MarshalFlag()
			})

			Context("success", func() {
				It("marshal loggerlevel correctly with no error", func() {
					Expect(getLevel).To(Equal(level))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		When("UnmarshalFlag", func() {
			JustBeforeEach(func() {
				err = loggerlevel.UnmarshalFlag(level)
			})

			Context("success", func() {
				It("unmarshal level with no error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
