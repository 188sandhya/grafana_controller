package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/sirupsen/logrus"
)

const UserContextContextKey = "UserContext"

// Authenticate user and add id to context for handler
func Authenticate(authenticator auth.IAuthenticator, log logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := authenticator.Authenticate(c)
		if err != nil {
			details := errory.GetErrorDetails(errory.AuthErrors.New("Not authorized"))
			log.Errorf("Not authorized: %s", err.Error())
			c.AbortWithStatusJSON(details.Status, gin.H{
				"message": details.Message,
			})
			return
		}
		log.Debugf("Authenticated user: %d", user.ID)
		c.Set(UserContextContextKey, user)
		c.Next()
	}
}
