package grafana

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
)

type IClient interface {
	CreateDashboard(dashboard string, folderID, orgID int64, overwrite bool, cookie string) (*model.DashboardIDDTO, error)
	DeleteDashboard(sloID, orgID int64, cookie string) error
	CreateDatasource(datasource string, orgID int64, cookie string) (*model.DatasourceID, error)
	UpdateDatasource(dsID int64, datasource string, orgID int64, cookie string) error
	EnablePlugin(pluginSettings *model.PluginSettings, orgID int64, cookie string) error
	CreateTeam(name string, orgID int64, cookie string) (*model.CreateTeamResponse, error)
	GetTeam(name string, orgID int64, cookie string) (*model.Team, error)
	Login(username, password string) (string, error)
	GetFolders(orgID int64, cookie string) ([]*model.Folder, error)
	CreateFolder(orgID int64, cookie, title string) (*model.Folder, error)
}

type Client struct {
	baseURL string
	*http.Client
}

func New(baseURL string) (*Client, error) {
	_, err := url.Parse(baseURL)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	return &Client{
		baseURL,
		&http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) httpPost(apiPath string, orgID int64, query map[string]string, body io.Reader, cookie string) (*HTTPResponse, error) {
	return c.buildRequestAndDo(http.MethodPost, apiPath, orgID, query, body, cookie)
}

func (c *Client) httpPut(apiPath string, orgID int64, query map[string]string, body io.Reader, cookie string) (*HTTPResponse, error) {
	return c.buildRequestAndDo(http.MethodPut, apiPath, orgID, query, body, cookie)
}

func (c *Client) httpDelete(apiPath string, orgID int64, query map[string]string, body io.Reader, cookie string) (*HTTPResponse, error) {
	return c.buildRequestAndDo(http.MethodDelete, apiPath, orgID, query, body, cookie)
}

func (c *Client) httpGet(apiPath string, orgID int64, query map[string]string, cookie string) (*HTTPResponse, error) {
	return c.buildRequestAndDo(http.MethodGet, apiPath, orgID, query, nil, cookie)
}

func (c *Client) buildRequestAndDo(method, apiPath string, orgID int64, query map[string]string, body io.Reader, cookie string) (*HTTPResponse, error) {
	reqPath, err := c.buildURL(apiPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, reqPath, body)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	if len(query) != 0 {
		q := req.URL.Query()

		for key := range query {
			q.Add(key, query[key])
		}
		req.URL.RawQuery = q.Encode()
	}

	return c.httpRequest(req, orgID, cookie)
}

func (c *Client) httpRequest(request *http.Request, orgID int64, cookie string) (*HTTPResponse, error) {
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")

	// TODO this can be reconfigured in grafana, pass via configuration
	if cookie != "" {
		request.AddCookie(&http.Cookie{Name: "grafana_session", Value: cookie})
	}
	request.Header.Add("X-Grafana-Org-Id", fmt.Sprint(orgID))
	r, err := c.Do(request)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}
	defer r.Body.Close()
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errory.GrafanaClientErrors.Wrap(err)
	}

	if r.StatusCode == http.StatusUnauthorized {
		return nil, errory.GrafanaClientAuthErrors.New("grafana client")
	}

	respCookie := ""
	for _, cur := range r.Cookies() {
		if cur.Name == "grafana_session" {
			respCookie = cur.Value
			break
		}
	}

	return &HTTPResponse{
		Body:         bytes,
		StatusCode:   r.StatusCode,
		SessionToken: respCookie,
	}, nil
}

func (c *Client) buildURL(apiPath string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", errory.GrafanaClientErrors.Wrap(err)
	}
	u.Path = path.Join(u.Path, apiPath)
	return u.String(), nil
}

type HTTPResponse struct {
	Body         []byte
	StatusCode   int
	SessionToken string
}
