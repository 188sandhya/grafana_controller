// +build unitTests

package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Logger Middleware", func() {
	var ginEngine *gin.Engine
	var responseRecorder *httptest.ResponseRecorder
	var req *http.Request
	var writer *testWriter
	var loggerMiddleware gin.HandlerFunc
	const timeout = 2 * time.Second

	handler := func(context *gin.Context) {
		context.Status(http.StatusOK)
	}
	warningHandler := func(context *gin.Context) {
		time.Sleep(1100 * time.Millisecond)
		context.Status(http.StatusOK)
	}

	Context("When making a request with logger middleware connected", func() {
		BeforeEach(func() {
			writer = &testWriter{}
			logger := createLoggerWithStandardFields(writer)

			gin.SetMode(gin.TestMode)
			ginEngine = gin.New()
			loggerMiddleware = middleware.LoggerMiddleware(logger, timeout)
			ginEngine.Use(loggerMiddleware)
			ginEngine.GET("/info/", handler)

			responseRecorder = httptest.NewRecorder()
			req, _ = http.NewRequest(http.MethodGet, "/info/", nil)
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should create log entry correctly on info level", func() {
			var logAsMap map[string]interface{}
			_ = json.Unmarshal(writer.result, &logAsMap)
			Expect(logAsMap).To(HaveKey("@timestamp"))
			Expect(logAsMap).To(HaveKey("status"))
			Expect(logAsMap).To(HaveKey("userAgent"))
			Expect(logAsMap).To(HaveKey("retention"))
			Expect(logAsMap).To(HaveKey("@service-name"))
			Expect(logAsMap).To(HaveKey("@service-version"))
			Expect(logAsMap).To(HaveKey("elapsed"))
			Expect(logAsMap).To(HaveKey("method"))
			Expect(logAsMap).To(HaveKey("msg"))
			Expect(logAsMap).To(HaveKey("level"))
			Expect(logAsMap).To(HaveKey("uri"))
			Expect(logAsMap).To(HaveKey("@hostname"))
			Expect(logAsMap).To(HaveKey("@vertical"))
			Expect(logAsMap).To(HaveKey("@type"))
			Expect(logAsMap).To(HaveKey("correlationID"))
			Expect(logAsMap).To(HaveKey("ip"))
			Expect(logAsMap).To(HaveKey("event"))
			Expect(logAsMap).To(HaveKey("protocol"))

			Expect(logAsMap["level"]).To(Equal("info"))
		})
	})
	Context("When making a request with logger middleware connected", func() {
		BeforeEach(func() {
			writer = &testWriter{}
			logger := createLoggerWithStandardFields(writer)

			gin.SetMode(gin.TestMode)
			ginEngine = gin.New()
			loggerMiddleware = middleware.LoggerMiddleware(logger, timeout)
			ginEngine.Use(loggerMiddleware)
			ginEngine.GET("/warning/", warningHandler)

			responseRecorder = httptest.NewRecorder()
			req, _ = http.NewRequest(http.MethodGet, "/warning/", nil)
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("should create log entry correctly on warning level", func() {
			var logAsMap map[string]interface{}
			_ = json.Unmarshal(writer.result, &logAsMap)
			Expect(logAsMap).To(HaveKey("@timestamp"))
			Expect(logAsMap).To(HaveKey("status"))
			Expect(logAsMap).To(HaveKey("userAgent"))
			Expect(logAsMap).To(HaveKey("retention"))
			Expect(logAsMap).To(HaveKey("@service-name"))
			Expect(logAsMap).To(HaveKey("@service-version"))
			Expect(logAsMap).To(HaveKey("elapsed"))
			Expect(logAsMap).To(HaveKey("method"))
			Expect(logAsMap).To(HaveKey("msg"))
			Expect(logAsMap).To(HaveKey("level"))
			Expect(logAsMap).To(HaveKey("uri"))
			Expect(logAsMap).To(HaveKey("@hostname"))
			Expect(logAsMap).To(HaveKey("@vertical"))
			Expect(logAsMap).To(HaveKey("@type"))
			Expect(logAsMap).To(HaveKey("correlationID"))
			Expect(logAsMap).To(HaveKey("ip"))
			Expect(logAsMap).To(HaveKey("event"))
			Expect(logAsMap).To(HaveKey("protocol"))

			Expect(logAsMap["level"]).To(Equal("warning"))
		})
	})
	Context("When making a request with logger middleware connected", func() {
		BeforeEach(func() {
			writer = &testWriter{}
			logger := createLoggerWithStandardFields(writer)

			gin.SetMode(gin.TestMode)
			ginEngine = gin.New()
			loggerMiddleware = middleware.LoggerMiddleware(logger, 1*time.Second)
			ginEngine.Use(loggerMiddleware)
			ginEngine.GET("/warning/", warningHandler)

			responseRecorder = httptest.NewRecorder()
			req, _ = http.NewRequest(http.MethodGet, "/warning/", nil)
			ginEngine.ServeHTTP(responseRecorder, req)
		})
		It("shouldn't create log it's special case of timeout", func() {
			var logAsMap map[string]interface{}
			_ = json.Unmarshal(writer.result, &logAsMap)
			Expect(string(writer.result)).Should(BeEmpty())
		})
	})

})

func createLoggerWithStandardFields(writer io.Writer) logrus.FieldLogger {
	newLogger := logrus.New()
	newLogger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap:        logrus.FieldMap{logrus.FieldKeyTime: "@timestamp"},
	}

	logger := newLogger.WithFields(logrus.Fields{
		middleware.LogfieldOrganization:   "errorbudget",
		middleware.LogfieldServiceName:    "cruiser",
		middleware.LogfieldHostname:       "cruiser",
		middleware.LogfieldServiceVersion: "spike",
		middleware.LogfieldRetention:      "technical",
		middleware.LogfieldType:           "service",
	})

	logger.Logger.Out = writer

	return logger
}
