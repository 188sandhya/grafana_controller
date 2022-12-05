// +build unitTests

package grafana_test

import (
	"net/http"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Login", func() {
	var client *Client
	var server *ghttp.Server
	var statusCode int
	var username, password string
	var returnString = "{}"

	Describe("Login()", func() {
		Describe("When Pass", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/login"),
					ghttp.RespondWithPtr(&statusCode, &returnString, http.Header{"Set-Cookie": []string{"grafana_session=0a4a1b72e81d7285fc7a7be63c45ad49"}}),
				))
				username = "admin"
				password = "admin"
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns no error", func() {
					cookie, err := client.Login(username, password)
					Expect(err).NotTo(HaveOccurred())
					Expect(cookie).To(Equal("0a4a1b72e81d7285fc7a7be63c45ad49"))
					receivedRequests := server.ReceivedRequests()
					Expect(receivedRequests).To(HaveLen(1))
				})
			})
		})

		Describe("When Fail", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/login"),
					ghttp.RespondWithPtr(&statusCode, nil),
				))
				username = "admin"
				password = "admin"
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When http.StatusBadRequest response is received", func() {
				BeforeEach(func() {
					statusCode = http.StatusBadRequest
				})
				It("Returns error and nil", func() {
					_, err := client.Login(username, password)
					Expect(err).To(HaveOccurred())
					Expect(err).NotTo(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
