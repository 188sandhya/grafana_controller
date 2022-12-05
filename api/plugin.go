package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	gModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	pService "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service/plugin"
	"github.com/sirupsen/logrus"
)

type PluginAPI struct {
	Plugin              pService.IPluginService
	OrganizationService service.IOrganizationService
	Log                 logrus.FieldLogger
}

// @Summary Enable OMA Plugin
// @Description Enables and Initialize OMA Plugin for given Organization
// @Tags plugin
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Param skipds query string false "skip DS creation, default false"
// @Success 200 {object} api.ID
// @Router /plugin/{id} [post]
func (api *PluginAPI) Enable(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot extract user context").Create(), api.Log)
		return
	}

	org, err := api.getOrganization(c)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot get organization").Create(), api.Log)
		return
	}

	skip, err := strconv.ParseBool(c.Query("skipds"))
	if err != nil {
		skip = false
	}
	if err = api.Plugin.EnablePluginWithDataSources(org, userContext.Cookie, skip); err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot enable plugin").Create(), api.Log)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": org.ID,
	})
}

func (api *PluginAPI) getOrganization(c *gin.Context) (*gModel.Organization, error) {
	orgID, err := GetIDParam(c)
	if err != nil {
		return nil, err
	}

	return api.OrganizationService.GetOrganizationByID(orgID)
}
