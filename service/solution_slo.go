package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
)

type ISolutionSloService interface {
	GetSolutionSlo(orgNameQuery string) (*model.SolutionSlo, error)
}

type SolutionSloService struct {
	SloProvider provider.ISolutionSloProvider
}

func (s *SolutionSloService) GetSolutionSlo(orgNameQuery string) (*model.SolutionSlo, error) {
	solutionSlo, err := s.SloProvider.GetSolutionSlo(orgNameQuery)
	if err != nil {
		return solutionSlo, err
	}
	return solutionSlo, nil
}
