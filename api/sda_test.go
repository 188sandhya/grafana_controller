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
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("SDAAPI", func() {
	var mockController *gomock.Controller
	var sdaAPI *api.SDAAPI
	var mockSDAService *service.MockISDAService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockSDAService = service.NewMockISDAService(mockController)
		sdaAPI = &api.SDAAPI{
			Service: mockSDAService,
			Log:     logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		serviceDiscoveryRoutes := ginEngine.Group("/v1/sda")
		{
			serviceDiscoveryRoutes.GET("", sdaAPI.GetConfig)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("GetConfig()", func() {

		Context("When sdaService returns a correct configuration", func() {
			BeforeEach(func() {
				jsonData := `[{"errorbudget": {"git": [{"url": "git@github.com:metro-digital-inner-source/errorbudget-grafana-controller.git"},{"url": "git@github.com:metro-digital-inner-source/errorbudget-sda.git"},{"url": "git@github.com:metro-digital-inner-source/errorbudget-susie.git"}],"cicd": [{"url": "https://errorbudget.cip.metronom.com","jobs_scoring_whitelist": ["autoslo-prod","errorbudget-v2-prod","sda-prod","susie-prod"],"jobs_scraping_whitelist": []}],"jira": [{"url": "https://jira.metrosystems.net","project": "OPEB"}]}}]`
				var raw []map[string]interface{}
				json.Unmarshal([]byte(jsonData), &raw)

				mockSDAService.EXPECT().GetConfig().Times(1).Return(raw, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/sda", nil)
				ginEngine.ServeHTTP(w, req)
				expErr = nil
			})
			It("returns 200 code with proper message", func() {
				var expRes []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(expRes[0]).To(HaveKey("errorbudget"))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When sdaService returns no data", func() {
			BeforeEach(func() {
				jsonData := `[]`
				var raw []map[string]interface{}
				json.Unmarshal([]byte(jsonData), &raw)

				mockSDAService.EXPECT().GetConfig().Times(1).Return(raw, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/sda", nil)
				ginEngine.ServeHTTP(w, req)
				expErr = nil
			})
			It("returns 200 code with proper message", func() {
				var expRes []map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When sloService does not return a correct list of slos", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("errory")
				mockSDAService.EXPECT().GetConfig().Times(1).Return(nil, someErr)
				expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/sda", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SDA configuration (errory)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SDA configuration, cause: ebt.provider_error: errory")
			})
		})
	})
})
