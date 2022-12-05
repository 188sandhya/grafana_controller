//go:build unitTests
// +build unitTests

package service_test

import (
	"io/ioutil"
	"path/filepath"

	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

const (
	cookie         = "test cookie"
	folderID int64 = 20
)

var _ = Describe("Dashboard service test", func() {
	var mockController *gomock.Controller
	var mockIGrafana *grafana.MockIClient
	var mockIDatasourceProvider *provider.MockIDatasourceProvider
	logger, logHook := logrustest.NewNullLogger()

	var dashboardService service.DashboardService

	var overwrite bool
	var slo model.Slo
	var userContext auth.UserContext
	var folderID int64
	var folders []*grafanaModel.Folder

	var (
		resultDashboardFilename string
		resultDashboard         []byte
		err                     error
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIDatasourceProvider = provider.NewMockIDatasourceProvider(mockController)
		mockIGrafana = grafana.NewMockIClient(mockController)
		dashboardService = service.DashboardService{
			DatasourceProvider: mockIDatasourceProvider,
			Grafana:            mockIGrafana,
			Log:                logger,
			ResourcePath:       "../resource/",
		}
		logHook.Reset()

		userContext = auth.UserContext{
			ID:     3,
			Cookie: "test cookie",
		}

	})

	AfterEach(func() {
		err = nil
		mockController.Finish()
	})

	Describe("CreateDashboard()", func() {
		Describe("Creating new dashboard for slo with compliance and success rate", func() {
			BeforeEach(func() {
				overwrite = false
				slo = model.Slo{
					ID:                              1,
					OrgID:                           2,
					Name:                            "TestSLO",
					SuccessRateExpectedAvailability: "11.2",
					ComplianceExpectedAvailability:  "22.2",
				}
				folderID = 20

				folders = []*grafanaModel.Folder{
					{
						ID: 35,
						UID: "-MgKWCt7k",
						Title: "ABC",
					},
					{
						ID: 20,
						UID: "Ay0iyUt7k",
						Title: "SLOs",
					},
				}
				
			})

			Context("When during creation of new datadog metric dashboard, dashboard is rendered and created in grafana", func() {
				BeforeEach(func() {
					slo = model.Slo{
						ID:                              1,
						OrgID:                           2,
						DatasourceID:                    2,
						ExternalID:                      "ALAMAKOTA",
						ExternalType:                    model.ExternalSloTypeMetric,
						Name:                            "TestSLO",
						SuccessRateExpectedAvailability: "11.2",
						ComplianceExpectedAvailability:  "0.0",
					}

				})
				JustBeforeEach(func() {
					resultDashboardFilename = "result-create-slo-dashboard-dd-metric"
					resultDashboard, err = ioutil.ReadFile(filepath.Join("testdata", resultDashboardFilename+".golden"))
					Expect(err).ToNot(HaveOccurred())
					mockIDatasourceProvider.EXPECT().GetDatasourceByID(int64(2)).Times(1).Return(&grafanaModel.Datasource{URL: "https://pf-metrodr.datadoghq.com//api/v1"}, nil)
					mockIGrafana.EXPECT().GetFolders(slo.OrgID, cookie).Return(folders, nil)

					mockIGrafana.EXPECT().CreateDashboard(gomock.Any(), folderID, slo.OrgID, overwrite, cookie).
						DoAndReturn(func(dashboard string, folderID, orgID int64, overwrite bool, cookie string) (*grafanaModel.DashboardIDDTO, error) {
							err = ioutil.WriteFile(filepath.Join("testdata", resultDashboardFilename+".testresult"), []byte(dashboard), 0644)
							Expect(err).ToNot(HaveOccurred())
							Expect(dashboard).To(MatchJSON(resultDashboard))
							return nil, nil
						})
				})
				It("should return no error", func() {
					err := dashboardService.CreateDashboard(&userContext, &slo, overwrite)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("When during creation of new compliance metric only dashboard, template and metrics are loaded, dashboard is rendered and created in grafana", func() {
				BeforeEach(func() {
					slo = model.Slo{
						ID:                              1,
						OrgID:                           2,
						Name:                            "TestSLO",
						SuccessRateExpectedAvailability: "0.0",
						ComplianceExpectedAvailability:  "22.2",
						DatasourceID:                    2,
						ExternalID:                      "ALAMAKOTA",
						ExternalType:                    model.ExternalSloTypeMonitor,
					}
				})
				JustBeforeEach(func() {
					resultDashboardFilename = "result-create-slo-dashboard-dd-monitor"
					resultDashboard, err = ioutil.ReadFile(filepath.Join("testdata", resultDashboardFilename+".golden"))
					Expect(err).ToNot(HaveOccurred())
					mockIDatasourceProvider.EXPECT().GetDatasourceByID(int64(2)).Times(1).Return(&grafanaModel.Datasource{URL: "https://pf-metrodr.datadoghq.com//api/v1"}, nil)
					mockIGrafana.EXPECT().GetFolders(slo.OrgID, cookie).Return([]*grafanaModel.Folder{}, nil)
					mockIGrafana.EXPECT().CreateFolder(slo.OrgID, cookie, "SLOs").Return(&grafanaModel.Folder{
						ID: 20,
						UID: "UJCA-Cpnz",
						Title: "SLOs",
						URL: "/dashboards/f/UJCA-Cpnz/slos",
				}, nil)
					mockIGrafana.EXPECT().CreateDashboard(gomock.Any(), folderID, slo.OrgID, overwrite, cookie).
						DoAndReturn(func(dashboard string, folderID, orgID int64, overwrite bool, cookie string) (*grafanaModel.DashboardIDDTO, error) {
							err = ioutil.WriteFile(filepath.Join("testdata", resultDashboardFilename+".testresult"), []byte(dashboard), 0644)
							Expect(err).ToNot(HaveOccurred())
							Expect(dashboard).To(MatchJSON(resultDashboard))
							return nil, nil
						})
				})
				It("should return no error", func() {
					err := dashboardService.CreateDashboard(&userContext, &slo, overwrite)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("DeleteDashboard()", func() {
		sloIdToDelete := int64(2)
		orgIdToDelete := int64(3)
		Context("When Grafana client correctly deletes the dashboard", func() {
			It("Should not return an error", func() {
				mockIGrafana.EXPECT().DeleteDashboard(sloIdToDelete, orgIdToDelete, cookie).Times(1).Return(nil)
				err := dashboardService.DeleteDashboard(&userContext, sloIdToDelete, orgIdToDelete)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("GetTags()", func() {
		BeforeEach(func() {
			slo = model.Slo{}
		})
		It("Should return correct tags", func() {
			tags, err := service.GetTags(&slo)
			Expect(err).ToNot(HaveOccurred())
			Expect(tags).To(Equal(`["OMA","SLO"]`))
		})
	})
})
