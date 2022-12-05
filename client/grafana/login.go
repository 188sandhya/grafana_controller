package grafana

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
)

func (c *Client) Login(username, password string) (string, error) {
	bodyString := fmt.Sprintf(`{
		"user": "%s", 
		"password": "%s"
	}`, username, password)
	r, err := c.httpPost("login", 0, nil, strings.NewReader(bodyString), "")
	if err != nil {
		return "", err
	}

	if r.StatusCode != http.StatusOK {
		return "", errory.GrafanaClientErrors.New(string(r.Body))
	}

	token := r.SessionToken
	if token == "" {
		return "", errory.GrafanaClientErrors.New("session does not exist")
	}
	return token, nil
}
