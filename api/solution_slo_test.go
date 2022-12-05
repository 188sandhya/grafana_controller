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

var _ = Describe("SolutionSloAPI", func() {
	var mockController *gomock.Controller
	var solAPI *api.SolutionSloAPI
	var mockSolService *service.MockISolutionSloService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockSolService = service.NewMockISolutionSloService(mockController)
		solAPI = &api.SolutionSloAPI{
			SloService: mockSolService,
			Log:        logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		serviceDiscoveryRoutes := ginEngine.Group("/v1/solutionSlo")
		{
			serviceDiscoveryRoutes.GET("", solAPI.GetSolutionSlo)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("GetSolutionSlo()", func() {

		Context("When solService returns data", func() {
			data := model.SolutionSlo{
				OrgID:          1,
				NoCriticalSlos: 3,
				DashboardPath:  "/some/path",
			}
			BeforeEach(func() {
				mockSolService.EXPECT().GetSolutionSlo("errorbudget").Times(1).Return(&data, nil)
				expErr = nil
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutionSlo?orgName=errorbudget", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 200 code with proper message", func() {
				var expRes model.SolutionSlo
				err := json.Unmarshal(w.Body.Bytes(), &expRes)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(expRes.OrgID).To(Equal(int64(1)))
				Expect(expRes.NoCriticalSlos).To(Equal(int64(3)))
				Expect(expRes.DashboardPath).To(Equal("/some/path"))
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("When orgName param not present occurs", func() {
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutionSlo", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("When error occurs", func() {
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("errory")
				mockSolService.EXPECT().GetSolutionSlo("incorrect").Times(1).Return(nil, expErr)
			})
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", "/v1/solutionSlo?orgName=incorrect", nil)
				ginEngine.ServeHTTP(w, req)
			})
			It("returns 500 code with error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "errory", expErr)
				assertions.AssertLogger(logHook, "ebt.provider_error: errory")
			})
		})
	})
})
