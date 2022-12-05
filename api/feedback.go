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

type FeedbackAPI struct {
	Service   service.IFeedbackService
	Validator validator.IFeedbackValidator
	Log       logrus.FieldLogger
}

// @Summary Create Feedback
// @Description Returns Feedback
// @Tags feedback
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param feedback body model.Feedback true "Feedback"
// @Success 201 {object} api.ID
// @Router /feedback/ [post]
func (api *FeedbackAPI) Create(c *gin.Context) {
	feedback, err := extractAndValidateFeedback(c, true, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create feedback").Create(), api.Log)
		return
	}

	if err := api.Service.Create(feedback); err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage("Cannot create feedback").Create(), api.Log)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": feedback.ID,
	})
}

// @Summary Get Feedback
// @Description Returns Feedback
// @Tags feedback
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param id path int true "Org ID"
// @Success 200 {object} api.ID
// @Router /feedback/{id} [get]
func (api *FeedbackAPI) Get(c *gin.Context) {
	orgID, err := GetIDParam(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get feedback").Create(), api.Log)
		return
	}

	feedbacks, err := api.Service.GetByOrgID(orgID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get feedback").Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, feedbacks)
}

func extractAndValidateFeedback(c *gin.Context, create bool, v validator.IFeedbackValidator) (*model.Feedback, error) {
	var feedback model.Feedback
	if err := c.ShouldBindBodyWith(&feedback, binding.JSON); err != nil {
		return nil, errory.GetValidationError(err, feedback)
	}

	ct := context.WithValue(context.Background(), ctx.Create, create)

	if err := v.Validate(ct, feedback); err != nil {
		return nil, err
	}

	return &feedback, nil
}
