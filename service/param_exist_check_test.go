// +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

const idToFind = int64(32)

var _ = Describe("ParamExistCheck service test", func() {
	var mockController *gomock.Controller
	var mockIOrgProvider *provider.MockIOrganizationProvider
	var mockISLOProvider *provider.MockISLOProvider
	logger, _ := logrustest.NewNullLogger()

	var paramExistCheckService service.ParamExistCheckService

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIOrgProvider = provider.NewMockIOrganizationProvider(mockController)
		mockISLOProvider = provider.NewMockISLOProvider(mockController)
		paramExistCheckService = service.ParamExistCheckService{OrgProvider: mockIOrgProvider,
			SloProvider: mockISLOProvider,
			Log:         logger,
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("CheckOrgByID(id int64) error", func() {
		Context("When org can be find by id", func() {
			BeforeEach(func() {
				mockIOrgProvider.EXPECT().GetOrganizationByID(idToFind).Times(1).Return(&grafanaModel.Organization{ID: idToFind}, nil)
			})
			It("should not return an error", func() {
				err := paramExistCheckService.CheckOrgByID(idToFind)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When org provider returns errory", func() {
			BeforeEach(func() {
				mockIOrgProvider.EXPECT().GetOrganizationByID(idToFind).Times(1).Return(&grafanaModel.Organization{ID: idToFind}, errory.ProviderErrors.Builder().WithPayload("ID", idToFind).Create())
			})
			It("should return an error", func() {
				err := paramExistCheckService.CheckOrgByID(idToFind)
				assertions.AssertErr(err, errory.NotFoundErrors.Builder().WithPayload("ID", idToFind).Create())
			})
		})
		Context("When org provider returns empty org", func() {
			BeforeEach(func() {
				mockIOrgProvider.EXPECT().GetOrganizationByID(idToFind).Times(1).Return(nil, nil)
			})
			It("should return an error", func() {
				err := paramExistCheckService.CheckOrgByID(idToFind)
				assertions.AssertErr(err, errory.NotFoundErrors.Builder().WithPayload("ID", idToFind).Create())
			})
		})
	})

	Describe("CheckSLOByID(id int64) error", func() {
		Context("When SLO can be find by id", func() {
			BeforeEach(func() {
				mockISLOProvider.EXPECT().GetSlo(idToFind).Times(1).Return(&model.Slo{ID: idToFind}, nil)
			})
			It("should not return an error", func() {
				err := paramExistCheckService.CheckSloByID(idToFind)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("When SLO provider returns an error", func() {
			BeforeEach(func() {
				mockISLOProvider.EXPECT().GetSlo(idToFind).Times(1).Return(&model.Slo{ID: idToFind}, errory.ProviderErrors.New("expected error"))
			})
			It("should return an error", func() {
				err := paramExistCheckService.CheckSloByID(idToFind)
				assertions.AssertErr(err, errory.NotFoundErrors.Builder().WithPayload("ID", idToFind).Create())
			})
		})

		Context("When SLO provider returns an nil", func() {
			BeforeEach(func() {
				mockISLOProvider.EXPECT().GetSlo(idToFind).Times(1).Return(nil, nil)
			})
			It("should return an error", func() {
				err := paramExistCheckService.CheckSloByID(idToFind)
				assertions.AssertErr(err, errory.NotFoundErrors.Builder().WithPayload("ID", idToFind).Create())
			})
		})
	})

})
