package mongokit

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMongoKit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Mongo Kit")
}
