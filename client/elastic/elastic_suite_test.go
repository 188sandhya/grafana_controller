// +build unitTests

package elastic_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestElastic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Elastic Suite")
}
