package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/sirupsen/logrus"
)

type SolutionsAPI struct {
	Service service.ISolutionsService
	Log     logrus.FieldLogger
}

// @Summary Get Solutions
// @Description Returns all solutions with products
// @Tags fs
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param long query string false "provide a list with products, default false"
// @Param solutionScope query string false "filter by solution scope, default no filter, scope types: \"Unknown\" | \"Elementary\" | \"Differentiating\""
// @Param allowedOnly query string false "provide a list with products which allow share devops metrics, default false"
// @Success 200 {string} string
// @Router /solutions [get]
func (api *SolutionsAPI) GetSolutions(c *gin.Context) {
	long, err := strconv.ParseBool(c.Query("long"))
	if err != nil {
		long = false
	}

	allowedOnly, err := strconv.ParseBool(c.Query("allowedOnly"))
	if err != nil {
		allowedOnly = false
	}

	solutionScope := c.Query("solutionScope")

	solutions, err := api.Service.GetSolutions(long, allowedOnly, solutionScope)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get Solutions").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, solutions)
}
