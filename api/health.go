package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/sirupsen/logrus"
)

type HealthAPI struct {
	HealthService service.IHealthService
	Log           logrus.FieldLogger
}

func (api *HealthAPI) HealthCheck(c *gin.Context) {
	err := api.HealthService.CheckHealth()
	if err == nil {
		c.String(http.StatusNoContent, "")
	} else {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Database healthcheck failed").Create(), api.Log)
	}
}
