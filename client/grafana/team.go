package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
)

const createTeamAPIPath = "/api/teams/"
const searchTeamAPIPath = createTeamAPIPath + "search"

func (c *Client) CreateTeam(name string, orgID int64, cookie string) (*model.CreateTeamResponse, error) {
	body := fmt.Sprintf(`{
		"name": "%s"
	}`, name)
	r, err := c.httpPost(createTeamAPIPath, orgID, map[string]string{"orgId": strconv.FormatInt(orgID, 10)}, strings.NewReader(body), cookie)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	var result model.CreateTeamResponse
	err = json.Unmarshal(r.Body, &result)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	return &result, nil
}

func (c *Client) GetTeam(name string, orgID int64, cookie string) (*model.Team, error) {
	r, err := c.httpGet(searchTeamAPIPath, orgID, map[string]string{"name": name, "orgId": strconv.FormatInt(orgID, 10)}, cookie)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	var result model.TeamResult
	err = json.Unmarshal(r.Body, &result)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	if len(result.Teams) == 0 {
		return nil, nil
	}

	return &result.Teams[0], nil
}
