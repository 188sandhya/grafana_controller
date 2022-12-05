package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/sirupsen/logrus"
)

type SolutionSloAPI struct {
	SloService service.ISolutionSloService
	Log        logrus.FieldLogger
}

// @Summary Get Solution SLO details of org
// @Description Returns attributes of Solution SLO
// @Tags solutionslo
// @Produce json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param orgName query string false "filter by organization name"
// @Success 200 {array} model.SolutionSlo
// @Router /solutionSlo [get]
func (api *SolutionSloAPI) GetSolutionSlo(c *gin.Context) {
	orgNameQuery := c.Query("orgName")
	if orgNameQuery == "" {
		c.String(http.StatusBadRequest, "Param 'orgName' is manadatory!")
		return
	}
	solutionSlo, err := api.SloService.GetSolutionSlo(orgNameQuery)
	if err != nil {
		details := errory.GetErrorDetails(err)
		if details.Status == http.StatusNotFound {
			c.JSON(details.Status, Message{Message: details.Message})
		} else {
			setErrorResponse(c, err, api.Log)
		}
		return
	}
	c.JSON(http.StatusOK, solutionSlo)
}
