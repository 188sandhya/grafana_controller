// +build unitTests

package grafana_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

const FakeTeamName = "TestTeam"

var _ = Describe("Team", func() {
	Describe("GetTeam()", func() {
		const fakeCookie = "test_cookie"

		var client *Client
		var server *ghttp.Server
		var statusCode int
		var returnedTeamResult interface{}
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)

		Describe("When Pass", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				teamResultJsonResponse := `{
					"totalCount": 0,
					"teams": [],
					"page": 1,
					"perPage": 1000
				}`
				query := fmt.Sprintf("name=%s&orgId=%d", FakeTeamName, fakeOrgIDInt)
				err := json.Unmarshal([]byte(teamResultJsonResponse), &returnedTeamResult)
				Expect(err).NotTo(HaveOccurred())

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/teams/search", query),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &returnedTeamResult),
				))
			})
			AfterEach(func() {
				server.Close()
			})

			Context("When the response is successful and team is not found", func() {
				It("Returns false and no error", func() {
					_, err := client.GetTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("When the response is successful and the team is returned", func() {
				BeforeEach(func() {
					teamResultJsonResponse := `{
						"totalCount": 1,
						"teams": [
							{
								"id": 1,
								"orgId": 1,
								"name": "TestTeam",
								"email": "",
								"avatarUrl": "/avatar/689e35bb18383390f6f113b9fbc94fbe",
								"memberCount": 0,
								"permission": 0
							}
						],
						"page": 1,
						"perPage": 1000
					}`
					err := json.Unmarshal([]byte(teamResultJsonResponse), &returnedTeamResult)
					Expect(err).NotTo(HaveOccurred())
				})
				It("Returns true and no error", func() {
					_, err := client.GetTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})

		Describe("When Fail", func() {
			BeforeEach(func() {
				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				returnedTeamResult = model.TeamResult{}
				query := fmt.Sprintf("name=%s&orgId=%d", FakeTeamName, fakeOrgIDInt)

				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/teams/search", query),
					ghttp.RespondWithJSONEncodedPtr(&statusCode, &returnedTeamResult),
				))
			})
			AfterEach(func() {
				server.Close()
			})

			Context("When the response is not OK", func() {
				BeforeEach(func() {
					statusCode = http.StatusForbidden
				})
				It("Returns false and error", func() {
					_, err := client.GetTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})

			Context("When the response cannot be parsed", func() {
				BeforeEach(func() {
					returnedTeamResult = `{}`
				})
				It("Returns false and  error", func() {
					_, err := client.GetTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
					assertions.AssertErr(err, errory.GrafanaClientErrors.New(""))
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})
			})
		})
	})

	Describe("CreateTeam()", func() {
		const fakeCookie = "test_cookie"

		var client *Client
		var server *ghttp.Server
		var statusCode int
		var request string
		var fakeOrgIDInt, _ = strconv.ParseInt(FakeOrgID, 10, 64)

		var returnString string

		Describe("During creation of team", func() {
			BeforeEach(func() {
				returnString = `{
					"message": "team create",
					"teamId": 1
				}`

				client, server = ClientWithMockServer()
				statusCode = http.StatusOK
				//DO NOT FORMAT THIS STRING IT MATCHES ORGINAL BODY STRUCT
				request = fmt.Sprintf(`{
		"name": "%s"
	}`, FakeTeamName)
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/teams", fmt.Sprintf("orgId=%d", fakeOrgIDInt)),
					ghttp.VerifyBody([]byte(request)),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})

			Context("When the response is successful", func() {
				It("Returns no error", func() {
					_, err := client.CreateTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("When the response is not successful", func() {
				BeforeEach(func() {
					statusCode = http.StatusForbidden
				})
				It("Returns an error", func() {
					_, err := client.CreateTeam(FakeTeamName, fakeOrgIDInt, fakeCookie)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
