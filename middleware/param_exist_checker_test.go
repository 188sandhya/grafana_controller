// +build unitTests

package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const parameteredURL = "test/:id/type"

var _ = Describe("ParamExistChecker", func() {
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	var req *http.Request
	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}
	var paramExistCheckerMiddleware gin.HandlerFunc
	var mockController *gomock.Controller
	var mockIParamExistCheckService *service.MockIParamExistCheckService

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIParamExistCheckService = service.NewMockIParamExistCheckService(mockController)
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
	})

	AfterEach(func() {
		mockController.Finish()
	})
	JustBeforeEach(func() {
		ginEngine.Use(paramExistCheckerMiddleware)
		ginEngine.GET(parameteredURL, handler)
		ginEngine.POST(parameteredURL, handler)
		responseRecorder = httptest.NewRecorder()
		ginEngine.ServeHTTP(responseRecorder, req)
	})

	Describe("CheckIfExist(idSelector IDSelector, existCheckService service.IExistCheckService,objectType service.Type, requiredPermission auth.Permission)", func() {
		AssertUnauthorizedBehavior := func() {
			It("should return unauthorized status code", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusUnauthorized))
			})
		}

		AssertProperBehavior := func() {
			It("should pass", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		}

		Describe("when id cannot of url cannot be parsed", func() {
			BeforeEach(func() {
				idSelector := api.GetIDParam
				paramExistCheckerMiddleware = middleware.CheckIfExist(idSelector, mockIParamExistCheckService, service.Slo)
				req, _ = http.NewRequest(http.MethodGet, "/test/xxx/type", nil)
			})
			AssertUnauthorizedBehavior()
		})
		Describe("when id of SLO can be parsed and exists", func() {
			BeforeEach(func() {
				idSelector := api.GetIDParam
				mockIParamExistCheckService.EXPECT().CheckSloByID(int64(12)).Times(1).Return(nil)
				paramExistCheckerMiddleware = middleware.CheckIfExist(idSelector, mockIParamExistCheckService, service.Slo)
				req, _ = http.NewRequest(http.MethodGet, "/test/12/type", nil)
			})
			AssertProperBehavior()
		})
		Describe("when id of org can be parsed and exists", func() {
			BeforeEach(func() {
				idSelector := api.GetIDParam
				mockIParamExistCheckService.EXPECT().CheckOrgByID(int64(123)).Times(1).Return(nil)
				paramExistCheckerMiddleware = middleware.CheckIfExist(idSelector, mockIParamExistCheckService, service.Organization)
				req, _ = http.NewRequest(http.MethodGet, "/test/123/type", nil)
			})
			AssertProperBehavior()
		})
		Describe("when id of org can be parsed and does not exists", func() {
			BeforeEach(func() {
				idSelector := api.GetIDParam
				mockIParamExistCheckService.EXPECT().CheckOrgByID(int64(123)).Times(1).Return(errory.NotFoundErrors.New("expected error"))
				paramExistCheckerMiddleware = middleware.CheckIfExist(idSelector, mockIParamExistCheckService, service.Organization)
				req, _ = http.NewRequest(http.MethodGet, "/test/123/type", nil)
			})
			It("should return Notfound status code", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
		Describe("when id of SLO can be parsed and does not exists", func() {
			BeforeEach(func() {
				idSelector := api.GetIDParam
				mockIParamExistCheckService.EXPECT().CheckSloByID(int64(123)).Times(1).Return(errory.NotFoundErrors.New("expected error"))
				paramExistCheckerMiddleware = middleware.CheckIfExist(idSelector, mockIParamExistCheckService, service.Slo)
				req, _ = http.NewRequest(http.MethodGet, "/test/123/type", nil)
			})
			It("should return Notfound status code", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
