//go:build unitTests
// +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Test Products Status service", func() {
	var mockController *gomock.Controller
	var mockPSProvider *provider.MockIProductsStatusProvider
	var psService *service.ProductsStatusService
	logger, logHook := logrustest.NewNullLogger()

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockPSProvider = provider.NewMockIProductsStatusProvider(mockController)

		psService = &service.ProductsStatusService{
			Provider: mockPSProvider,
			Log:      logger,
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("GetProductsStatus", func() {
		Context("When provider finds data", func() {
			It("Should not return an error", func() {
				data := []*model.ProductStatus{
					{
						MonitorID:    "962730",
						MarcID:       "A6789",
						ProductName:  "oma",
						ServiceClass: "Bronze",
						Level:        "Solution",
						ClassDefinition: &model.ServiceClassDefinition{
							Name:                         "Bronze",
							MaxDowntimeFrequencyAllowed:  "month",
							MaxContinuousDowntimeAllowed: 8,
							AvailabilityUpperThreshold:   "98.20",
							MinimumAvailability:          "98.00",
							ServiceTimeHours:             10,
							ServiceTimeDays:              5,
						},
						Availability:                "99.99",
						NumberOfDowntimes:           1,
						MaxContinuousDowntime:       100,
						AvailabilityStatus:          "Green",
						NumberOfDowntimesStatus:     "Green",
						MaxContinuousDowntimeStatus: "Green",
						MonitorStatus:               "OK",
						FromDate:                    1593561600000,
						ToDate:                      1595807999000,
					},
					{
						MonitorID:    "4364759",
						MarcID:       "A12345",
						ProductName:  "transportation",
						ServiceClass: "Gold",
						Level:        "Product",
						ClassDefinition: &model.ServiceClassDefinition{
							Name:                         "Gold",
							MaxDowntimeFrequencyAllowed:  "quarter",
							MaxContinuousDowntimeAllowed: 2,
							AvailabilityUpperThreshold:   "99.60",
							MinimumAvailability:          "99.55",
							ServiceTimeHours:             24,
							ServiceTimeDays:              7,
						},
						Availability:                "99.99",
						NumberOfDowntimes:           1,
						MaxContinuousDowntime:       100,
						AvailabilityStatus:          "Green",
						NumberOfDowntimesStatus:     "Green",
						MaxContinuousDowntimeStatus: "Green",
						MonitorStatus:               "OK",
						FromDate:                    1593561600000,
						ToDate:                      1595807999000,
					},
				}
				mockPSProvider.EXPECT().GetProductsStatus().Times(1).Return(data, nil)

				result, err := psService.GetProductsStatus()

				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].MarcID).To(Equal("A6789"))
				Expect(result[0].ProductName).To(Equal("oma"))
			})
		})

		Context("When provider DOES NOT find data", func() {
			It("Should return an error", func() {
				mockPSProvider.EXPECT().GetProductsStatus().Times(1).Return(nil, errory.NotFoundErrors.New("Data not found"))

				result, err := psService.GetProductsStatus()

				Expect(err).To(HaveOccurred())
				Expect(len(result)).To(Equal(0))
			})
		})
	})
})
