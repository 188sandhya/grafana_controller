//nolint:dupl
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"

	"context"

	"github.com/sirupsen/logrus"
)

type OrgAPI struct {
	SloService        service.ISloService
	OrgService        service.IOrganizationService
	DatasourceService service.IDatasourceService
	HappinessService  service.IHappinessMetricService
	Validator         validator.ITranslatedValidator
	Log               logrus.FieldLogger
}

// @Summary Get SLOs
// @Description Returns all SLOs for Organization
// @Tags organizations
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {array} model.Slo
// @Router /org/{id}/slo [get]
func (api *OrgAPI) GetSlos(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLOs for organization").Create(), api.Log)
		return
	}

	slos, err := api.SloService.GetByOrgID(orgID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLOs for organization").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, slos)
}

// @Summary Get Datasources
// @Description Returns all datasources for Organization
// @Tags organizations
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {array} grafana.Datasource
// @Router /org/{id}/datasource [get]
func (api *OrgAPI) GetDatasources(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get datasources for organization").Create(), api.Log)
		return
	}

	slos, err := api.DatasourceService.GetDatasourcesByOrganizationID(orgID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get datasources for organization").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, slos)
}

// @Summary Get filtered SLOs
// @Description Returns all SLOs for Organization matching filter params
// @Tags organizations
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Param sloQueryParams body model.SloQueryParams true "Slo Query Params"
// @Success 200 {array} model.Slo
// @Router /org/{id}/slo [post]
func (api *OrgAPI) FindSlos(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLOs").Create(), api.Log)
		return
	}

	var sloQuery model.SloQueryParams
	if err = c.ShouldBindBodyWith(&sloQuery, binding.JSON); err != nil {
		setErrorResponse(c, errory.OnGetErrors.
			Builder().
			Wrap(errory.GetValidationError(err, sloQuery)).
			WithMessage("Cannot get SLOs").
			Create(), api.Log)
		return
	}

	sloQuery.OrgID = orgID

	if err = api.Validator.Validate(context.Background(), sloQuery); err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLOs").Create(), api.Log)
		return
	}

	slos, err := api.SloService.FindSlos(&sloQuery)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLOs").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, slos)
}

// @Summary Get all happiness metrics created by a user, filtered by orgId
// @Description Returns array of happiness metrics
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {array} model.HappinessMetric
// @Router /org/{id}/user_happiness [get]
func (api *OrgAPI) GetAllHappinessMetricsForUser(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get happiness metrics").Create(), api.Log)
		return
	}

	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get happiness metrics").Create(), api.Log)
		return
	}

	metrics, err := api.HappinessService.GetAllHappinessMetricsForUser(orgID, userContext.ID)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// @Summary Get all happiness metrics created by a team, filtered by orgId
// @Description Returns array of happiness metrics
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {array} model.HappinessMetric
// @Router /org/{id}/team_happiness [get]
func (api *OrgAPI) GetAllHappinessMetricsForTeam(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get happiness metrics").Create(), api.Log)
		return
	}

	metrics, err := api.HappinessService.GetAllHappinessMetricsForTeam(orgID)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// @Summary Store team's happiness metric average
// @Description Store team's happiness metric average
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {object} api.ID
// @Router /org/{id}/team_happiness/average [post]
func (api *OrgAPI) SaveTeamAverage(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create happiness metrics average").Create(), api.Log)
		return
	}

	id, err := api.HappinessService.SaveTeamAverage(orgID)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	setIDResponse(http.StatusCreated, id, c)
}

// @Summary Get users whose input is missing for a current period
// @Description Returns array of users' names
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {array} model.UserMissingInput
// @Router /org/{id}/team_happiness/missing [get]
func (api *OrgAPI) GetUsersMissingInput(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get missing users inputs").Create(), api.Log)
		return
	}

	missingInput, err := api.HappinessService.GetUsersMissingInput(orgID)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	c.JSON(http.StatusOK, missingInput)
}

// @Summary Get Organization
// @Description Returns organization details
// @Tags organizations
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Organization ID"
// @Success 200 {object} grafana.Organization
// @Router /org/{id} [get]
func (api *OrgAPI) GetOrg(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get organization details").Create(), api.Log)
		return
	}

	org, err := api.OrgService.GetOrganizationByID(orgID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get organization details").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, org)
}
