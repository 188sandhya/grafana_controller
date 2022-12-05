// +build unitTests

package middleware_test

import (
	"net/http"

	"encoding/json"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

const WHATEVER = "¯\\_(ツ)_/¯"

var _ = Describe("Recovery Middleware", func() {
	var ginEngine *gin.Engine

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
	})

	failingHandler := func(context *gin.Context) {
		panic(WHATEVER)
	}

	Describe("Panic handler", func() {
		var responseRecorder *httptest.ResponseRecorder
		var writer *testWriter

		BeforeEach(func() {
			logger := logrus.New()
			logger.Formatter = &logrus.JSONFormatter{}

			writer = &testWriter{}
			logger.Out = writer

			recoveryMiddleware := middleware.Recovery(logger)
			ginEngine.Use(recoveryMiddleware)
			ginEngine.GET("/failing", failingHandler)
			req, _ := http.NewRequest(http.MethodGet, "/failing", nil)
			responseRecorder = httptest.NewRecorder()
			ginEngine.ServeHTTP(responseRecorder, req)

		})

		It("should return HTTP 500", func() {
			Expect(responseRecorder.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should log the panic", func() {
			Expect(writer.result).NotTo(BeNil())
		})

		It("should log as error", func() {
			var logAsMap map[string]interface{}
			err := json.Unmarshal(writer.result, &logAsMap)
			Expect(logAsMap).To(HaveKeyWithValue("level", "error"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should log as message", func() {
			var logAsMap map[string]interface{}
			err := json.Unmarshal(writer.result, &logAsMap)
			Expect(logAsMap).To(HaveKey("msg"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have a message containing panic text", func() {
			var logAsMap map[string]interface{}
			err := json.Unmarshal(writer.result, &logAsMap)
			Expect(logAsMap["msg"]).To(ContainSubstring(WHATEVER))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have a body with standdard message", func() {
			actualResponse := make(map[string]interface{})
			err := json.Unmarshal(responseRecorder.Body.Bytes(), &actualResponse)
			Expect(actualResponse["message"]).To(Equal(middleware.AbortMessage))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

type testWriter struct {
	result []byte
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	t.result = p
	return len(p), nil
}
