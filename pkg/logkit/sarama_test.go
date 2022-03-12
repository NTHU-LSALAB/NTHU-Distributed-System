package logkit

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SaramaLogger", func() {
	Describe("NewSaramaLogger", func() {
		var logger *Logger
		var saramaLogger *SaramaLogger

		BeforeEach(func() {
			logger = NewLogger(&LoggerConfig{Development: true})
		})

		JustBeforeEach(func() {
			saramaLogger = NewSaramaLogger(logger)
		})

		When("success", func() {
			It("returns new SaramaLogger without error", func() {
				Expect(saramaLogger).NotTo(BeNil())
			})
		})
	})
})
