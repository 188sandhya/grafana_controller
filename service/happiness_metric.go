package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IHappinessMetricService interface {
	Create(userContext *auth.UserContext, metric *model.HappinessMetric) error
	Update(userContext *auth.UserContext, metric *model.HappinessMetric) error
	Delete(userContext *auth.UserContext, id int64) error
	Get(id int64) (*model.HappinessMetric, error)
	GetAllHappinessMetricsForUser(orgID, userID int64) ([]*model.HappinessMetric, error)
	GetAllHappinessMetricsForTeam(orgID int64) ([]*model.HappinessMetric, error)
	SaveTeamAverage(orgID int64) (int64, error)
	GetUsersMissingInput(orgID int64) ([]*model.UserMissingInput, error)
}

type HappinessMetricService struct {
	Provider provider.IHappinessMetricProvider
	Log      logrus.FieldLogger
}

func (s *HappinessMetricService) Create(userContext *auth.UserContext, metric *model.HappinessMetric) error {
	return s.Provider.CreateHappinessMetric(metric)
}

func (s *HappinessMetricService) Update(userContext *auth.UserContext, metric *model.HappinessMetric) error {
	return s.Provider.UpdateHappinessMetric(metric)
}

func (s *HappinessMetricService) Delete(userContext *auth.UserContext, id int64) error {
	return s.Provider.DeleteHappinessMetric(id)
}

func (s *HappinessMetricService) Get(id int64) (*model.HappinessMetric, error) {
	return s.Provider.GetHappinessMetric(id)
}

func (s *HappinessMetricService) GetAllHappinessMetricsForUser(orgID, userID int64) ([]*model.HappinessMetric, error) {
	return s.Provider.GetAllHappinessMetricsForUser(orgID, userID)
}

func (s *HappinessMetricService) GetAllHappinessMetricsForTeam(orgID int64) ([]*model.HappinessMetric, error) {
	return s.Provider.GetAllHappinessMetricsForTeam(orgID)
}

func (s *HappinessMetricService) SaveTeamAverage(orgID int64) (int64, error) {
	return s.Provider.SaveTeamAverage(orgID)
}

func (s *HappinessMetricService) GetUsersMissingInput(orgID int64) ([]*model.UserMissingInput, error) {
	return s.Provider.GetUsersMissingInput(orgID)
}
