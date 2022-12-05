package service

import (
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IOrganizationService interface {
	GetOrganizationByID(id int64) (*grafanaModel.Organization, error)
}

type OrganizationService struct {
	Provider provider.IOrganizationProvider
	Log      logrus.FieldLogger
}

func (o *OrganizationService) GetOrganizationByID(id int64) (*grafanaModel.Organization, error) {
	return o.Provider.GetOrganizationByID(id)
}
