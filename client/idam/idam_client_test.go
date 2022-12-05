// +build unitTests

package idam_test

import (
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("An idamClient to provide general idam info, such as the public keys", func() {
	var server *ghttp.Server
	var client *idam.IDAMRestClient
	var err error
	logger, loggerHook := test.NewNullLogger()
	clientID := "deployment-service"

	BeforeEach(func() {
		server = ghttp.NewServer()
		client = idam.NewIDAMClient(server.URL(), logger, clientID)
		loggerHook.Reset()
	})

	AfterEach(func() {
		//shut down the server between tests
		server.Close()
	})

	Describe("updating JWKs", func() {
		Context("when the request is successful", func() {
			Context("and the payload contains valid keys", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
							ghttp.RespondWith(http.StatusOK, `{
						"keys":[{ 
								"kty":"RSA",
								"e":"AQAB",
								"use":"sig",
								"kid":"token-signing-keypair",
								"alg":"RS256",
								"n":"xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw"
							}]
						}`),
						),
					)

					err = client.UpdateJWKs()
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("calls the correct endpoint", func() {
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("should update the JWKMap correctly", func() {
					Expect(client.JWKMap).To(HaveKeyWithValue("token-signing-keypair",
						idam.JWK{
							KTY: "RSA",
							E:   "AQAB",
							KID: "token-signing-keypair",
							N:   "xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw",
						}))
				})
			})

			Context("and the payload contains invalid JWKs", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
							ghttp.RespondWith(http.StatusOK, `{ "keys":[{ "kty":"RSA" }] }`),
						),
					)

					err = client.UpdateJWKs()
				})

				It("calls the correct endpoint", func() {
					Expect(server.ReceivedRequests()).To(HaveLen(1))
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("could not parse the JWKs received from IDAM:"))
				})
			})
		})

		Context("when the request is not successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				err = client.UpdateJWKs()
			})

			It("calls the correct endpoint", func() {
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("could not retrieve JWKs from IDAM: IDAM responded with status 500"))
			})
		})
	})

	Describe("at application startup", func() {
		Context("when retrieving the token is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
						ghttp.RespondWith(http.StatusOK, `{
						"keys":[{ 
								"kty":"RSA",
								"e":"AQAB",
								"use":"sig",
								"kid":"token-signing-keypair",
								"alg":"RS256",
								"n":"xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw"
							}]
						}`),
					),
				)

				client.Startup()
			})

			It("calls the correct endpoint", func() {
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("should update the JWKMap correctly", func() {
				Expect(client.JWKMap).To(HaveKeyWithValue("token-signing-keypair",
					idam.JWK{
						KTY: "RSA",
						E:   "AQAB",
						KID: "token-signing-keypair",
						N:   "xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw",
					}))
			})

			It("logs the received tokens", func() {
				Expect(loggerHook.Entries).To(HaveLen(1))
				Expect(loggerHook.Entries[0].Level).To(Equal(logrus.InfoLevel))
				Expect(loggerHook.Entries[0].Data).To(HaveKeyWithValue("event", "idam-retrieved-jwks"))
				Expect(loggerHook.Entries[0].Data).To(HaveKey("idam-jwk-map"))
				Expect(loggerHook.Entries[0].Data["idam-jwk-map"]).ToNot(BeNil())
			})
		})

		Context("when the request is not successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
						ghttp.RespondWith(http.StatusInternalServerError, `{}`),
					),
				)

				client.Startup()
			})

			It("calls the correct endpoint", func() {
				Expect(server.ReceivedRequests()).To(HaveLen(1))
			})

			It("creates two log entries", func() {
				Expect(loggerHook.Entries).To(HaveLen(2))
			})

			It("logs the errror", func() {
				Expect(loggerHook.Entries[0].Level).To(Equal(logrus.ErrorLevel))
				Expect(loggerHook.Entries[0].Data["error"]).
					To(MatchError("could not retrieve JWKs from IDAM: IDAM responded with status 500"))
				Expect(loggerHook.Entries[0].Data).To(HaveKeyWithValue("event", "idam-retrieve-jwks-failed"))
			})

			It("loads the fallback data", func() {
				Expect(client.JWKMap).To(HaveKeyWithValue("token-signing-keypair",
					idam.JWK{
						KTY: "RSA",
						E:   "AQAB",
						KID: "token-signing-keypair",
						N:   "xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw",
					}))
			})

			It("logs a warning about using fallback tokens", func() {
				Expect(loggerHook.Entries[1].Level).To(Equal(logrus.WarnLevel))
				Expect(loggerHook.Entries[1].Data).To(HaveKeyWithValue("event", "idam-jwks-fallback-used"))
				Expect(loggerHook.Entries[1].Data).To(HaveKey("idam-jwk-map"))
				Expect(loggerHook.Entries[1].Data["idam-jwk-map"]).ToNot(BeNil())
			})
		})
	})
})
