//go:build unitTests
// +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	client "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/elastic"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Slo service test", func() {
	var mockController *gomock.Controller
	var mockISLOProvider *provider.MockISLOProvider
	var mockIDatasourceProvider *provider.MockIDatasourceProvider
	var mockDashboardService *service.MockIDashboardService
	var mockElasticClient *client.MockIClient
	logger, logHook := logrustest.NewNullLogger()
	var sloService service.SloService
	var userContext auth.UserContext

	const expectedCookie = "test cookie"

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockISLOProvider = provider.NewMockISLOProvider(mockController)
		mockElasticClient = client.NewMockIClient(mockController)
		mockIDatasourceProvider = provider.NewMockIDatasourceProvider(mockController)
		mockDashboardService = service.NewMockIDashboardService(mockController)

		userContext = auth.UserContext{
			ID:     3,
			Cookie: "test cookie",
		}

		sloService = service.SloService{SloProvider: mockISLOProvider,
			DSProvider:       mockIDatasourceProvider,
			DashboardService: mockDashboardService,
			ElasticClient:    mockElasticClient,
			Log:              logger}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("Create(slo *model.Slo)", func() {
		orgID := int64(2)
		sloID := int64(66)
		slo := model.Slo{
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}
		ds := grafana.Datasource{
			Type: grafana.DatasourceTypeDatadog,
		}
		BeforeEach(func() {
			slo.DatasourceID = 0
		})

		Context("When provider tries to find slo with the same name but returns an error", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, int64(0)).Times(1).Return(false, errory.NotUniqueErrors.New("duplicated slo name"))

				err := sloService.Create(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When provider finds slo with the same name", func() {
			It("Should return proper error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, int64(0)).Times(1).Return(true, nil)

				err := sloService.Create(&userContext, &slo)

				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.NotUniqueErrors)).To(BeTrue())
			})
		})

		Context("When provider tries to create new SLO but fails", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, int64(0)).Times(1).Return(false, nil)
				mockISLOProvider.EXPECT().CreateSlo(&slo).Times(1).Return(errory.ProviderErrors.New("provider create slo error"))
				mockIDatasourceProvider.EXPECT().GetDatasourceByID(slo.DatasourceID).Return(&ds, nil)
				err := sloService.Create(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When grafana client tries to create new dashboard but fails", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, int64(0)).Times(1).Return(false, nil)
				mockISLOProvider.EXPECT().CreateSlo(&slo).Times(1).Return(nil)
				//assign id to freshly created slo
				slo.ID = sloID
				mockDashboardService.EXPECT().CreateDashboard(&userContext, gomock.Any(), false).Return(errory.ProviderErrors.New("dashboard error"))
				mockIDatasourceProvider.EXPECT().GetDatasourceByID(slo.DatasourceID).Return(&ds, nil)
				err := sloService.Create(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When no error during creation occurred", func() {
			It("Should correctly create an SLO", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, int64(0)).Times(1).Return(false, nil)
				mockISLOProvider.EXPECT().CreateSlo(&slo).Times(1).Return(nil)
				//assign id to freshly created slo
				slo.ID = sloID
				mockDashboardService.EXPECT().CreateDashboard(&userContext, gomock.Any(), false).Return(nil)
				mockIDatasourceProvider.EXPECT().GetDatasourceByID(slo.DatasourceID).Return(&ds, nil)
				err := sloService.Create(&userContext, &slo)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Update(slo *model.Slo)", func() {

		orgID := int64(2)

		slo := model.Slo{
			ID:                              66,
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}

		Context("When provider tries to find slo with the same name but returns an error", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, slo.ID).Times(1).Return(false, errory.NotUniqueErrors.New("duplicated slo name"))

				err := sloService.Update(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When provider finds slo with the same name", func() {
			It("Should return proper error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, slo.ID).Times(1).Return(true, nil)

				err := sloService.Update(&userContext, &slo)

				Expect(err).To(HaveOccurred())
				Expect(errory.IsOfType(err, errory.NotUniqueErrors)).To(BeTrue())
			})
		})

		Context("When provider tries to create new SLO but fails", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, slo.ID).Times(1).Return(false, nil)
				mockISLOProvider.EXPECT().UpdateSlo(&slo).Times(1).Return(errory.ProviderErrors.New("provider update slo error"))

				err := sloService.Update(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When grafana client tries to update a dashboard but fails", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().ContainSlosWithSameName(orgID, slo.Name, slo.ID).Times(1).Return(false, nil)
				mockISLOProvider.EXPECT().UpdateSlo(&slo).Times(1).Return(nil)

				mockDashboardService.EXPECT().CreateDashboard(&userContext, gomock.Any(), true).Return(errory.ProviderErrors.New("dashboard error"))

				err := sloService.Update(&userContext, &slo)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Delete(id int64)", func() {
		searchedID := int64(66)
		orgID := int64(2)
		slo := model.Slo{
			ID:                              66,
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}

		Context("When provider tries to find proper slo and fails", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(nil, errory.NotFoundErrors.New("Cannot find slo with given id"))

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When grafana client fails with deletion of the dashboard", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)
				mockDashboardService.EXPECT().DeleteDashboard(&userContext, slo.ID, slo.OrgID).Times(1).Return(errory.ProviderErrors.New("Cannot delete dashboard with given sloID"))

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When provider fails with deletion of the slo", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)
				mockDashboardService.EXPECT().DeleteDashboard(&userContext, slo.ID, slo.OrgID).Times(1).Return(nil)
				mockISLOProvider.EXPECT().DeleteSlo(&slo).Times(1).Return(errory.ProviderErrors.New("Cannot delete slo"))

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When provider fails with deletion of the slo", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)
				mockDashboardService.EXPECT().DeleteDashboard(&userContext, slo.ID, slo.OrgID).Times(1).Return(nil)
				mockISLOProvider.EXPECT().DeleteSlo(&slo).Times(1).Return(errory.ProviderErrors.New("Cannot delete slo"))

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).To(HaveOccurred())
			})
		})

		Context("When grafana client fails with error dashboard not found", func() {
			It("Should succeeded", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)
				mockDashboardService.EXPECT().DeleteDashboard(&userContext, slo.ID, slo.OrgID).Times(1).Return(errory.ProviderErrors.New("Dashboard not found"))
				mockISLOProvider.EXPECT().DeleteSlo(&slo).Times(1).Return(nil)

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When there is no error during deletion of the slo", func() {
			It("Should succeeded", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)
				mockDashboardService.EXPECT().DeleteDashboard(&userContext, slo.ID, slo.OrgID).Times(1).Return(nil)
				mockISLOProvider.EXPECT().DeleteSlo(&slo).Times(1).Return(nil)

				err := sloService.Delete(&userContext, searchedID)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Get(id int64)", func() {
		searchedID := int64(66)
		orgID := int64(2)
		slo := model.Slo{
			ID:                              66,
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}
		Context("When provider does not find correct slo", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(nil, errory.ProviderErrors.New("cannot find slo"))

				foundSlo, err := sloService.Get(searchedID)

				Expect(err).To(HaveOccurred())
				Expect(foundSlo).To(BeNil())
			})
		})
		Context("When provider finds correct slo", func() {
			It("Should succeed", func() {
				mockISLOProvider.EXPECT().GetSlo(searchedID).Times(1).Return(&slo, nil)

				foundSlo, err := sloService.Get(searchedID)

				Expect(err).NotTo(HaveOccurred())
				Expect(foundSlo).NotTo(BeNil())
				Expect(foundSlo.ID).To(Equal(searchedID))
			})
		})

	})

	Describe("DeleteSloHistory(id int64)", func() {
		sloID := int64(66)

		Context("Everything is OK", func() {
			It("Should succeed", func() {
				mockElasticClient.EXPECT().DeleteSloHistory(sloID).Return(nil)
				err := sloService.DeleteSloHistory(sloID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("could not send request", func() {
			It("Should return an error", func() {
				mockElasticClient.EXPECT().DeleteSloHistory(sloID).Return(errory.ElasticClientErrors.New("could not send request"))
				err := sloService.DeleteSloHistory(sloID)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetByOrgID(orgID int64)", func() {
		orgID := int64(2)
		sloList := []*model.Slo{{
			ID:                              66,
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}, {
			ID:                              67,
			OrgID:                           orgID,
			Name:                            "TestSLO2",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}}
		Context("When provider cannot list SLOs by organization", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().GetSlosByOrganizationID(orgID).Times(1).Return(nil, errory.NotFoundErrors.New("cannot find slo"))

				foundSloList, err := sloService.GetByOrgID(orgID)

				Expect(err).To(HaveOccurred())
				Expect(foundSloList).To(BeNil())
			})
		})
		Context("When provider lists SLOs by organization", func() {
			It("Should list proper list of SLOs", func() {
				mockISLOProvider.EXPECT().GetSlosByOrganizationID(orgID).Times(1).Return(sloList, nil)

				foundSloList, err := sloService.GetByOrgID(orgID)

				Expect(err).NotTo(HaveOccurred())
				Expect(foundSloList).NotTo(BeNil())
				Expect(len(foundSloList)).To(Equal(2))
			})
		})
	})

	Describe("FindSlos", func() {
		orgID := int64(2)
		sloParams := model.SloQueryParams{
			DatasourceType: "elasticsearch",
			MetricType:     model.MetricTypeCompliance,
			OrgID:          orgID,
		}
		slo := &model.Slo{
			ID:                              66,
			OrgID:                           orgID,
			Name:                            "TestSLO",
			SuccessRateExpectedAvailability: "99",
			ComplianceExpectedAvailability:  "99",
		}
		Context("When provider does not find correct slo", func() {
			It("Should return an error", func() {
				mockISLOProvider.EXPECT().FindSlos(&sloParams).Times(1).Return(nil, errory.NotFoundErrors.New("cannot find slo"))

				foundSlo, err := sloService.FindSlos(&sloParams)

				Expect(err).To(HaveOccurred())
				Expect(foundSlo).To(BeNil())
			})
		})
		Context("When provider finds correct slo", func() {
			It("Should succeed", func() {
				mockISLOProvider.EXPECT().FindSlos(&sloParams).Times(1).Return([]*model.Slo{slo}, nil)

				foundSlos, err := sloService.FindSlos(&sloParams)

				Expect(err).NotTo(HaveOccurred())
				Expect(foundSlos[0]).To(Equal(slo))
			})
		})

	})

	Describe("GetDetailedSlos", func() {
		var slos []*model.DetailedSlo

		BeforeEach(func() {
			slos = []*model.DetailedSlo{
				{
					Slo: model.Slo{
						ID:    5,
						OrgID: 25,
						Name:  "test-mcc9",
					},
					OrgName: "test-org",
				},
				{
					Slo: model.Slo{
						ID:    12,
						OrgID: 48,
						Name:  "test-mcc11",
					},
					OrgName: "test-org",
				},
			}
		})

		Context("When provider return slos", func() {
			It("Should return detailed slos", func() {
				mockISLOProvider.EXPECT().GetDetailedSlos("test-%", "test-org").Times(1).Return(slos, nil)

				foundSlos, err := sloService.GetDetailedSlos("test-%", "test-org")

				Expect(err).NotTo(HaveOccurred())
				Expect(len(foundSlos)).To(Equal(2))

				Expect(foundSlos[0].ID).To(Equal(int64(5)))
				Expect(foundSlos[0].OrgID).To(Equal(int64(25)))
				Expect(foundSlos[0].Name).To(Equal("test-mcc9"))
				Expect(foundSlos[0].OrgName).To(Equal("test-org"))

				Expect(foundSlos[1].ID).To(Equal(int64(12)))
				Expect(foundSlos[1].OrgID).To(Equal(int64(48)))
				Expect(foundSlos[1].Name).To(Equal("test-mcc11"))
				Expect(foundSlos[1].OrgName).To(Equal("test-org"))

			})
		})

		Context("When provider return error", func() {
			It("Should return error", func() {
				mockISLOProvider.EXPECT().GetDetailedSlos("test-%", "test-org").Times(1).Return(nil, errory.ProviderErrors.New("fatal provider error"))

				_, err := sloService.GetDetailedSlos("test-%", "test-org")
				Expect(err).To(HaveOccurred())
			})
		})

	})
})
