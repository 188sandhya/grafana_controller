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

var _ = Describe("SolutionsAPI", func() {
	var mockController *gomock.Controller
	var fsAPI *api.SolutionsAPI
	var mockFSService *service.MockISolutionsService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockFSService = service.NewMockISolutionsService(mockController)
		fsAPI = &api.SolutionsAPI{
			Service: mockFSService,
			Log:     logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		serviceDiscoveryRoutes := ginEngine.Group("/v1/solutions")
		{
			serviceDiscoveryRoutes.GET("", fsAPI.GetSolutions)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("GetSolutions()", func() {

		Context("When fsService returns data", func() {
			BeforeEach(func() {
				data := []*model.Solution{
					{
						ID:          1,
						Name:        "M|COMPANION",
						ProductID:   new(int64),
						ProductName: "M|COMPANION",
					},
					{
						ID:          7,
						Name:        "M|CREDIT",
						ProductID:   new(int64),
						ProductName: "Customer Credit Limit Management System",
					},
				}
				mockFSService.EXPECT().GetSolutions(true, false, "").Times(1).Return(data, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutions?long=true", nil)
				ginEngine.ServeHTTP(w, req)
				expErr = nil
			})
			It("returns 200 code with proper message", func() {
				var expRes []*model.Solution
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(expRes[0].ID).To(Equal(int64(1)))
				Expect(expRes[0].Name).To(Equal("M|COMPANION"))
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When fsService returns no data", func() {
			BeforeEach(func() {
				mockFSService.EXPECT().GetSolutions(true, false, "").Times(1).Return([]*model.Solution{}, nil)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutions?long=true", nil)
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

		Context("When error and no solutions", func() {
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("errory")
				mockFSService.EXPECT().GetSolutions(true, false, "").Times(1).Return(nil, expErr)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutions?long=true", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get Solutions (errory)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get Solutions, cause: ebt.provider_error: errory")
			})
		})

		Context("When error and wrong solution scope type", func() {
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("unknown solution scope type: xyz")
				mockFSService.EXPECT().GetSolutions(true, false, "xyz").Times(1).Return(nil, expErr)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutions?long=true&solutionScope=xyz", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get Solutions (unknown solution scope type: xyz)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get Solutions, cause: ebt.provider_error: unknown solution scope type: xyz")
			})
		})
	})
})
