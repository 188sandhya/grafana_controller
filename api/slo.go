//nolint:dupl
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"context"

	"github.com/shopspring/decimal"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/ctx"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	v "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
	"github.com/sirupsen/logrus"
)

type SloAPI struct {
	SloService service.ISloService
	Validator  v.ISLOValidator
	Log        logrus.FieldLogger
}

// @Summary Add SLO
// @Description Creates SLO
// @Tags slos
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param slo body model.Slo true "neither id nor creation_date is checked"
// @Success 201 {object} api.ID
// @Router /slo [post]
func (api *SloAPI) Create(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create SLO").Create(), api.Log)
		return
	}

	slo, err := extractAndValidateSlo(c, true, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create SLO").Create(), api.Log)
		return
	}
	if err := api.SloService.Create(userContext, slo); err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create SLO").Create(), api.Log)
		return
	}
	setIDResponse(http.StatusCreated, slo.ID, c)
}

// @Summary Update SLO
// @Description Updates SLO
// @Tags slos
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Slo ID"
// @Param slo body model.Slo true "SLO"
// @Success 200 {object} api.ID
// @Router /slo/{id} [put]
func (api *SloAPI) Update(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update SLO").Create(), api.Log)
		return
	}

	slo, err := extractAndValidateSlo(c, false, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update SLO").Create(), api.Log)
		return
	}

	if err = api.SloService.Update(userContext, slo); err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update SLO").Create(), api.Log)
		return
	}

	setIDResponse(http.StatusOK, slo.ID, c)
}

// @Summary Delete SLO
// @Description Deletes SLO
// @Tags slos
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Slo ID"
// @Success 200 {object} model.Slo
// @Router /slo/{id} [delete]
func (api *SloAPI) Delete(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete SLO").Create(), api.Log)
		return
	}

	sloID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete SLO").Create(), api.Log)
		return
	}

	if err := api.SloService.Delete(userContext, sloID); err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete SLO").Create(), api.Log)
		return
	}

	setIDResponse(http.StatusOK, sloID, c)
}

// @Summary Delete SLO History
// @Description Deletes SLO History
// @Tags slos
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Slo ID"
// @Success 200 {object} model.Slo
// @Router /slo/{id}/history [delete]
func (api *SloAPI) DeleteSloHistory(c *gin.Context) {
	sloID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete SLO History").Create(), api.Log)
		return
	}

	if err := api.SloService.DeleteSloHistory(sloID); err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete SLO History").Create(), api.Log)
		return
	}

	setIDResponse(http.StatusOK, sloID, c)
}

// @Summary Get SLO
// @Description Returns SLO
// @Tags slos
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Slo ID"
// @Success 200 {object} model.Slo
// @Router /slo/{id} [get]
func (api *SloAPI) Get(c *gin.Context) {
	sloID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLO").Create(), api.Log)
		return
	}

	slo, err := api.SloService.Get(sloID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get SLO").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, slo)
}

// @Summary Get detailed and filtered SLOs visible to user
// @Description Returns array of SLOs
// @Tags slos
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param name query string false "filter by SLO name"
// @Param orgName query string false "filter by organization name"
// @Success 200 {array} model.DetailedSlo
// @Router /slo [get]
func (api *SloAPI) GetDetailed(c *gin.Context) {
	sloNameQuery := c.Query("name")
	orgNameQuery := c.Query("orgName")

	slos, err := api.SloService.GetDetailedSlos(sloNameQuery, orgNameQuery)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	c.JSON(http.StatusOK, slos)
}

func extractAndValidateSlo(c *gin.Context, create bool, validator v.ISLOValidator) (*model.Slo, error) {
	var slo model.Slo
	if err := c.ShouldBindBodyWith(&slo, binding.JSON); err != nil {
		return nil, errory.GetValidationError(err, slo)
	}

	if !create {
		sloID, err := GetIDParam(c)
		if err != nil {
			return nil, err
		}
		slo.ID = sloID
	}

	ct := context.WithValue(context.Background(), ctx.Create, create)
	if err := validator.Validate(ct, slo); err != nil {
		return nil, err
	}

	// parsing errors should never occur as fields are already validated
	complianceDecEncoded, err := decimal.NewFromString(slo.ComplianceExpectedAvailability)
	if err != nil {
		return nil, err
	}
	slo.ComplianceExpectedAvailability = complianceDecEncoded.String()

	successRateDecEncoded, err := decimal.NewFromString(slo.SuccessRateExpectedAvailability)
	if err != nil {
		return nil, err
	}
	slo.SuccessRateExpectedAvailability = successRateDecEncoded.String()

	return &slo, nil
}
