// +build unitTests

package idam_test

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("retrieving public keys and veriying token signatures", func() {
	var (
		idamClient *idam.IDAMRestClient
	)

	Describe("when the server has our own public key in the jwks", func() {
		BeforeEach(func() {
			idamClient, _, _ = IDAMClientWithFakeServer("ds-prod")
			err := idamClient.UpdateJWKs()
			Expect(err).ToNot(HaveOccurred())
		})

		It("is possible to verify the self-signed token with the RequestAuthenticator", func() {
			token := idam.CreateAndSignToken("jwk-for-testing")
			claims, success, msg := idamClient.TokenAuthenticator(token)
			Expect(msg).To(Equal(""))
			Expect(claims).NotTo(BeNil())
			Expect(success).To(BeTrue())
		})
	})

	Describe("when the server has IDAMs public key in the jwks", func() {
		BeforeEach(func() {
			idamClient, _, _ = IDAMClientWithFakeServerAndRealJWK("", "token-signing-keypair")
			err := idamClient.UpdateJWKs()
			Expect(err).ToNot(HaveOccurred())
		})

		It("considers a token invalid, when it was not signed by the right authority", func() {
			// we are now pretending to be IDAM by specifying a kid which points
			// to an actual IDAM public key. However instead we're singing the token
			// ourselves with our testing key pair

			token := idam.CreateAndSignToken("token-signing-keypair")
			claims, success, msg := idamClient.TokenAuthenticator(token)
			Expect(claims).To(BeNil())
			Expect(success).To(BeFalse())
			Expect(msg).To(Equal("the token signature is invalid"))
		})
	})
})
