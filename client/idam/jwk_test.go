// +build unitTests

package idam_test

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSON Web Keys", func() {
	Context("when validation a JWK", func() {
		Describe("and e is missing", func() {
			jwk := idam.JWK{
				N:   "some-n-string",
				KTY: "RSA",
				KID: "some-id",
			}

			It("has a proper error message", func() {
				Expect(jwk.Validate()).To(MatchError("jwk is invalid: e is missing"))
			})
		})

		Describe("and n is missing", func() {
			jwk := idam.JWK{
				E:   "AQAB",
				KTY: "RSA",
				KID: "some-id",
			}

			It("has a proper error message", func() {
				Expect(jwk.Validate()).To(MatchError("jwk is invalid: n is missing"))
			})
		})

		Describe("and kty is missing", func() {
			jwk := idam.JWK{
				E:   "AQAB",
				N:   "some-n-string",
				KID: "some-id",
			}

			It("has a proper error message", func() {
				Expect(jwk.Validate()).To(MatchError("jwk is invalid: kty is missing"))
			})
		})

		Describe("and kid is missing", func() {
			jwk := idam.JWK{
				E:   "AQAB",
				N:   "some-n-string",
				KTY: "RSA",
			}

			It("has a proper error message", func() {
				Expect(jwk.Validate()).To(MatchError("jwk is invalid: kid is missing"))
			})
		})

		Describe("and multiple values are missing", func() {
			jwk := idam.JWK{
				E: "AQAB",
				N: "some-n-string",
			}

			It("returns an error", func() {
				Expect(jwk.Validate()).To(HaveOccurred())
			})

			It("has a proper error message mentioning all missing fields", func() {
				Expect(jwk.Validate().Error()).To(ContainSubstring("jwk is invalid"))
				Expect(jwk.Validate().Error()).To(ContainSubstring("kid is missing"))
				Expect(jwk.Validate().Error()).To(ContainSubstring("kty is missing"))
			})
		})

		Describe("and all fields are set", func() {
			jwk := idam.JWK{
				E:   "AQAB",
				N:   "some-n-string",
				KID: "some-id",
				KTY: "RSA",
			}

			It("returns no error", func() {
				Expect(jwk.Validate()).ToNot(HaveOccurred())
			})
		})
	})

	Context("when converting a jwk to a pem public key", func() {
		Describe("and the kty is not rsa", func() {
			It("returns an error", func() {
				jwk := idam.JWK{KTY: "HSA"}

				_, err := jwk.PEMPublicKeyString()

				Expect(err).To(MatchError("invalid key type: HSA"))
			})
		})

		Describe("and the e is of a non-standard value", func() {
			It("returns an error", func() {
				jwk := idam.JWK{KTY: "RSA", E: "AAAAA"}

				_, err := jwk.PEMPublicKeyString()

				Expect(err).To(MatchError("unrecognized value for e: AAAAA"))
			})
		})

		Describe("and all data is valid", func() {
			var (
				err       error
				pemString string
			)

			// Note: this is an actual JWK from IDAM
			jwk := idam.JWK{
				KTY: "RSA",
				E:   "AQAB",
				KID: "some-id",
				N:   "xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw",
			}

			BeforeEach(func() {
				pemString, err = jwk.PEMPublicKeyString()
			})

			It("returns no error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the correct pem public key", func() {
				// Note: This is IDAM's public key
				expectedPEM := `-----BEGIN RSA PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAxBagyHGKzi0LZeu0l1WZ
FR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9m
Y25lIGE/oxwyC1vE/xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA
3eEtl3zJhQ/fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb/DInjr749hGV2a/
rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buh
badlkPWdd7jQEQKyaJlM1Za7o4s29N+UBfyCel10RDhpXv4f4vj8JhpaYONZXPV/
vwIDAQAB
-----END RSA PUBLIC KEY-----
`

				Expect(pemString).To(Equal(expectedPEM))
			})
		})
	})

	Context("when parsing a JWK map from json", func() {
		Describe("and the json is invalid", func() {
			input := []byte(`{`)

			It("returns an error", func() {
				_, err := idam.NewJWKMapFromJSON(input)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("and no keys are present", func() {
			input := []byte(`{ "keys": [] }`)

			It("returns an error", func() {
				_, err := idam.NewJWKMapFromJSON(input)
				Expect(err).To(MatchError("the provided jwk json does not contain any keys"))
			})
		})

		Describe("and there are only valid keys present", func() {
			var (
				err    error
				jwkMap idam.JWKMap
			)

			kid := "some token id"
			input := []byte(`{ "keys": [{
				"n": "some-n-string",
				"e": "AQAB",
				"kid": "` + kid + `",
				"kty": "RSA"
			}] }`)

			BeforeEach(func() {
				jwkMap, err = idam.NewJWKMapFromJSON(input)
			})

			It("returns no error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("has the correct key", func() {
				Expect(jwkMap).To(HaveKeyWithValue(kid, idam.JWK{
					N:   "some-n-string",
					E:   "AQAB",
					KID: kid,
					KTY: "RSA",
				}))
			})
		})

		Describe("and there are invalid keys present", func() {
			var (
				err error
			)

			kid := "some token id"
			input := []byte(`{ "keys": [{
				"n": "some-n-string",
				"e": "AQAB",
				"kid": "` + kid + `",
				"kty": "RSA"
			}, 
			{
				"e": "AQAB"
			}] }`)

			BeforeEach(func() {
				_, err = idam.NewJWKMapFromJSON(input)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("invalid jwk at position 1: jwk is invalid: n is missing; kty is missing; kid is missing"))
			})
		})
	})
})
