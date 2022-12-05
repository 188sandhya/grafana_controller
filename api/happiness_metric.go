//nolint:dupl
package api

import (
	"context"
	"net/http"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/ctx"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
)

type HappinessMetricAPI struct {
	Service   service.IHappinessMetricService
	Validator validator.IHappinessMetricValidator
	Log       logrus.FieldLogger
}

// @Summary Create Happiness Metric
// @Description Returns Happiness Metric
// @Tags happiness metrics
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param happinessMetric body model.HappinessMetric true "Happiness Metric"
// @Success 201 {object} api.ID
// @Router /happiness_metric/ [post]
func (api *HappinessMetricAPI) Create(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create happiness metric").Create(), api.Log)
		return
	}

	happinessMetric, err := extractAndValidateHappinessMetric(c, true, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create happiness metric").Create(), api.Log)
		return
	}

	if err := api.Service.Create(userContext, happinessMetric); err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create happiness metric").Create(), api.Log)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": happinessMetric.ID,
	})
}

// @Summary Update Happiness Metric
// @Description Updates Happiness Metric
// @Tags happiness metrics
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Happiness Metric ID"
// @Param happinessMetric body model.HappinessMetric true "Happiness Metric"
// @Success 200 {object} api.ID
// @Router /happiness_metric/{id} [put]
func (api *HappinessMetricAPI) Update(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update happiness metric").Create(), api.Log)
		return
	}

	happinessMetric, err := extractAndValidateHappinessMetric(c, false, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update happiness metric").Create(), api.Log)
		return
	}

	if err = api.Service.Update(userContext, happinessMetric); err != nil {
		setErrorResponse(c, errory.OnUpdateErrors.Builder().Wrap(err).WithMessage("Cannot update happiness metric").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": happinessMetric.ID,
	})
}

// @Summary Delete Happiness Metric
// @Description Deletes Happiness Metric
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Happiness Metric ID"
// @Success 200 {object} api.ID
// @Router /happiness_metric/{id} [delete]
func (api *HappinessMetricAPI) Delete(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete happiness metric").Create(), api.Log)
		return
	}

	metricID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete happiness metric").Create(), api.Log)
		return
	}

	if err := api.Service.Delete(userContext, metricID); err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage("Cannot delete happiness metric").Create(), api.Log)
		return
	}

	setIDResponse(http.StatusOK, metricID, c)
}

// @Summary Get Happiness Metric
// @Description Returns Happiness Metric
// @Tags happiness metrics
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Happiness Metric ID"
// @Success 200 {object} api.ID
// @Router /happiness_metric/{id} [get]
func (api *HappinessMetricAPI) Get(c *gin.Context) {
	metricID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get happiness metric").Create(), api.Log)
		return
	}

	metric, err := api.Service.Get(metricID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get happiness metric").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, metric)
}

func extractAndValidateHappinessMetric(c *gin.Context, create bool, v validator.IHappinessMetricValidator) (*model.HappinessMetric, error) {
	var happinessMetric model.HappinessMetric
	if err := c.ShouldBindBodyWith(&happinessMetric, binding.JSON); err != nil {
		return nil, errory.GetValidationError(err, happinessMetric)
	}

	if !create {
		metricID, err := GetIDParam(c)
		if err != nil {
			return nil, err
		}
		happinessMetric.ID = metricID
	}

	ct := context.WithValue(context.Background(), ctx.Create, create)

	if err := v.Validate(ct, happinessMetric); err != nil {
		return nil, err
	}

	return &happinessMetric, nil
}
