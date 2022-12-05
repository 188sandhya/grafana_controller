// +build unitTests

package service_test

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"

	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Datasource service test", func() {
	var mockController *gomock.Controller
	var mockIDatasourceProvider *provider.MockIDatasourceProvider
	logger, logHook := logrustest.NewNullLogger()

	var dataSourceService service.DatasourceService

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIDatasourceProvider = provider.NewMockIDatasourceProvider(mockController)

		dataSourceService = service.DatasourceService{Provider: mockIDatasourceProvider, Log: logger}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("GetDatasourcesByOrganizationID(orgID int64)", func() {
		orgID := int64(2)
		dataSourceList := []*grafanaModel.Datasource{{
			ID:   22,
			Name: "Elastic MCC5",
			Type: "elasticsearch",
		}, {
			ID:   23,
			Name: "Prometheus",
			Type: "prometheus",
		}}
		Context("When provider cannot list Datasources by organization", func() {
			It("Should return an error", func() {
				mockIDatasourceProvider.EXPECT().GetDatasourcesByOrganizationID(orgID).Times(1).Return(nil, errory.ProviderErrors.New("cannot find datasource"))

				foundDatasourceList, err := dataSourceService.GetDatasourcesByOrganizationID(orgID)

				Expect(err).To(HaveOccurred())
				Expect(foundDatasourceList).To(BeNil())
			})
		})
		Context("When provider lists Datasources by organization", func() {
			It("Should list proper list of Datasources", func() {
				mockIDatasourceProvider.EXPECT().GetDatasourcesByOrganizationID(orgID).Times(1).Return(dataSourceList, nil)

				foundDatasourceList, err := dataSourceService.GetDatasourcesByOrganizationID(orgID)

				Expect(err).NotTo(HaveOccurred())
				Expect(foundDatasourceList).NotTo(BeNil())
				Expect(len(foundDatasourceList)).To(Equal(2))
			})
		})
	})

	Describe("GetDatasourcesByID(orgID int64)", func() {
		dsID := int64(2)
		var expErr error
		datasource := &grafanaModel.Datasource{
			ID:   22,
			Name: "Elastic MCC5",
			Type: "elasticsearch",
		}
		Context("When provider returns an error", func() {
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("cannot find datasource")
			})
			It("Should return an error", func() {
				mockIDatasourceProvider.EXPECT().GetDatasourceByID(dsID).Times(1).Return(nil, errory.ProviderErrors.New("cannot find datasource"))

				foundDatasourceList, err := dataSourceService.GetDatasourceByID(dsID)

				assertions.AssertErr(err, expErr)
				Expect(foundDatasourceList).To(BeNil())
			})
		})

		Context("When provider returns datasource", func() {
			It("Should return datasource", func() {
				mockIDatasourceProvider.EXPECT().GetDatasourceByID(dsID).Times(1).Return(datasource, nil)

				foundDatasource, err := dataSourceService.GetDatasourceByID(dsID)

				Expect(err).NotTo(HaveOccurred())
				Expect(foundDatasource).To(Equal(datasource))
			})
		})
	})
})
