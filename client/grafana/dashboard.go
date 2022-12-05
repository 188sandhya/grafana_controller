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

func (c *Client) CreateDashboard(dashboard string, folderID, orgID int64, overwrite bool, cookie string) (*model.DashboardIDDTO, error) {
	body := fmt.Sprintf(`{
		"dashboard": %s,
		"folderId": %d,
		"overwrite": %t
	}`, dashboard, folderID, overwrite)

	r, err := c.httpPost("api/dashboards/db", orgID, nil, strings.NewReader(body), cookie)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	var result model.DashboardIDDTO
	err = json.Unmarshal(r.Body, &result)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	return &result, nil
}

func (c *Client) GetDashboardByUID(uid string, orgID int64, cookie string) (json.RawMessage, error) {
	path := fmt.Sprintf("api/dashboards/uid/%s", uid)

	r, err := c.httpGet(path, orgID, nil, cookie)
	if err != nil {
		return nil, err
	}

	if r.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	return r.Body, nil
}

func (c *Client) DeleteDashboard(sloID, orgID int64, cookie string) error {
	path := fmt.Sprintf("api/dashboards/uid/eb-dash-%s", strconv.FormatInt(sloID, 10))

	r, err := c.httpDelete(path, orgID, nil, nil, cookie)
	if err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return errory.GrafanaClientErrors.New(string(r.Body))
	}
	return nil
}

func (c *Client) GetFolders(orgID int64, cookie string) (folders []*model.Folder, err error) {
	r, err := c.httpGet("api/folders", orgID, nil, cookie)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	if err = json.Unmarshal(r.Body, &folders); err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	return folders, nil
}

func (c *Client) CreateFolder(orgID int64, cookie, title string) (*model.Folder, error) {
	body := fmt.Sprintf(`{
		"title": "%s"
	}`, title)

	r, err := c.httpPost("api/folders", orgID, nil, strings.NewReader(body), cookie)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, errory.GrafanaClientErrors.New(string(r.Body))
	}

	var result *model.Folder
	err = json.Unmarshal(r.Body, &result)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	return result, nil
}
