package service

import (
	"encoding/json"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type ISDAService interface {
	GetConfig() ([]map[string]interface{}, error)
	GetSDAFeatureByOrg(orgID int64) (*model.SDAFeatures, error)
}

type SDAService struct {
	Provider provider.ISDAProvider
	Log      logrus.FieldLogger
}

func (s *SDAService) GetConfig() ([]map[string]interface{}, error) {
	var raw []map[string]interface{}

	cfg, err := s.Provider.GetConfig()

	if err != nil {
		return raw, err
	}

	err = json.Unmarshal([]byte(cfg), &raw)

	return raw, err
}

func (s *SDAService) GetSDAFeatureByOrg(orgID int64) (*model.SDAFeatures, error) {
	return s.Provider.GetSDAFeatureByOrg(orgID)
}
