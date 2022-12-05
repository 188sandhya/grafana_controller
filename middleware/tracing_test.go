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

var _ = Describe("Tracing Middleware", func() {
	var ginEngine *gin.Engine
	var actualCorrelationID func() string
	var valueExists func() bool

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		var handler func(c *gin.Context)
		handler, actualCorrelationID, valueExists = ValueRecordingHandler()
		middlewareUnderTest := middleware.Tracing
		ginEngine.Use(middlewareUnderTest)
		ginEngine.GET("/test", handler)
	})

	Describe("Given a HTTP Request", func() {

		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			responseRecorder := httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})

		It("generates correlation id and adds it to the context", func() {
			Expect(valueExists()).To(BeTrue())
		})
		It("generates a correlation id with uuid", func() {
			Expect(actualCorrelationID()).To(MatchRegexp("[a-f0-9-]+"))
		})
	})

	Describe("Given a HTTP Request with a correlation ID", func() {

		const sampleCorrelationID = "sample-correlation-id"

		BeforeEach(func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Add("X-Correlation-ID", sampleCorrelationID)
			responseRecorder := httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)
		})

		It("has a correlation id in the context", func() {
			Expect(valueExists()).To(BeTrue())
		})
		It("adds the correlation id in the incoming request to context", func() {
			Expect(actualCorrelationID()).To(Equal(sampleCorrelationID))
		})
	})
})

func ValueRecordingHandler() (fhandler func(c *gin.Context), fActualCorrelationID func() string, fValueExists func() bool) {
	var extractedValue string
	valueExists := false

	return func(c *gin.Context) {
			var v interface{}
			v, valueExists = c.Get("X-Correlation-ID")
			if valueExists {
				extractedValue = v.(string)
			}
			c.Status(http.StatusOK)
		},
		func() string {
			return extractedValue
		},
		func() bool {
			return valueExists
		}
}
