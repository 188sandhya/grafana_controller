package idam

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

const (
	RoleAllVerticalsFullAccess     = "2TR_ALL_VERTICALS_FULL_ACCESS"
	RoleSpecificVerticalFullAccess = "2TR_VERTICAL_FULL_ACCESS"
	RoleOMASuperAdmin              = "OMA_ADMIN"
	RoleOMAViewAll                 = "OMA_VIEW_ALL"
	ContextCustomer                = "2tr_customer"
	ContextVertical                = "vertical"
	UserTypeEmployee               = "EMP"
	UserTypeClient                 = "CLIENT"
)

// StandardAndIdamClaims combines both the jwt standard claims with the custom
// IDAMClaims used in the Authorization package
type StandardAndIdamClaims struct {
	jwt.StandardClaims
	Authorization     Roles  `json:"authorization"`
	UserType          string `json:"userType,omitempty"`
	UserPrincipalName string `json:"upn,omitempty"`
	Realm             string `json:"realm,omitempty"`
}

// Roles is a list of Role elements, enriched with searching receiver methods
type Roles []Role
type Role map[string]ContextSet
type ContextSet []ContextCombination
type ContextCombination map[string]ContextValues
type ContextValues []string

// getRole
func (r Roles) GetRole(desiredRole string) (found bool, role ContextSet, warning string) {
	if len(r) == 0 {
		return false, nil, warning
	}

	for _, role := range r {
		found, context, _ := role.Matches(desiredRole)
		if found {
			if len(context) != 0 {
				warning = fmt.Sprintf("We found a role that should be context-less, such as an admin role. "+
					"However, it contains some context. See role: %s", desiredRole)
			}

			return true, context, warning
		}
	}

	return false, nil, warning
}

// Matches checks whether the current role matches the desired role
func (r Role) Matches(desiredRole string) (found bool, contextSet ContextSet, warning string) {
	if context, ok := r[desiredRole]; ok {
		return true, context, warning
	}

	return false, ContextSet{}, warning
}

func (c *IDAMRestClient) TokenAuthenticator(tokenString string) (claims *StandardAndIdamClaims, ok bool, warning string) {
	token, err := jwt.ParseWithClaims(tokenString, &StandardAndIdamClaims{}, c.prepareSingatureVerification)

	if token != nil && token.Valid {
		// Authentication for technical users (= IDAM clients)
		if token.Claims.(*StandardAndIdamClaims).UserType == UserTypeClient {
			return token.Claims.(*StandardAndIdamClaims), true, ""
		}

		// Authentication for real users
		if token.Claims.(*StandardAndIdamClaims).UserType == UserTypeEmployee {
			if token.Claims.(*StandardAndIdamClaims).Audience == c.clientID || c.audienceCheckDisabled() {
				return token.Claims.(*StandardAndIdamClaims), true, ""
			}

			return nil, false, "the token provided was not issued for this service"
		}

		return nil, false, "the token provided has incorrect UserType"
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		switch {
		case ve.Errors&jwt.ValidationErrorMalformed != 0:
			return nil, false, "the token provided is malformed"
		case ve.Errors&jwt.ValidationErrorExpired != 0:
			return nil, false, "the token provided has expired"
		case ve.Errors&(jwt.ValidationErrorNotValidYet|jwt.ValidationErrorIssuedAt) != 0:
			return nil, false, "the token provided is not valid yet or was issued in the future"
		case ve.Errors&(jwt.ValidationErrorSignatureInvalid) != 0:
			return nil, false, "the token signature is invalid"
		}
	}

	return nil, false, fmt.Sprintf("The token could not be validated: %s", err)
}

func (c *IDAMRestClient) prepareSingatureVerification(token *jwt.Token) (interface{}, error) {
	kid, err := c.kidFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("could not extract kid: %s", err)
	}
	publicKey, err := c.publicKeyPEMforKID(kid)
	if err != nil {
		return nil, fmt.Errorf("could not get public key: %s", err)
	}

	rsaPublicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return rsaPublicKey, nil
}

func (c *IDAMRestClient) kidFromToken(token *jwt.Token) (string, error) {
	kid, ok := token.Header["kid"]
	if !ok {
		return "", fmt.Errorf("the token header does not contain a kid")
	}

	kidString, ok := kid.(string)
	if !ok {
		return "", fmt.Errorf("the kid is not a string")
	}

	return kidString, nil
}

func (c *IDAMRestClient) publicKeyPEMforKID(kid string) ([]byte, error) {
	publicKey, ok := c.JWKMap[kid]
	if !ok {
		_ = c.UpdateJWKs()
		publicKey, ok = c.JWKMap[kid]
		if !ok {
			return nil, fmt.Errorf("no public key found for kid '%s'", kid)
		}
	}

	publicKeyString, err := publicKey.PEMPublicKeyString()
	if err != nil {
		return nil, fmt.Errorf("could not covert publicKey to PEM for kid '%s'", kid)
	}

	return []byte(publicKeyString), nil
}

func (c *IDAMRestClient) audienceCheckDisabled() bool {
	return c.clientID == ""
}
