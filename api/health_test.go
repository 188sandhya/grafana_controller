// +build unitTests

package api_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("HealthAPI", func() {
	var mockController *gomock.Controller
	var healthAPI *HealthAPI
	var healthServiceMock *service.MockIHealthService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		healthServiceMock = service.NewMockIHealthService(mockController)
		healthAPI = &HealthAPI{
			HealthService: healthServiceMock,
			Log:           logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		ginEngine.GET("/api/health", healthAPI.HealthCheck)
		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("Get()", func() {
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/api/health", nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				healthServiceMock.EXPECT().CheckHealth().Times(1).Return(nil)
			})

			It("returns 204 code with empty body", func() {
				Expect(w.Code).To(Equal(http.StatusNoContent))
				Expect(w.Body.String()).To(BeEmpty())
			})
		})

		Context("when datasource service returns unexpected  error", func() {
			var expErr error
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("datasource unexepected error")
				healthServiceMock.EXPECT().CheckHealth().Times(1).Return(expErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Database healthcheck failed (datasource unexepected error)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Database healthcheck failed, cause: ebt.provider_error: datasource unexepected error")
			})
		})
	})
})
