// +build unitTests

package idam_test

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("An idamClient", func() {

	var _ = Describe("configured with a non-empty client id", func() {
		Describe("JWT authentication function", func() {
			var (
				customizedIdamClient *idam.IDAMRestClient
				server               *ghttp.Server
			)

			clientName := "ds-prod"

			BeforeEach(func() {
				customizedIdamClient, server, _ = IDAMClientWithFakeServer(clientName)
				err := customizedIdamClient.UpdateJWKs()
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				server.Close()
			})

			Describe("when no Authorization header is present", func() {
				var (
					authenticated bool
					explanation   string
					claims        jwt.Claims
				)

				BeforeEach(func() {
					claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator("")
				})

				It("returns that the request is not authenticated", func() {
					Expect(authenticated).To(BeFalse())
				})

				It("returns no claims", func() {
					Expect(claims).To(BeNil())
				})

				It("returns an appropriate explanation message", func() {
					Expect(explanation).To(Equal("the token provided is malformed"))
				})
			})

			Describe("when an Authorization header is present but it is not for bearer authentication", func() {
				var (
					authenticated bool
					explanation   string
					claims        jwt.Claims
				)

				BeforeEach(func() {
					claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator("abc123")
				})

				It("returns that the request is not authenticated", func() {
					Expect(authenticated).To(BeFalse())
				})

				It("returns no claims", func() {
					Expect(claims).To(BeNil())
				})

				It("returns an appropriate explanation message", func() {
					Expect(explanation).To(Equal("the token provided is malformed"))
				})
			})

			Describe("when an Authorization header is present and contains a token", func() {
				Describe("with a valid kid", func() {
					var token string

					BeforeEach(func() {
						token = idam.CreateAndSignToken("jwk-for-testing")
					})

					Describe("and the token has expired", func() {
						var (
							authenticated bool
							explanation   string
							claims        jwt.Claims
						)

						BeforeEach(func() {
							jwt.TimeFunc = func() time.Time {
								return time.Now().Add(20 * time.Minute)
							}

							claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator(token)
						})

						It("returns that the request is not authenticated", func() {
							Expect(authenticated).To(BeFalse())
						})

						It("returns no claims", func() {
							Expect(claims).To(BeNil())
						})

						It("returns an appropriate explanation message", func() {
							Expect(explanation).To(Equal("the token provided has expired"))
						})

						AfterEach(func() {
							jwt.TimeFunc = func() time.Time {
								return time.Now()
							}
						})
					})

					Describe("and the token is not valid yet", func() {
						var (
							authenticated bool
							explanation   string
							claims        jwt.Claims
						)

						BeforeEach(func() {
							jwt.TimeFunc = func() time.Time {
								return time.Now().Add(-10 * time.Minute)
							}

							claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator(token)
						})

						It("returns that the request is not authenticated", func() {
							Expect(authenticated).To(BeFalse())
						})

						It("returns no claims", func() {
							Expect(claims).To(BeNil())
						})

						It("returns an appropriate explanation message", func() {
							Expect(explanation).To(Equal("the token provided is not valid yet or was issued in the future"))
						})

						AfterEach(func() {
							jwt.TimeFunc = func() time.Time {
								return time.Now()
							}
						})
					})

					Describe("and the token is within the validity period", func() {
						BeforeEach(func() {
							jwt.TimeFunc = func() time.Time {
								return time.Now()
							}
						})

						Describe("and the audience in the token does not match our client", func() {
							var (
								authenticated                 bool
								explanation                   string
								claims                        jwt.Claims
								idamClientWithDifferentClient *idam.IDAMRestClient
							)

							BeforeEach(func() {
								idamClientWithDifferentClient, _, _ = IDAMClientWithFakeServer("we-are-someone-else-now")
								err := idamClientWithDifferentClient.UpdateJWKs()
								Expect(err).ToNot(HaveOccurred())
								claims, authenticated, explanation = idamClientWithDifferentClient.TokenAuthenticator(token)
							})

							It("returns that the request is not authenticated", func() {
								Expect(authenticated).To(BeFalse())
							})

							It("returns no claims", func() {
								Expect(claims).To(BeNil())
							})

							It("returns an appropriate explanation message", func() {
								Expect(explanation).To(Equal("the token provided was not issued for this service"))
							})
						})

						Describe("and the audience in the token matches our client", func() {
							var (
								authenticated bool
								explanation   string
								claims        jwt.Claims
							)

							BeforeEach(func() {
								claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator(token)
							})

							It("returns that the request is authenticated", func() {
								Expect(authenticated).To(BeTrue())
							})

							It("returns claims", func() {
								Expect(claims).ToNot(BeNil())
							})

							It("does not return any explanation message", func() {
								Expect(explanation).To(BeEmpty())
							})
						})
					})
				})

				Describe("with no kid at all", func() {
					var (
						token         string
						authenticated bool
						explanation   string
						claims        jwt.Claims
					)
					BeforeEach(func() {
						token = idam.CreateAndSignTokenWithoutKID()
						jwt.TimeFunc = func() time.Time {
							return time.Now()
						}

						claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator(token)
					})

					It("returns that the request is not authenticated", func() {
						Expect(authenticated).To(BeFalse())
					})

					It("returns no claims", func() {
						Expect(claims).To(BeNil())
					})

					It("returns an appropriate explanation message", func() {
						Expect(explanation).To(Equal("The token could not be validated: could not extract kid: the token header does not contain a kid"))
					})
				})

				Describe("with a kid pointing to a key we don't know about", func() {
					var (
						token         string
						authenticated bool
						explanation   string
						claims        jwt.Claims
					)

					makeIdamClientRequestAndAssign := func(idamClient *idam.IDAMRestClient, kid string) {
						token = idam.CreateAndSignToken(kid)

						jwt.TimeFunc = func() time.Time {
							return time.Now()
						}

						claims, authenticated, explanation = idamClient.TokenAuthenticator(token)
					}
					BeforeEach(func() {
						anotherKid := "another-kid-known-to-idam-a-bit-later"
						customizedIdamClient, server, _ = IDAMClientWithFakeServerWithKid(clientName, anotherKid)
						err := customizedIdamClient.UpdateJWKs()
						Expect(err).ToNot(HaveOccurred())
					})

					Describe("and idam does now know this kid", func() {
						BeforeEach(func() {
							anotherKid := "another-kid-known-to-idam-a-bit-later"
							makeIdamClientRequestAndAssign(customizedIdamClient, anotherKid)
						})

						It("returns that the request is authenticated", func() {
							Expect(authenticated).To(BeTrue())
						})

						It("returns claims", func() {
							Expect(claims).ToNot(BeNil())
						})

						It("returns an appropriate explanation message", func() {
							Expect(explanation).To(Equal(""))
						})

						It("triggers an update for JWKs once more", func() {
							Expect(server.ReceivedRequests()).To(HaveLen(2))
						})

					})

					Describe("and idam still doesn't know this kid", func() {
						BeforeEach(func() {
							makeIdamClientRequestAndAssign(customizedIdamClient, "wrong-kid")
						})
						It("returns that the request is not authenticated", func() {
							Expect(authenticated).To(BeFalse())
						})

						It("returns no claims", func() {
							Expect(claims).To(BeNil())
						})

						It("returns an appropriate explanation message", func() {
							Expect(explanation).To(Equal("The token could not be validated: could not get public key: no public key found for kid 'wrong-kid'"))
						})

						It("triggers an update for JWKs once more", func() {
							Expect(server.ReceivedRequests()).To(HaveLen(2))
						})
					})
				})
			})

			Describe("when an Authorization header is present and contains a malformed token", func() {
				var (
					authenticated bool
					explanation   string
					claims        jwt.Claims
				)

				BeforeEach(func() {
					claims, authenticated, explanation = customizedIdamClient.TokenAuthenticator("abc123")
				})

				It("returns that the request is not authenticated", func() {
					Expect(authenticated).To(BeFalse())
				})

				It("returns no claims", func() {
					Expect(claims).To(BeNil())
				})

				It("returns an appropriate explanation message", func() {
					Expect(explanation).To(Equal("the token provided is malformed"))
				})
			})
		})
	})

	Describe("configured with an empty client id", func() {
		// Unfortunately IDAM sets the audience filed incorrectly.
		// For some reason, they are setting it to the same value as the subject
		// This means, we unfortunately can't validate the audience field as of now.
		// Thus, we have this feature to ignore audience validation if the client
		// is configured to have an empty client id

		Describe("and the request is otherwise valid", func() {
			var (
				authenticated bool
				explanation   string
				claims        jwt.Claims
				server        *ghttp.Server
				idamClient    *idam.IDAMRestClient
			)

			BeforeEach(func() {
				idamClient, server, _ = IDAMClientWithFakeServer("")
				err := idamClient.UpdateJWKs()
				Expect(err).ToNot(HaveOccurred())
				token := idam.CreateAndSignToken("jwk-for-testing")
				claims, authenticated, explanation = idamClient.TokenAuthenticator(token)
			})

			AfterEach(func() {
				server.Close()
			})

			It("returns claims", func() {
				Expect(explanation).To(BeEmpty())
				Expect(authenticated).To(BeTrue())
				Expect(claims).ToNot(BeNil())
			})
		})
	})
})
