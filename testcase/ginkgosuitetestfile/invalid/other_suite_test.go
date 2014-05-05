package invalid_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestInvalid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Suite")
}
