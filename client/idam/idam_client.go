package idam

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/sirupsen/logrus"
)

const LogfieldEvent = "event"

// IDAMRestClient retrieves and caches configuration info from IDAM,
// such as public keys in the form of JWKs
// nolint // code copied from webhooks, don't want to change naming
type IDAMRestClient struct {
	idamBaseURL string
	JWKMap      JWKMap
	logger      logrus.FieldLogger
	clientID    string
}

// IDAMClient retrieves and internally caches JWKs on UpdateJWKs()
type IIDAMClient interface {
	UpdateJWKs() error
	TokenAuthenticator(tokenString string) (*StandardAndIdamClaims, bool, string)
	Startup()
}

// NewIDAMClient createas a new IDAMRestClient
func NewIDAMClient(baseURL string, logger logrus.FieldLogger, ownClientID string) *IDAMRestClient {
	return &IDAMRestClient{
		idamBaseURL: baseURL,
		JWKMap:      JWKMap{},
		logger:      logger,
		clientID:    ownClientID,
	}
}

// UpdateJWKs refreshes the key map for the IDAM client
// This method is intended to be called a single time at application startup
// and can additionally be called when a token signature verification fails
// to make sure the token is up to date.
//nolint:gosec
func (c *IDAMRestClient) UpdateJWKs() error {
	idamJWKUrl := fmt.Sprintf("%s/.well-known/openid-configuration/jwks", c.idamBaseURL)
	resp, err := http.Get(idamJWKUrl)
	if err != nil {
		return fmt.Errorf("could not make request to IDAM's jwks endpoint: %s", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not retrieve JWKs from IDAM: IDAM responded with status %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body from IDAM's jwk endpoint: %s", err)
	}

	updatedMap, err := NewJWKMapFromJSON(bodyBytes)
	if err != nil {
		return fmt.Errorf("could not parse the JWKs received from IDAM: %s", err)
	}

	c.JWKMap = updatedMap

	return nil
}

// Startup retrives the IDAM JWK list from IDAM.
// If the retrieval fails, we fallback to a known set of JWKs
func (c *IDAMRestClient) Startup() {
	if err := c.UpdateJWKs(); err != nil {
		c.logger.
			WithField(LogfieldEvent, "idam-retrieve-jwks-failed").
			WithError(err).
			Error("could not retrieve public keys from IDAM")

		c.JWKMap = IDAMFallbackJWKMap

		c.logger.
			WithField(LogfieldEvent, "idam-jwks-fallback-used").
			WithField("idam-jwk-map", c.JWKMap).
			Warn("using fallback for IDAM JWKS, because could not retrieve JWKs")
	} else {
		c.logger.
			WithField(LogfieldEvent, "idam-retrieved-jwks").
			WithField("idam-jwk-map", c.JWKMap).
			Info("successfully retrieved public keys from IDAM")
	}
}
