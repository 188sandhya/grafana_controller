package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const AbortMessage = "Unexpected Failure"

func Recovery(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if logger != nil {
					r := c.Request
					logger.Errorf("[Recovery] panic recovered:\n%s %s\n%s\n\n%+v\n", r.Method, r.URL.String(), r.UserAgent(), err)
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": AbortMessage})
			}
		}()
		c.Next()
	}
}
