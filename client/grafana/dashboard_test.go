//go:build unitTests
// +build unitTests

package grafana_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
)

var _ = Describe("Dashboard", func() {
	const fakeCookie = "test_cookie"

	var client *Client
	var server *ghttp.Server
	var statusCode int
	var returnString = ""
	var testDashboardContent = `{
		"title": "Home redirect",
		"version": 0
	}`
	Describe("CreateDashboard()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		Describe("When Pass", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				returnString = "{}"

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/dashboards/db"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns resp", func() {
					dashboard, err := client.CreateDashboard(testDashboardContent, 1, fakeOrgIDInt, true, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					Expect(dashboard).NotTo(BeNil())
					receivedRequests := server.ReceivedRequests()
					Expect(receivedRequests).To(HaveLen(1))
				})
			})
		})

		Describe("When Fail", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				returnString = "{}"

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/dashboards/db"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the dashbord can't be send", func() {
				BeforeEach(func() {
					statusCode = http.StatusNotFound
				})
				It("Returns error and nil", func() {
					dashboard, err := client.CreateDashboard(testDashboardContent, 1, fakeOrgIDInt, true, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(dashboard).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
			Context("When no response", func() {
				BeforeEach(func() {
					server.Close()
				})
				It("Returns error and nil", func() {
					dashboard, err := client.CreateDashboard(testDashboardContent, 1, fakeOrgIDInt, true, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(dashboard).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(0))
				})
			})
			Context("When response can't be parsed", func() {
				BeforeEach(func() {
					returnString = "<xml></xml>"
				})
				It("Returns error and nil", func() {
					dashboard, err := client.CreateDashboard(testDashboardContent, 1, fakeOrgIDInt, true, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(dashboard).To(BeNil())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})

	Describe("GetDashboardByUID()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		const (
			dashboardUID = "eb-dash-88"
		)
		BeforeEach(func() {
			returnString = "{}"
			client, server = ClientWithMockServer()
			statusCode = http.StatusOK
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", fmt.Sprintf("/api/dashboards/uid/%s", dashboardUID)),
				ghttp.RespondWithPtr(&statusCode, &returnString),
			))
		})
		AfterEach(func() {
			server.Close()
		})
		Context("When the response is successful", func() {
			It("Returns nil", func() {
				_, err := client.GetDashboardByUID(dashboardUID, fakeOrgIDInt, fakeCookie)
				Expect(err).NotTo(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
		Context("When dashboard not found", func() {
			BeforeEach(func() {
				statusCode = http.StatusNotFound
			})
			It("Returns nil", func() {
				tmp, err := client.GetDashboardByUID(dashboardUID, fakeOrgIDInt, fakeCookie)
				Expect(err).NotTo(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(tmp).To(BeNil())
			})
		})
		Context("When no response", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("Returns error", func() {
				_, err := client.GetDashboardByUID(dashboardUID, fakeOrgIDInt, fakeCookie)
				Expect(err).To(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})
	})

	Describe("DeleteDashboard()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		const (
			dashboardID = 88
		)
		BeforeEach(func() {
			returnString = "{}"
			client, server = ClientWithMockServer()
			statusCode = http.StatusOK
			server.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("DELETE", fmt.Sprintf("/api/dashboards/uid/eb-dash-%d", dashboardID)),
				ghttp.RespondWithPtr(&statusCode, &returnString),
			))
		})
		AfterEach(func() {
			server.Close()
		})
		Context("When the response is successful", func() {
			It("Returns nil", func() {
				err := client.DeleteDashboard(dashboardID, fakeOrgIDInt, fakeCookie)
				Expect(err).NotTo(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
		Context("When dashboard not found", func() {
			BeforeEach(func() {
				statusCode = http.StatusNotFound
			})
			It("Returns error", func() {
				err := client.DeleteDashboard(dashboardID, fakeOrgIDInt, fakeCookie)
				Expect(err).To(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})
		})
		Context("When no response", func() {
			BeforeEach(func() {
				server.Close()
			})
			It("Returns error", func() {
				err := client.DeleteDashboard(dashboardID, fakeOrgIDInt, fakeCookie)
				Expect(err).To(HaveOccurred())
				Expect(server.ReceivedRequests()).To(HaveLen(0))
			})
		})
	})

	Describe("GetFolders()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		Describe("When success", func() {
			BeforeEach(func() {
				returnString = `[			    
						{			
							"id": 7103,
							"uid": "9M8VsNEGk",
							"title": "Dashboard-playground"		
						}						
				]`
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/folders")),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns nil", func() {
					_, err := client.GetFolders(fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("When GetFolders fail", func() {
			var folder *grafanaModel.Folder
			BeforeEach(func() {
				returnString = `{
					"message": "Unauthorized"
				}`
				client, server = ClientWithMockServer()
				statusCode = http.StatusUnauthorized
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", fmt.Sprintf("/api/folders")),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response failed", func() {
				It("Returns nil", func() {
					folders, err := client.GetFolders(fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(folders).To(BeNil())
				})
			})

			Context("When the response cannot be parsed", func() {
				BeforeEach(func() {
					statusCode = http.StatusOK
					returnString = ""
				})
				It("Returns error", func() {
					_, err := client.GetFolders(fakeOrgIDInt, fakeCookie)
					err = json.Unmarshal([]byte(returnString), &folder)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unexpected end of JSON input"))
				})
			})
		})
	})

	Describe("CreateFolder()", func() {
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)
		var folder grafanaModel.Folder
		Describe("When success", func() {
			BeforeEach(func() {
				returnString = `{
					"id": 35,
					"uid": "-MgKWCt7k",
					"title": "ABC",
					"url": "/dashboards/f/-MgKWCt7k/abc",
					"hasAcl": false,
					"canSave": true,
					"canEdit": true,
					"canAdmin": true,
					"createdBy": "admin",
					"created": "2021-12-01T06:32:00Z",
					"updatedBy": "admin",
					"updated": "2021-12-01T06:32:00Z",
					"version": 1
				}`
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/api/folders")),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response is successful", func() {
				It("Returns nil", func() {
					_, err := client.CreateFolder(fakeOrgIDInt, fakeCookie, "ABC")
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("When fail", func() {
			BeforeEach(func() {
				returnString = `{
					"message": "a folder or dashboard in the general folder with the same name already exists"
				}`
				client, server = ClientWithMockServer()
				statusCode = http.StatusConflict
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/api/folders")),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})
			Context("When the response failed", func() {
				It("Returns error and nil", func() {
					_, err := client.CreateFolder(fakeOrgIDInt, fakeCookie, "ABC")
					Expect(err).To(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("When the response cannot be parsed", func() {
				BeforeEach(func() {
					statusCode = http.StatusOK
					returnString = ""
				})
				It("Returns error", func() {
					_, err := client.CreateFolder(fakeOrgIDInt, fakeCookie, "ABC")
					err = json.Unmarshal([]byte(returnString), &folder)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unexpected end of JSON input"))
				})
			})
		})
	})
})
