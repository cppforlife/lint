package valid_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestValid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite")
}
