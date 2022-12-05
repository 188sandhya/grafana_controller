package service

import (
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IDatasourceService interface {
	GetDatasourcesByOrganizationID(id int64) ([]*grafanaModel.Datasource, error)
	GetDatasourceByID(id int64) (*grafanaModel.Datasource, error)
}

type DatasourceService struct {
	Provider provider.IDatasourceProvider
	Log      logrus.FieldLogger
}

func (o *DatasourceService) GetDatasourcesByOrganizationID(id int64) ([]*grafanaModel.Datasource, error) {
	return o.Provider.GetDatasourcesByOrganizationID(id)
}

func (o *DatasourceService) GetDatasourceByID(id int64) (*grafanaModel.Datasource, error) {
	return o.Provider.GetDatasourceByID(id)
}
