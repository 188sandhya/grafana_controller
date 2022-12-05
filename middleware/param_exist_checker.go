package middleware

import (
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
)

func CheckIfExist(idSelector IDSelector, paramExistCheckService service.IParamExistCheckService, objectType service.Type) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := idSelector(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}
		switch objectType {
		case service.Organization:
			err = paramExistCheckService.CheckOrgByID(id)
		case service.Slo:
			err = paramExistCheckService.CheckSloByID(id)
		}
		if err != nil {
			details := errory.GetErrorDetails(err)
			c.AbortWithStatusJSON(details.Status, gin.H{
				"message": details.Message,
			})
			return
		}
		c.Next()
	}
}
