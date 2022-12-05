// +build unitTests

package grafana_test

import (
	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const FakeOrgID = "999"

func ClientWithMockServer() (*Client, *ghttp.Server) {
	server := ghttp.NewServer()
	client, err := New(server.URL())
	Expect(err).NotTo(HaveOccurred())
	return client, server
}
