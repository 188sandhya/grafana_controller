// +build unitTests

package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	authModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Access Middleware", func() {
	var mockController *gomock.Controller
	var mockAuthenticator *auth.MockIAuthenticator
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	var req *http.Request
	logger, _ := logrustest.NewNullLogger()
	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}
	var authenticateMiddleware gin.HandlerFunc
	var afterTestMiddleware gin.HandlerFunc
	const url = "test/:id"

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAuthenticator = auth.NewMockIAuthenticator(mockController)
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
	})
	JustBeforeEach(func() {
		ginEngine.Use(authenticateMiddleware)
		ginEngine.Use(afterTestMiddleware)
		ginEngine.GET(url, handler)
		ginEngine.POST(url, handler)
		responseRecorder = httptest.NewRecorder()
		ginEngine.ServeHTTP(responseRecorder, req)
	})

	Describe("Authenticate(authenticator auth.IAuthenticator, log logrus.FieldLogger)", func() {
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

		Context("when authenticate pass", func() {
			BeforeEach(func() {
				authenticateMiddleware = middleware.Authenticate(mockAuthenticator, logger)
				mockAuthenticator.EXPECT().Authenticate(gomock.Any()).Times(1).Return(&authModel.UserContext{ID: 12, Cookie: "test cookie"}, nil)
				req, _ = http.NewRequest(http.MethodGet, "/test/xxx", nil)
				afterTestMiddleware = assertContextMiddleware(12, "test cookie")
			})
			AssertProperBehavior()
		})

		Context("when authenticate fails", func() {
			BeforeEach(func() {
				authenticateMiddleware = middleware.Authenticate(mockAuthenticator, logger)
				mockAuthenticator.EXPECT().Authenticate(gomock.Any()).Times(1).Return(nil, errory.AuthErrors.New("authentication error"))
				req, _ = http.NewRequest(http.MethodGet, "/test/xxx", nil)
				afterTestMiddleware = assertNotExecutedMiddleware()
			})
			AssertFailedBehavior(http.StatusUnauthorized)
		})

		Context("when authenticate fails with incorrectly created error", func() {
			BeforeEach(func() {
				authenticateMiddleware = middleware.Authenticate(mockAuthenticator, logger)
				mockAuthenticator.EXPECT().Authenticate(gomock.Any()).Times(1).Return(nil, errory.AuthErrors.New("should have payload"))
				req, _ = http.NewRequest(http.MethodGet, "/test/xxx", nil)
				afterTestMiddleware = assertNotExecutedMiddleware()
			})
			AssertFailedBehavior(http.StatusUnauthorized)
		})

	})
})

func assertContextMiddleware(userID int64, cookie string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("UserContext")
		Expect(exists).To(BeTrue())
		userContext := u.(*authModel.UserContext)
		Expect(userContext.ID).To(Equal(userID))
		Expect(userContext.Cookie).To(Equal(cookie))
		c.Next()
	}
}

func assertNotExecutedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		Fail("this middleware should not be executed")
	}
}
