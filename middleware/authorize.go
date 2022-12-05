package middleware

import (
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Authorizer func(entityID, userID int64, requiredPermission auth.Permission) (bool, error)

type IAuthorizer interface {
	AuthorizeForSLO(entityID, userID int64, requiredRole auth.Permission) (bool, error)
	AuthorizeForDatasource(entityID, userID int64, requiredRole auth.Permission) (bool, error)
	AuthorizeForOrganization(entityID, userID int64, requiredRole auth.Permission) (bool, error)
	AuthorizeForTeam(entityID, userID int64, requiredRole auth.Permission) (bool, error)
	AuthorizeForHappinessMetric(entityID, userID int64, requiredRole auth.Permission) (bool, error)
}

func Authorize(idSelector IDSelector, authorizer Authorizer, requiredPermission auth.Permission, log logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get(UserContextContextKey)
		if !exists {
			log.Errorf("no userID found")
			authorizationError(c)
			return
		}

		user := u.(*auth.UserContext)

		id, err := idSelector(c)
		if err != nil {
			details := errory.GetErrorDetails(err)
			log.Errorf("ID param error: %s", err)
			c.AbortWithStatusJSON(details.Status, gin.H{
				"message": details.Message,
			})
			return
		}

		authorized, err := authorizer(id, user.ID, requiredPermission)
		if err != nil {
			log.Errorf("authorize error: %s", err)
			authorizationError(c)
			return
		}
		if !authorized {
			log.Errorf("not authorized: %s", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Not authorized",
			})
			return
		}

		c.Next()
	}
}

func authorizationError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"message": "Authorization error",
	})
}
