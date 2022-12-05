// +build unitTests

package elastic_test

import (
	"fmt"
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/elastic"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Client", func() {
	logger, _ := logrustest.NewNullLogger()

	Describe("New(baseURL string)", func() {
		Context("when baseURL is unparseable", func() {
			It("should return error", func() {
				notParsingURL := `ś://nourl`
				newClient, err := New(notParsingURL, logger)
				Expect(newClient).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
			})
		})
		Context("when baseURL is parseable", func() {
			It("should return elastic client", func() {
				newClient, err := New("http://elastic.url", logger)
				Expect(newClient).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("GetIndices()", func() {
		var server *ghttp.Server
		var client *Client
		var statusCode int
		var returnString string
		var clientErr error
		BeforeEach(func() {
			server = ghttp.NewServer()
			client, clientErr = New(server.URL(), logger)

		})
		AfterEach(func() {
			server.Close()
		})
		Context("successful scenario", func() {

			BeforeEach(func() {
				returnString = `[
					{
						"health": "yellow",
						"status": "open",
						"index": "docker-loadbalancer-demo-2019.08.28",
						"uuid": "HJgP-iuETyeE7XGUowWBxA",
						"pri": "5",
						"rep": "1",
						"docs.count": "36",
						"docs.deleted": "0",
						"store.size": "1.3mb",
						"pri.store.size": "1.3mb"
					}
				]`
				statusCode = http.StatusOK
				Expect(clientErr).NotTo(HaveOccurred())
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/_cat/indices"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})

			Context("when server responds with list of indices", func() {
				It("parses it into indices model and returns it", func() {
					indices, err := client.GetIndices()
					Expect(err).NotTo(HaveOccurred())
					Expect(indices).To(HaveLen(1))
				})
			})
		})

		Context("unsuccessful scenarios", func() {
			Context("when response has status code other than statusOK", func() {
				BeforeEach(func() {
					returnString = `[
					]`
					statusCode = http.StatusBadRequest
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/_cat/indices"),
						ghttp.RespondWithPtr(&statusCode, &returnString),
					))
				})
				It("returns error ", func() {
					indices, err := client.GetIndices()
					Expect(indices).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
				})

			})

			Context("when response body cannot be read", func() {
				BeforeEach(func() {
					returnString = `[
				]`
					//handler with erraneous body
					handler := func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("Content-Length", "1")
					}
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/_cat/indices"),
						handler,
					))
				})
				It("returns error ", func() {
					indices, err := client.GetIndices()
					Expect(indices).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
				})
			})

			Context("when server responds with data unable to be parsed by index model", func() {
				BeforeEach(func() {
					returnString = `[
						{
				"nutthing valueble to read"
						}
					]`
					statusCode = http.StatusOK
					Expect(clientErr).NotTo(HaveOccurred())
					server.AppendHandlers(ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/_cat/indices"),
						ghttp.RespondWithPtr(&statusCode, &returnString),
					))
				})
				It("parses it into indices model and returns it", func() {
					indices, err := client.GetIndices()
					Expect(indices).To(BeNil())
					Expect(err).To(HaveOccurred())
					Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
				})
			})
		})
		Describe("CreateIndex(name, mapping string)", func() {
			var server *ghttp.Server
			var client *Client
			var statusCode int
			var returnString string
			var clientErr error
			const name = "testName"
			const mapping = `elasticUrl: "http://elastic:9200"
								elasticTemplate: '{
								"name": "jira or other stuff",
								"type": "elasticsearch",
								"url": "jira or other stuff",
								"access": "proxy",
								"basicAuth": false,
								"isDefault": false,
								"database": "database",
								"jsonData": {
									"keepCookies": [],
									"timeField": "@timestamp",
									"esVersion": 56,
									"maxConcurrentShardRequests": 256
								},
								"readOnly": false
								}'`
			BeforeEach(func() {
				server = ghttp.NewServer()
				client, clientErr = New(server.URL(), logger)
				Expect(clientErr).NotTo(HaveOccurred())
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/%s", name)),
					ghttp.VerifyBody([]byte(mapping)),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})

			Context("successful scenario", func() {
				Context("when response returns status ok", func() {
					BeforeEach(func() {
						statusCode = http.StatusOK
					})
					It("returns does not return an error", func() {
						err := client.CreateIndex(name, mapping)
						Expect(err).To(BeNil())
					})
				})
			})
			Context("unsuccessful scenarios", func() {
				Context("when response status code is different than status ok", func() {
					BeforeEach(func() {
						statusCode = http.StatusNotFound
					})
					It("returns elasticClientError", func() {
						err := client.CreateIndex(name, mapping)
						Expect(err).To(HaveOccurred())
						Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
					})
				})
			})
		})

		Describe("DeleteSloHistory(id)", func() {
			var server *ghttp.Server
			var client *Client
			var statusCode int
			var returnString string
			var clientErr error
			id := int64(11)
			query := fmt.Sprintf(`{
				"query": {
					"bool": {
						"filter": [
							{
								"term": {
									"id": %d
								}
							},
							{
								"range": {
								  "date": {
									"gte": "now-14d/d",
									"lte": "now/d"
								  }
								}
							}
						]
					}
				}
			}`, id)

			BeforeEach(func() {
				server = ghttp.NewServer()
				client, clientErr = New(server.URL(), logger)
				Expect(clientErr).NotTo(HaveOccurred())
				server.AppendHandlers(ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", fmt.Sprintf("/%s", "oma-datadog-slo-*/_delete_by_query")),
					ghttp.VerifyJSON(query),
					ghttp.VerifyContentType("application/json"),
					ghttp.RespondWithPtr(&statusCode, &returnString),
				))
			})
			AfterEach(func() {
				server.Close()
			})

			Context("successful scenario", func() {
				Context("when response returns status ok", func() {
					BeforeEach(func() {
						statusCode = http.StatusOK
					})
					It("returns does not return an error", func() {
						err := client.DeleteSloHistory(id)
						Expect(err).To(BeNil())
					})
				})
			})
			Context("unsuccessful scenarios", func() {
				Context("when response status code is different than status ok", func() {
					BeforeEach(func() {
						statusCode = http.StatusNotFound
					})
					It("returns elasticClientError", func() {
						err := client.DeleteSloHistory(id)
						Expect(err).To(HaveOccurred())
						Expect(errory.IsOfType(err, errory.ElasticClientErrors)).To(BeTrue())
					})
				})
			})
		})
	})
})
