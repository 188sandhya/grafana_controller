package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IHealthService interface {
	CheckHealth() error
}

type HealthService struct {
	Provider provider.IHealthProvider
	Log      logrus.FieldLogger
}

func (p *HealthService) CheckHealth() error {
	return p.Provider.CheckHealth()
}
