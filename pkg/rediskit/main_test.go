package rediskit

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRedisKit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Redis Kit")
}
