//go:build unitTests
// +build unitTests

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("ProductsStatusAPI", func() {
	var mockController *gomock.Controller
	var psAPI *api.ProductsStatusAPI
	var mockPSService *service.MockIProductsStatusService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockPSService = service.NewMockIProductsStatusService(mockController)
		psAPI = &api.ProductsStatusAPI{
			Service: mockPSService,
			Log:     logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		psRoutes := ginEngine.Group("/v1/products_status")
		{
			psRoutes.GET("", psAPI.GetProductsStatus)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("GetProductsStatus()", func() {

		Context("When psService returns data", func() {
			BeforeEach(func() {
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
				mockPSService.EXPECT().GetProductsStatus().Times(1).Return(data, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/products_status", nil)
				ginEngine.ServeHTTP(w, req)
				expErr = nil
			})
			It("returns 200 code with proper message", func() {
				var expRes []*model.ProductStatus
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(expRes[0].MarcID).To(Equal("A6789"))
				Expect(expRes[0].ProductName).To(Equal("oma"))
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When psService returns no data", func() {
			BeforeEach(func() {
				mockPSService.EXPECT().GetProductsStatus().Times(1).Return([]*model.ProductStatus{}, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/products_status", nil)
				ginEngine.ServeHTTP(w, req)
				expErr = nil
			})
			It("returns 200 code with proper message", func() {
				var expRes []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(len(expRes)).To(Equal(0))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When error and no products status", func() {
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("errory")
				mockPSService.EXPECT().GetProductsStatus().Times(1).Return(nil, expErr)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/products_status", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get Products Status (errory)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get Products Status, cause: ebt.provider_error: errory")
			})
		})
	})
})
