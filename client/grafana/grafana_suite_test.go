// +build unitTests

package grafana_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGrafana(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Grafana Suite")
}
