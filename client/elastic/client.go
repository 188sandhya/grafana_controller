package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"net/url"
	"time"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/elastic"
	"github.com/sirupsen/logrus"
)

type IClient interface {
	GetIndices() ([]*model.Index, error)
	CreateIndex(name, mapping string) error
	DeleteSloHistory(id int64) error
}

type Client struct {
	baseURL string
	*http.Client
	Log logrus.FieldLogger
}

func New(baseURL string, log logrus.FieldLogger) (*Client, error) {
	_, err := url.Parse(baseURL)
	if err != nil {
		return nil, errory.ElasticClientErrors.Wrap(err)
	}

	return &Client{
		baseURL,
		&http.Client{
			Timeout: 30 * time.Second,
		},
		log,
	}, nil
}

func (c *Client) GetIndices() ([]*model.Index, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/_cat/indices?format=json", c.baseURL), nil)
	if err != nil {
		return nil, errory.ElasticClientErrors.Wrap(err)
	}

	r, err := c.Do(req)
	if err != nil {
		return nil, errory.ElasticClientErrors.Wrap(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, errory.ElasticClientErrors.Builder().WithPayload("code", r.StatusCode).Create()
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errory.ElasticClientErrors.Wrap(err)
	}

	indices := []*model.Index{}
	if err := json.Unmarshal(data, &indices); err != nil {
		return nil, errory.ElasticClientErrors.Wrap(err)
	}
	return indices, nil
}

func (c *Client) CreateIndex(name, mapping string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", c.baseURL, name), strings.NewReader(mapping))
	if err != nil {
		return errory.ElasticClientErrors.Wrap(err)
	}

	req.Header.Add("Content-Type", "application/json")
	r, err := c.Do(req)
	if err != nil {
		return errory.ElasticClientErrors.Wrap(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return errory.ElasticClientErrors.Builder().
			WithPayload("code", r.StatusCode).WithPayload("index", name).
			Create()
	}

	return nil
}

func (c *Client) DeleteSloHistory(id int64) error {
	query := fmt.Sprintf(`{
		"query": {
			"bool": {
				"filter": [
					{
						"term": {
							"id": %d
						}
					},
					{
						"range": {
						  "date": {
							"gte": "now-14d/d",
							"lte": "now/d"
						  }
						}
					}
				]
			}
		}
	}`, id)

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/oma-datadog-slo-*/_delete_by_query", bytes.NewBuffer([]byte(query)))
	if err != nil {
		c.Log.WithError(err).Error("could not send request")
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	r, err := c.Do(req)
	if err != nil {
		return errory.ElasticClientErrors.Wrap(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		c.Log.Errorf("unexpected request status code:%v", r.StatusCode)
		return errory.ElasticClientErrors.Builder().
			WithPayload("code", r.StatusCode).WithPayload("id", id).
			Create()
	}

	return nil
}
