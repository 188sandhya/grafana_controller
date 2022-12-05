package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
)

func (c *Client) CreateDatasource(datasource string, orgID int64, cookie string) (*model.DatasourceID, error) {
	r, err := c.httpPost("api/datasources", orgID, nil, strings.NewReader(datasource), cookie)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	var result model.DatasourceID
	err = json.Unmarshal(r.Body, &result)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}
	return &result, nil
}

func (c *Client) UpdateDatasource(dsID int64, datasource string, orgID int64, cookie string) error {
	r, err := c.httpPut(fmt.Sprintf("api/datasources/%d", dsID), orgID, nil, strings.NewReader(datasource), cookie)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errory.GrafanaClientErrors.New(string(r.Body))
	}

	return nil
}
