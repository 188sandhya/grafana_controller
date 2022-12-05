package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IProductsStatusService interface {
	GetProductsStatus() ([]*model.ProductStatus, error)
}

type ProductsStatusService struct {
	Provider provider.IProductsStatusProvider
	Log      logrus.FieldLogger
}

func (ps *ProductsStatusService) GetProductsStatus() ([]*model.ProductStatus, error) {
	return ps.Provider.GetProductsStatus()
}
