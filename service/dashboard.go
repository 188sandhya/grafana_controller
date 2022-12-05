package service

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	"github.com/cbroglie/mustache"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IDashboardService interface {
	DeleteDashboard(userContext *auth.UserContext, sloID, orgID int64) error
	UpdateDashboard(userContext *auth.UserContext, slo *model.Slo) error
	CreateDashboard(userContext *auth.UserContext, slo *model.Slo, overwrite bool) error
}

var urlRegexp = regexp.MustCompile(`^[^/]*(?:/[^/]*){2}`)

type DashboardService struct {
	DatasourceProvider provider.IDatasourceProvider
	Grafana            grafana.IClient
	Log                logrus.FieldLogger
	ResourcePath       string
}

func (d *DashboardService) UpdateDashboard(userContext *auth.UserContext, slo *model.Slo) error {
	return d.CreateDashboard(userContext, slo, true)
}

func (d *DashboardService) CreateDashboard(userContext *auth.UserContext, slo *model.Slo, overwrite bool) error {
	dashboard, err := d.prepareDashboard(slo)
	if err != nil || dashboard == "" {
		return err
	}

	folder, err := d.GetFolderByTitle(slo.OrgID, userContext.Cookie, "SLOs")
	if err != nil {
		return err
	}

	_, err = d.Grafana.CreateDashboard(dashboard, folder.ID, slo.OrgID, overwrite, userContext.Cookie)
	if err != nil {
		return err
	}

	return err
}

func (d *DashboardService) prepareDashboard(slo *model.Slo) (string, error) {
	dashboardTemplate, err := loadDashboardTemplate(slo.ExternalType, d.ResourcePath)
	if err != nil {
		return "", err
	}

	sloName, err := jsonEscape(slo.Name)
	if err != nil {
		return "", err
	}

	tags, err := GetTags(slo)
	if err != nil {
		return "", err
	}

	var datasourceLink string
	datasource, dataerr := d.DatasourceProvider.GetDatasourceByID(slo.DatasourceID)
	if dataerr != nil {
		return "", dataerr
	}
	datasourceLink = urlRegexp.FindString(datasource.URL)

	data, err := mustache.Render(string(dashboardTemplate), map[string]interface{}{
		"ORG_ID":           strconv.FormatInt(slo.OrgID, 10),
		"SLO_SUCCESS_RATE": slo.SuccessRateExpectedAvailability,
		"SLO_COMPLIANCE":   slo.ComplianceExpectedAvailability,
		"SLO_ID":           strconv.FormatInt(slo.ID, 10),
		"DASHBOARD_TITLE":  sloName,
		"TAGS":             tags,
		"DASHBOARD_UID":    "eb-dash-" + strconv.FormatInt(slo.ID, 10),
		"EXTERNAL_ID":      slo.ExternalID,
		"DS_ID":            strconv.FormatInt(slo.DatasourceID, 10),
		"DATADOG_URL":      datasourceLink,
	})

	if err != nil {
		return "", err
	}

	return data, nil
}

func jsonEscape(i string) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", errory.FetchResourceErrors.Wrap(err)
	}
	return string(b[1 : len(b)-1]), nil
}

func loadDashboardTemplate(exType model.ExternalSloType, pathToResourceDir string) (dashboardTemplate []byte, err error) {
	if exType == model.ExternalSloTypeMonitor {
		dashboardTemplate, err = ioutil.ReadFile(pathToResourceDir + "dashboard-template-dd-compliance.json.mustache")
	} else if exType == model.ExternalSloTypeMetric {
		dashboardTemplate, err = ioutil.ReadFile(pathToResourceDir + "dashboard-template-dd-success-rate.json.mustache")
	}
	err = errory.FetchResourceErrors.Wrap(err)
	return
}

func (d *DashboardService) DeleteDashboard(userContext *auth.UserContext, sloID, orgID int64) error {
	return d.Grafana.DeleteDashboard(sloID, orgID, userContext.Cookie)
}

func GetTags(slo *model.Slo) (string, error) {
	tagSlice := []string{"OMA", "SLO"}

	tags, err := json.Marshal(tagSlice)
	if err != nil {
		return "", err
	}
	return string(tags), nil
}

func (d *DashboardService) GetFolderByTitle(orgID int64, cookie, title string) (folder *grafanaModel.Folder, err error) {
	folders, err := d.Grafana.GetFolders(orgID, cookie)
	if err != nil {
		return folder, err
	}

	for _, folder := range folders {
		if folder.Title == "SLOs" {
			return folder, nil
		}
	}

	return d.Grafana.CreateFolder(orgID, cookie, title)
}
