// +build unitTests

package middleware_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	authModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Authorize Middleware", func() {
	var mockController *gomock.Controller
	var mockAuthorizer *middleware.MockIAuthorizer
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	var req *http.Request
	logger, logHook := logrustest.NewNullLogger()
	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}
	var authorizeMiddleware gin.HandlerFunc
	const url = "test/:id"

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAuthorizer = middleware.NewMockIAuthorizer(mockController)
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		logHook.Reset()
	})
	JustBeforeEach(func() {
		ginEngine.Use(authorizeMiddleware)
		ginEngine.GET(url, handler)
		ginEngine.POST(url, handler)
		responseRecorder = httptest.NewRecorder()
		ginEngine.ServeHTTP(responseRecorder, req)
	})

	Describe("Authorize(idSelector IDSelector, objectType auth.Type, requiredPermission auth.Permission)", func() {
		AssertFailedBehavior := func(code int) {
			It("should return correct status code", func() {
				Expect(responseRecorder.Code).To(Equal(code))
			})
		}

		AssertProperBehavior := func() {
			It("should pass", func() {
				Expect(responseRecorder.Code).To(Equal(http.StatusOK))
			})
		}

		Describe("Not authenticated user", func() {
			Context("has no user in a request", func() {
				BeforeEach(func() {
					idSelector := api.GetIDParam
					authorizeMiddleware = middleware.Authorize(idSelector, mockAuthorizer.AuthorizeForSLO, authModel.Admin, logger)
					req, _ = http.NewRequest(http.MethodGet, "/test/xxx", nil)
				})
				AssertFailedBehavior(http.StatusInternalServerError)
			})
		})

		Describe("GetIDParam idSelector", func() {
			BeforeEach(func() {
				ginEngine.Use(auhenticateMiddleware(14))
				idSelector := api.GetIDParam
				authorizeMiddleware = middleware.Authorize(idSelector, mockAuthorizer.AuthorizeForSLO, authModel.Admin, logger)
				req, _ = http.NewRequest(http.MethodGet, "/test/123", nil)
			})

			Describe("when cannot parse id", func() {
				BeforeEach(func() {
					req, _ = http.NewRequest(http.MethodGet, "/test/xxx", nil)
				})
				AssertFailedBehavior(http.StatusBadRequest)
			})

			Context("when authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(123), int64(14), authModel.Admin).Times(1).Return(true, nil)
				})
				AssertProperBehavior()
			})

			Context("when not authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(123), int64(14), authModel.Admin).Times(1).Return(false, nil)
				})
				AssertFailedBehavior(http.StatusUnauthorized)
			})

			Context("when authorizer returns error", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(123), int64(14), authModel.Admin).Times(1).Return(false, errory.ProviderErrors.New("provider error"))
				})
				AssertFailedBehavior(http.StatusInternalServerError)
			})
		})

		Describe("OrgIDInStruct idSelector", func() {
			BeforeEach(func() {
				ginEngine.Use(auhenticateMiddleware(14))
				idSelector := middleware.OrgIDInStruct
				authorizeMiddleware = middleware.Authorize(idSelector, mockAuthorizer.AuthorizeForSLO, authModel.Admin, logger)
				req, _ = http.NewRequest(http.MethodPost, "/test/xxx", bytes.NewBufferString(`{"orgId": 4 }`))
			})

			Describe("when cannot extract id", func() {
				BeforeEach(func() {
					req, _ = http.NewRequest(http.MethodPost, "/test/xxx", bytes.NewBufferString(`{"json": "withNoOrgID" }`))
				})
				AssertFailedBehavior(http.StatusBadRequest)
			})

			Context("when authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(4), int64(14), authModel.Admin).Times(1).Return(true, nil)
				})
				AssertProperBehavior()
			})

			Context("when not authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(4), int64(14), authModel.Admin).Times(1).Return(false, nil)
				})
				AssertFailedBehavior(http.StatusUnauthorized)
			})

			Context("when authorizer returns error", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(4), int64(14), authModel.Admin).Times(1).Return(false, errory.ProviderErrors.New("provider error"))
				})
				AssertFailedBehavior(http.StatusInternalServerError)
			})
		})

		Describe("SloIDInStruct idSelector", func() {
			BeforeEach(func() {
				ginEngine.Use(auhenticateMiddleware(14))
				idSelector := middleware.SloIDInStruct
				authorizeMiddleware = middleware.Authorize(idSelector, mockAuthorizer.AuthorizeForSLO, authModel.Admin, logger)
				req, _ = http.NewRequest(http.MethodPost, "/test/xxx", bytes.NewBufferString(`{"sloId": 420 }`))
			})

			Describe("when cannot extract id", func() {
				BeforeEach(func() {
					req, _ = http.NewRequest(http.MethodPost, "/test/xxx", bytes.NewBufferString(`{"json": "withNoSloID" }`))
				})
				AssertFailedBehavior(http.StatusBadRequest)
			})

			Context("when authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(420), int64(14), authModel.Admin).Times(1).Return(true, nil)
				})
				AssertProperBehavior()
			})

			Context("when not authorized", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(420), int64(14), authModel.Admin).Times(1).Return(false, nil)
				})
				AssertFailedBehavior(http.StatusUnauthorized)
			})

			Context("when authorizer returns error", func() {
				BeforeEach(func() {
					mockAuthorizer.EXPECT().AuthorizeForSLO(int64(420), int64(14), authModel.Admin).Times(1).Return(false, errory.ProviderErrors.New("provider error"))
				})
				AssertFailedBehavior(http.StatusInternalServerError)
			})
		})
	})
})

func auhenticateMiddleware(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("UserContext", &authModel.UserContext{ID: userID, Cookie: "test cookie"})
		c.Next()
	}
}
