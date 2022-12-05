// +build unitTests

package middleware_test

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppJSONHeader middleware", func() {
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		middlewareUnderTest := middleware.AppJSONHeader()
		ginEngine.Use(middlewareUnderTest)
		ginEngine.POST("/test", handler)

	})

	Describe("Request with no contentType", func() {
		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodPost, "/test", nil)
			responseRecorder = httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should return Response with content type application/json", func() {
			Expect(responseRecorder.Header()).To(HaveKeyWithValue("Content-Type", []string{"application/json; charset=utf-8"}))
		})

	})
})
