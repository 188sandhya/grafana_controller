package service

import (
	"strings"

	elastic "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/elastic"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	grafana "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"

	"github.com/sirupsen/logrus"
)

type ISloService interface {
	Create(userContext *auth.UserContext, slo *model.Slo) error
	Update(userContext *auth.UserContext, slo *model.Slo) error
	Delete(userContext *auth.UserContext, id int64) error
	Get(id int64) (*model.Slo, error)
	GetDetailedSlos(sloNameQuery, orgNameQuery string) ([]*model.DetailedSlo, error)
	GetByOrgID(orgID int64) ([]*model.Slo, error)
	FindSlos(params *model.SloQueryParams) ([]*model.Slo, error)
	DeleteSloHistory(id int64) error
}

type SloService struct {
	SloProvider      provider.ISLOProvider
	DSProvider       provider.IDatasourceProvider
	DashboardService IDashboardService
	Log              logrus.FieldLogger
	ElasticClient    elastic.IClient
}

func (s *SloService) Create(userContext *auth.UserContext, slo *model.Slo) error {
	containSlo, err := s.SloProvider.ContainSlosWithSameName(slo.OrgID, slo.Name, 0)
	if err != nil {
		return errory.Decorate(err, "slo service create()")
	}
	if containSlo {
		return errory.NotUniqueErrors.Builder().WithPayload("name", slo.Name).Create()
	}

	var ds *grafana.Datasource
	ds, err = s.DSProvider.GetDatasourceByID(slo.DatasourceID)
	if err != nil {
		return errory.Decorate(err, "slo service create()")
	}
	if ds.Type != grafana.DatasourceTypeDatadog {
		return errory.CreateExternalSLOWithWrongDSForbiddenErrors.Builder().WithPayload("datasource", slo.DatasourceID).Create()
	}

	if err = s.SloProvider.CreateSlo(slo); err != nil {
		return errory.Decorate(err, "slo service create()")
	}

	err = s.DashboardService.CreateDashboard(userContext, slo, false)

	return err
}

func (s *SloService) Update(userContext *auth.UserContext, slo *model.Slo) error {
	containSlo, err := s.SloProvider.ContainSlosWithSameName(slo.OrgID, slo.Name, slo.ID)
	if err != nil {
		return err
	}
	if containSlo {
		return errory.NotUniqueErrors.Builder().WithPayload("name", slo.Name).Create()
	}

	if err = s.SloProvider.UpdateSlo(slo); err != nil {
		return err
	}

	err = s.DashboardService.CreateDashboard(userContext, slo, true)

	return err
}

func (s *SloService) Delete(userContext *auth.UserContext, id int64) error {
	slo, err := s.SloProvider.GetSlo(id)
	if err != nil {
		return err
	}

	err = s.DashboardService.DeleteDashboard(userContext, slo.ID, slo.OrgID)
	if err != nil && !strings.Contains(err.Error(), "Dashboard not found") {
		return err
	}

	return s.SloProvider.DeleteSlo(slo)
}

func (s *SloService) DeleteSloHistory(id int64) error {
	return s.ElasticClient.DeleteSloHistory(id)
}

func (s *SloService) Get(id int64) (*model.Slo, error) {
	return s.SloProvider.GetSlo(id)
}

func (s *SloService) GetByOrgID(orgID int64) ([]*model.Slo, error) {
	return s.SloProvider.GetSlosByOrganizationID(orgID)
}

func (s *SloService) GetDetailedSlos(sloNameQuery, orgNameQuery string) ([]*model.DetailedSlo, error) {
	slos, err := s.SloProvider.GetDetailedSlos(sloNameQuery, orgNameQuery)
	return slos, err
}

func (s *SloService) FindSlos(params *model.SloQueryParams) ([]*model.Slo, error) {
	return s.SloProvider.FindSlos(params)
}
