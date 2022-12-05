// +build unitTests

package idam_test

import (
	"fmt"
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	"github.com/onsi/gomega/ghttp"
	"github.com/sirupsen/logrus/hooks/test"
)

func IDAMClientWithFakeServer(clientID string) (*idam.IDAMRestClient, *ghttp.Server, *test.Hook) {
	return IDAMClientWithFakeServerWithKid(clientID, "jwk-for-testing")
}

// Only to be used from within test suites, contains Expectations
// Ideally, to be called from within beforEach
func IDAMClientWithFakeServerWithKid(clientID string, kid string) (*idam.IDAMRestClient, *ghttp.Server, *test.Hook) {
	logger, loggerHook := test.NewNullLogger()
	server := ghttp.NewServer()
	client := idam.NewIDAMClient(server.URL(), logger, clientID)
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
			ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{ "keys":[%s] }`, idam.TokenIssuerJWT)),
		),
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
			ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{ "keys":[%s] }`, idam.MakeTokenIssuerJWTWithKid(kid))),
		),
	)
	return client, server, loggerHook
}

func IDAMClientWithFakeServerAndRealJWK(clientID string, kid string) (*idam.IDAMRestClient, *ghttp.Server, *test.Hook) {
	realIdamJWK := `{"kty":"RSA","e":"AQAB","use":"sig","kid":"token-signing-keypair","alg":"RS256","n":"xBagyHGKzi0LZeu0l1WZFR0oT0rErOYblsUPClBWkgAUdewiDoWFLolfsAy2TMjSyPkttob4N1BwHRcwSp9mY25lIGE_oxwyC1vE_xJbFTuahizkNQ0PnT2p9h4VzeP7lcr1Xc6Fr24eUcUNaMdA3eEtl3zJhQ_fyM0IHBwNGOCXE3dypvT4PkCilX68wLGnmuDKb_DInjr749hGV2a_rozRHvMwiQOwVmT7qzGdnxYXhRoAjlxFwFw8DAkC5LFKnyj8BWFzoMH0HTqE6buhbadlkPWdd7jQEQKyaJlM1Za7o4s29N-UBfyCel10RDhpXv4f4vj8JhpaYONZXPV_vw"}`
	logger, loggerHook := test.NewNullLogger()
	server := ghttp.NewServer()
	client := idam.NewIDAMClient(server.URL(), logger, clientID)
	server.AppendHandlers(
		ghttp.CombineHandlers(
			ghttp.VerifyRequest("GET", "/.well-known/openid-configuration/jwks"),
			ghttp.RespondWith(http.StatusOK, fmt.Sprintf(`{ "keys":[%s] }`, realIdamJWK)),
		),
	)
	return client, server, loggerHook
}
