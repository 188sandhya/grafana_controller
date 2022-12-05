package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/sirupsen/logrus"
)

type SDAAPI struct {
	Service service.ISDAService
	Log     logrus.FieldLogger
}

// @Summary Get SDA configuration
// @Description Returns full SDA configuration for all organizations
// @Tags sda
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Success 200 {string} string
// @Router /sda [get]
func (api *SDAAPI) GetConfig(c *gin.Context) {
	config, err := api.Service.GetConfig()
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SDA configuration").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, config)
}
