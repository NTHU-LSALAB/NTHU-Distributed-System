package logkit

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLogKit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Log Kit")
}
