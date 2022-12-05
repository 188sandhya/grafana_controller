package idam

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
)

// JWK is a json web key as specified in https://tools.ietf.org/html/rfc7517
type JWK struct {
	E   string `json:"e"`
	N   string `json:"n"`
	KTY string `json:"kty"`
	KID string `json:"kid"`
}

// JWKMap is a collection of at JWKs identified by its kid
type JWKMap map[string]JWK

// Validate chekcks that all required fields are set
func (jwk JWK) Validate() error {
	var errorStrings []string

	if jwk.E == "" {
		errorStrings = append(errorStrings, "e is missing")
	}

	if jwk.N == "" {
		errorStrings = append(errorStrings, "n is missing")
	}

	if jwk.KTY == "" {
		errorStrings = append(errorStrings, "kty is missing")
	}

	if jwk.KID == "" {
		errorStrings = append(errorStrings, "kid is missing")
	}

	if len(errorStrings) != 0 {
		return fmt.Errorf("jwk is invalid: %s", strings.Join(errorStrings, "; "))
	}

	return nil
}

// PEMPublicKeyString converts a JWK to a PEM public key certificate string
func (jwk JWK) PEMPublicKeyString() (string, error) {
	if jwk.KTY != "RSA" {
		return "", fmt.Errorf("invalid key type: %s", jwk.KTY)
	}

	// decode the base64 bytes for n
	nb, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return "", err
	}

	e := 0
	// The default exponent is usually 65537, so just compare the
	// base64 for [1,0,1] or [0,1,0,1]
	if jwk.E == "AQAB" || jwk.E == "AAEAAQ" {
		e = 65537
	} else {
		// need to decode "e" as a big-endian int
		return "", fmt.Errorf("unrecognized value for e: %s", jwk.E)
	}

	pk := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: e,
	}

	der, err := x509.MarshalPKIXPublicKey(pk)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: der,
	}

	var out bytes.Buffer
	_ = pem.Encode(&out, block)
	return out.String(), nil
}

// NewJWKMapFromJSON takes the jwks json as provided, for example
// by IDAM at https://idam.metrosystems.net/.well-known/openid-configuration/jwks,
// and makes sure all keys contain the required fields
func NewJWKMapFromJSON(jsonBytes []byte) (JWKMap, error) {
	var list jwkList
	jwkMap := JWKMap{}

	if err := json.Unmarshal(jsonBytes, &list); err != nil {
		return jwkMap, err
	}

	if len(list.Keys) == 0 {
		return jwkMap, fmt.Errorf("the provided jwk json does not contain any keys")
	}

	for i, jwk := range list.Keys {
		if err := jwk.Validate(); err != nil {
			return jwkMap, fmt.Errorf("invalid jwk at position %d: %s", i, err)
		}

		jwkMap[jwk.KID] = jwk
	}

	return jwkMap, nil
}

type jwkList struct {
	Keys []JWK `json:"keys"`
}
