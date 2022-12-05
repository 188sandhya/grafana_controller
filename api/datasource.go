//nolint:dupl
package api

import (
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/sirupsen/logrus"
)

type DatasourceAPI struct {
	DatasourceService service.IDatasourceService
	Log               logrus.FieldLogger
}

// @Summary Get Datasource
// @Description Returns Datasource
// @Tags datasources
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Datasource ID"
// @Success 200 {object} grafana.Datasource
// @Router /datasource/{id} [get]
func (api *DatasourceAPI) Get(c *gin.Context) {
	datasourceID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get datasource").Create(), api.Log)
		return
	}

	datasource, err := api.DatasourceService.GetDatasourceByID(datasourceID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get datasource").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, datasource)
}
