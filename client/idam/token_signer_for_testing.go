// +build unitTests integrationTests journeyTests

package idam

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/onsi/gomega"
)

// Only to be used from tests
func CreateAndSignToken(kid string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, StandardAndIdamClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    "ds-test-setup",
			Audience:  "ds-prod",
			IssuedAt:  time.Now().Add(-1 * time.Minute).Unix(),
		},
		UserType: "EMP",
	})
	token.Header["kid"] = kid

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(TokenIssuerPrivateKey))
	Expect(err).ToNot(HaveOccurred())

	signedTokenString, err := token.SignedString(key)
	Expect(err).ToNot(HaveOccurred())

	return signedTokenString
}

func CreateAndSignTokenWithoutKID() string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, StandardAndIdamClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    "ds-test-setup",
			Audience:  "ds-prod",
			IssuedAt:  time.Now().Add(-1 * time.Minute).Unix(),
		},
		UserType: "EMP",
	})
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(TokenIssuerPrivateKey))
	Expect(err).ToNot(HaveOccurred())

	signedTokenString, err := token.SignedString(key)
	Expect(err).ToNot(HaveOccurred())

	return signedTokenString
}

// Only to be used from tests
func CreateAndSignTokenForEmployeeWithAuthorization(kid string, roles Roles) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, StandardAndIdamClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
			Issuer:    "ds-test-setup",
			Audience:  "ds-prod",
			IssuedAt:  time.Now().Add(-1 * time.Minute).Unix(),
		},
		Authorization: roles,
		UserType:      "EMP",
	})
	token.Header["kid"] = kid

	key, _ := jwt.ParseRSAPrivateKeyFromPEM([]byte(TokenIssuerPrivateKey))

	signedTokenString, _ := token.SignedString(key)

	return signedTokenString
}

// WARNING: This key pair was created solely for test purposes.
// Considering the private key is contained here in plain text
// it should never be used outside of testing purposes!
var (
	TokenIssuerPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCjCZ6BN+DyQws/Q+ABPLLM1GA+fLPp9KKByh0d7KmCfc4R+jFS
R9FVuCg5YU9N2ZdwC6HRYvQmcnqYzPxclz19MoU+bdxIFo9WJNPYBxTuESJuREwD
FvjhiJviYT0w3vy07IBrVZewb3m6V5yWN61Uqq0szV4ALABbaxMQ0nicyQIDAQAB
AoGAGPWTB3M3g78RzLimZWoWcVcd+NL8dBeYfUgk1vzxImICFyx3OoJ2IKpVthsY
mfFyxptxRW3htLUX4aaYB9C7f9zkY1EpmeeSOlYZHjbAQ8akg7MYH43WywZN8jv0
Oh3gPuOPqlQqYbSBeztvxe3hQCgF3QOopAPyJaNEwhV0M9UCQQDtoj2bKAemoUcj
+/pWQU2KB5y7ccbroMMfTrXkAy3C1xu+ZZEKoEAzJFUu0bCCkZh3DbX3RyQxUz6d
/60mqvoHAkEAr6NwImrkXzAsIk7wtA4R/7XlEGwkshwb7w5XE9wwUkGyfyeJcAXb
r4HTKPiWNYZW2I54QHk0myqcHUMEsiY+rwJAcqnRfjePkYjKsgNZJRu3lX3c09mv
uWzGGio5vD8IaravDW0m0nDG6aaDb+cAe9BTOEcmYZ4zSZW4ZjbDzx+7KwJAZvvi
9RtN+o5JYnh85GZXoWLrA90VCyY2Ls5uumNiJekFm074ZCnbLSZnRN+1W38AjwvC
cLNg6BZs4S95omeQWwJBAJ0qe0F0KHxi1mpwYl6A9+1JLCmIP4ORrqDolfaPbHS3
yLwZkLvS42QdLwcGAmw9xp8HwM1LamvszeGhkTBkF0c=
-----END RSA PRIVATE KEY-----`

	TokenIssuerPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCjCZ6BN+DyQws/Q+ABPLLM1GA+
fLPp9KKByh0d7KmCfc4R+jFSR9FVuCg5YU9N2ZdwC6HRYvQmcnqYzPxclz19MoU+
bdxIFo9WJNPYBxTuESJuREwDFvjhiJviYT0w3vy07IBrVZewb3m6V5yWN61Uqq0s
zV4ALABbaxMQ0nicyQIDAQAB
-----END PUBLIC KEY-----`

	TokenIssuerJWT = MakeTokenIssuerJWTWithKid("jwk-for-testing")
)

func MakeTokenIssuerJWTWithKid(kid string) string {
	return fmt.Sprintf(`
			{
				"kty": "RSA",
				"n": "owmegTfg8kMLP0PgATyyzNRgPnyz6fSigcodHeypgn3OEfoxUkfRVbgoOWFPTdmXcAuh0WL0JnJ6mMz8XJc9fTKFPm3cSBaPViTT2AcU7hEibkRMAxb44Yib4mE9MN78tOyAa1WXsG95ulecljetVKqtLM1eACwAW2sTENJ4nMk",
				"e": "AQAB",
				"kid": "%s"
			}
		`, kid)
}
