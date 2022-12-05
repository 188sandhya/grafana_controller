package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	LogfieldHostname       = "@hostname"
	LogfieldOrganization   = "@vertical"
	LogfieldType           = "@type"
	LogfieldRetention      = "retention"
	LogfieldEpic           = "epic"
	LogfieldEvent          = "event"
	LogfieldServiceName    = "@service-name"
	LogfieldServiceVersion = "@service-version"
)

func LoggerMiddleware(logger logrus.FieldLogger, dt time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		r := c.Request
		timeElapsed := time.Since(startTime)
		ip := r.RemoteAddr
		if index := strings.LastIndex(ip, ":"); index != -1 {
			ip = ip[:index]
		}

		entry := logger.WithFields(logrus.Fields{
			LogfieldEvent:   "request",
			"ip":            ip,
			"method":        r.Method,
			"uri":           r.URL.String(),
			"protocol":      r.Proto,
			"status":        c.Writer.Status(),
			"bytes":         c.Writer.Size(),
			"elapsed":       timeElapsed / time.Millisecond,
			"userAgent":     r.UserAgent(),
			"correlationID": correlationID(c),
		})

		if timeElapsed <= 1000*time.Millisecond {
			entry.Info()
		} else if timeElapsed < dt {
			entry.Warning()
		}
	}
}

func correlationID(c *gin.Context) string {
	v, exists := c.Get(CorrelationIDHeader)
	if exists {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
