// +build unitTests

package middleware_test

import (
	"encoding/json"
	"net/http"
	"time"

	"net/http/httptest"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

const TimeoutMessage = "Timeout Exceeded"

var _ = Describe("Timeout Logger Middleware", func() {

	BeforeEach(func() {
	})
	//nolint
	timeoutHandler := func(w http.ResponseWriter, r *http.Request) {
		for {
			time.Sleep(60 * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
	}

	Describe("Timeout handler", func() {
		var responseRecorder *httptest.ResponseRecorder
		var writer *testWriter

		BeforeEach(func() {
			logger := logrus.New()
			logger.Formatter = &logrus.JSONFormatter{}

			writer = &testWriter{}
			logger.Out = writer

			timeout := 100 * time.Millisecond
			handler := middleware.TimeoutLogHandler(http.TimeoutHandler(http.HandlerFunc(timeoutHandler), timeout, TimeoutMessage), logger, timeout)
			req, _ := http.NewRequest(http.MethodGet, "/timeout", nil)
			responseRecorder = httptest.NewRecorder()
			handler.ServeHTTP(responseRecorder, req)
		})

		It("should reponse HTTP 503 and timeout body message", func() {
			Expect(responseRecorder.Code).To(Equal(http.StatusServiceUnavailable))
			Expect(responseRecorder.Body.String()).To(Equal(TimeoutMessage))
		})

		It("should log status as 503 and level as error", func() {
			var logAsMap map[string]interface{}
			err := json.Unmarshal(writer.result, &logAsMap)
			Expect(err).NotTo(HaveOccurred())
			Expect(logAsMap).To(HaveKeyWithValue("status", BeNumerically("==", http.StatusServiceUnavailable)))
			Expect(logAsMap).To(HaveKeyWithValue("level", "error"))
		})
	})
})
