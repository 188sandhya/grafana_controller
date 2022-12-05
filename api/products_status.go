package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/sirupsen/logrus"
)

type ProductsStatusAPI struct {
	Service service.IProductsStatusService
	Log     logrus.FieldLogger
}

// @Summary Get Products Status
// @Description Returns all products status
// @Tags fs
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Success 200 {string} string
// @Router /products_status [get]
func (api *ProductsStatusAPI) GetProductsStatus(c *gin.Context) {
	productsStatus, err := api.Service.GetProductsStatus()
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get Products Status").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, productsStatus)
}
