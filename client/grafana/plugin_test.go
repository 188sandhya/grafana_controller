// +build unitTests

package grafana_test

import (
	"fmt"
	"net/http"
	"strconv"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("EnablePlugin", func() {
	const fakeCookie = "test_cookie"

	var client *Client
	var server *ghttp.Server
	var statusCode int
	var pluginName = "test-plugin-name"
	var testPluginSettings = model.PluginSettings{
		Name:    pluginName,
		Enabled: true,
	}
	Describe("EnablePlugin()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		BeforeEach(func() {
			client, server = ClientWithMockServer()
			statusCode = http.StatusOK

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", fmt.Sprintf("/api/plugins/%s/settings", pluginName)),
				ghttp.RespondWithPtr(&statusCode, nil),
			))
		})
		Describe("When Pass", func() {
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns no error", func() {
					err := client.EnablePlugin(&testPluginSettings, fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					receivedRequests := server.ReceivedRequests()
					Expect(receivedRequests).To(HaveLen(1))
				})
			})
		})

		Describe("When Fail", func() {
			AfterEach(func() {
				server.Close()
			})
			Context("When http.StatusBadRequest response is received", func() {
				BeforeEach(func() {
					statusCode = http.StatusBadRequest
				})
				It("Returns error and nil", func() {
					err := client.EnablePlugin(&testPluginSettings, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(err).NotTo(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
