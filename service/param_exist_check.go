package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type Type int

const (
	Slo Type = iota
	Organization
)

type IParamExistCheckService interface {
	CheckOrgByID(id int64) error
	CheckSloByID(id int64) error
}

type ParamExistCheckService struct {
	OrgProvider provider.IOrganizationProvider
	SloProvider provider.ISLOProvider
	Log         logrus.FieldLogger
}

func (e *ParamExistCheckService) CheckOrgByID(id int64) error {
	org, err := e.OrgProvider.GetOrganizationByID(id)
	if err != nil || org == nil {
		err = errory.NotFoundErrors.Builder().Wrap(err).WithPayload("ID", id).Create()
	}
	return err
}

func (e *ParamExistCheckService) CheckSloByID(id int64) error {
	slo, err := e.SloProvider.GetSlo(id)
	if err != nil || slo == nil {
		err = errory.NotFoundErrors.Builder().Wrap(err).WithPayload("ID", id).Create()
	}
	return err
}
