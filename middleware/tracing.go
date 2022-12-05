package middleware

import (
	"github.com/gin-gonic/gin"
	gouuid "github.com/satori/go.uuid"
)

const CorrelationIDHeader = "X-Correlation-ID"

func Tracing(c *gin.Context) {
	correlationID := c.Request.Header.Get(CorrelationIDHeader)
	if correlationID == "" {
		c.Set(CorrelationIDHeader, gouuid.NewV4().String())
	} else {
		c.Set(CorrelationIDHeader, correlationID)
	}
	c.Next()
}
