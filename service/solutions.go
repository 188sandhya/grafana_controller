package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type ISolutionsService interface {
	GetSolutions(long, allowedOnly bool, solutionScope string) ([]*model.Solution, error)
}

type SolutionsService struct {
	Provider provider.ISolutionsProvider
	Log      logrus.FieldLogger
}

func (fs *SolutionsService) GetSolutions(long, allowedOnly bool, solutionScope string) ([]*model.Solution, error) {
	return fs.Provider.GetSolutions(long, allowedOnly, solutionScope)
}
