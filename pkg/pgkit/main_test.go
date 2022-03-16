package pgkit

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPGKit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test PG Kit")
}
