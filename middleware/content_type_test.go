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

var _ = Describe("ContentType middleware", func() {
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		middlewareUnderTest := middleware.ContentType()
		ginEngine.Use(middlewareUnderTest)
		ginEngine.POST("/test", handler)

	})

	Describe("Request with no contentType", func() {
		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodPost, "/test", nil)
			responseRecorder = httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should return unsupported Media Type error", func() {
			Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
		})

	})

	Describe("Request with unsupported Media Type", func() {
		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Add("Content-Type", "unsupported-media-type")
			responseRecorder = httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should return unsupported Media Type error", func() {
			Expect(responseRecorder.Code).To(Equal(http.StatusUnsupportedMediaType))
		})

	})

	Describe("Request with Json Media Type (supported media type)", func() {
		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodPost, "/test", nil)
			req.Header.Add("Content-Type", "application/json")
			responseRecorder = httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should pass with status ok", func() {
			Expect(responseRecorder.Code).To(Equal(http.StatusOK))
		})

	})
})
