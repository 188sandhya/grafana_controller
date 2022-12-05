package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
)

func (c *Client) EnablePlugin(pluginSettings *model.PluginSettings, orgID int64, cookie string) error {
	pluginSettingsBytes, err := json.Marshal(pluginSettings)
	if err != nil {
		return err
	}

	r, err := c.httpPost(fmt.Sprintf("api/plugins/%s/settings", pluginSettings.Name), orgID, nil, bytes.NewReader(pluginSettingsBytes), cookie)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errory.GrafanaClientErrors.New(string(r.Body))
	}

	return nil
}
