// +build unitTests

package grafana_test

import (
	"net/http"
	"strconv"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Datasource", func() {
	const fakeCookie = "test_cookie"

	var client *Client
	var server *ghttp.Server
	var statusCode int
	var returnString = ""
	var testDatasource = `{
		"name":     "test_name",
		"access":   "proxy",
		"database": "test_db",
		"type":     "test_type",
		"url":      "test_url",
		"jsonData": {
			"key1": "val1",
			"key2": "val2"
		}`
	Describe("CreateDatasource()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		BeforeEach(func() {
			client, server = ClientWithMockServer()
			statusCode = http.StatusOK

			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/api/datasources"),
				ghttp.RespondWithPtr(&statusCode, &returnString),
			))
		})
		Describe("When Pass", func() {
			BeforeEach(func() {
				returnString = `{
					"id": 8,
					"name": "test_name"
				}`
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns no error", func() {
					dsID, err := client.CreateDatasource(testDatasource, fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					Expect(*dsID).To(Equal(model.DatasourceID{
						Name: "test_name",
						ID:   8,
					}))
					receivedRequests := server.ReceivedRequests()
					Expect(receivedRequests).To(HaveLen(1))
				})
			})
		})

		Describe("When Fail", func() {
			BeforeEach(func() {
				returnString = "{}"
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When http.StatusBadRequest response is received", func() {
				BeforeEach(func() {
					statusCode = http.StatusBadRequest
				})
				It("Returns error and nil", func() {
					dsID, err := client.CreateDatasource(testDatasource, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(dsID).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("When response can't be parsed", func() {
				BeforeEach(func() {
					returnString = "<xml></xml>"
				})
				It("Returns error and nil", func() {
					dsID, err := client.CreateDatasource(testDatasource, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(dsID).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})
})
