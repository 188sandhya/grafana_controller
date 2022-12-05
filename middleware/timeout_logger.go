package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/sirupsen/logrus"
)

const errorMessage = "[Timeout] timeout exceeded"

func TimeoutLogHandler(h http.Handler, l logrus.FieldLogger, dt time.Duration) http.Handler {
	return &timeoutLogHandler{
		logger:  l,
		handler: h,
		dt:      dt,
	}
}

type timeoutLogHandler struct {
	logger  logrus.FieldLogger
	handler http.Handler
	dt      time.Duration
}

func (h *timeoutLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	h.handler.ServeHTTP(w, r)

	timeElapsed := time.Since(startTime)
	if timeElapsed < h.dt {
		return
	}

	ip := r.RemoteAddr
	if index := strings.LastIndex(ip, ":"); index != -1 {
		ip = ip[:index]
	}
	h.logger.WithFields(logrus.Fields{
		LogfieldEvent: "request",
		"ip":          ip,
		"method":      r.Method,
		"uri":         r.URL.String(),
		"protocol":    r.Proto,
		"status":      http.StatusServiceUnavailable,
		"bytes":       0,
		"elapsed":     timeElapsed / time.Millisecond,
		"userAgent":   r.UserAgent(),
	}).WithError(errory.APIErrors.New(errorMessage)).Error("problem handling request")
}
