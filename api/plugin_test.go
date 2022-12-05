//go:build unitTests
// +build unitTests

package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	pluginService "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("PluginAPI", func() {
	var mockController *gomock.Controller
	var pluginAPI *PluginAPI
	var pluginServiceMock *pluginService.MockIPluginService
	var organizationServiceMock *service.MockIOrganizationService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	const cookie = "test cookie"
	var expErr error
	var userContext auth.UserContext
	var userContextMiddleware func(*gin.Context)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		pluginServiceMock = pluginService.NewMockIPluginService(mockController)
		organizationServiceMock = service.NewMockIOrganizationService(mockController)
		pluginAPI = &PluginAPI{
			Plugin:              pluginServiceMock,
			OrganizationService: organizationServiceMock,
			Log:                 logger,
		}

		userContext = auth.UserContext{
			ID:     4,
			Cookie: "test cookie",
		}
		userContextMiddleware = func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		gin.SetMode(gin.TestMode)

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("Enable()", func() {
		var err error
		const orgID = int64(99)
		BeforeEach(func() {
			req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/plugin/%d", orgID), nil)
			req.Header.Set("Content-Type", "application/json")
		})
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				ginEngine = gin.New()
				ginEngine.POST("/v1/plugin/:id", userContextMiddleware, pluginAPI.Enable)

				org := &grafanaModel.Organization{
					ID:   orgID,
					Name: "mrc",
				}

				organizationServiceMock.EXPECT().GetOrganizationByID(orgID).Times(1).Return(org, err)
				pluginServiceMock.EXPECT().EnablePluginWithDataSources(org, cookie, false).Times(1)
			})

			It("returns 201 code and id of modified organization", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusCreated))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, orgID)))
			})
		})

		Context("no UserContext in context", func() {
			BeforeEach(func() {
				ginEngine = gin.New()
				ginEngine.POST("/v1/plugin/:id", pluginAPI.Enable)
				expErr = errory.APIErrors.Builder().WithMessage("UserContext not found in context").Create()
			})

			It("returns error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot extract user context (could not extract UserContext)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_update_error: Cannot extract user context, cause: ebt.processing_error: could not extract UserContext")
			})
		})
	})
})
