// +build unitTests

package api_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("ConfigureUserAPI", func() {
	var mockController *gomock.Controller
	var userInfoServiceMock *service.MockIUserInfoService
	var configureUserAPI *ConfigureUserAPI
	var userInfo *model.UserInfo

	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		userInfoServiceMock = service.NewMockIUserInfoService(mockController)
		configureUserAPI = &ConfigureUserAPI{
			UserInfoService: userInfoServiceMock,
			Log:             logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContext := auth.UserContext{
			ID:     1234,
			Cookie: "abcd",
		}

		userContextMiddleware := func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		ginEngine.GET("/v1/configure_user", userContextMiddleware, configureUserAPI.ConfigureUser)
		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("Get()", func() {
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "/v1/configure_user", nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				userInfo = &model.UserInfo{Name: "nameXYZ", Login: "loginXYZ", Email: "email@XYZ"}
				userInfoServiceMock.EXPECT().GetUserInfo(int64(1234)).Times(1).Return(userInfo, nil)
			})
			It("returns 200 code with empty body", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal(`{"name":"nameXYZ","login":"loginXYZ","email":"email@XYZ","cookie":"abcd"}`))
			})
		})
	})
})
