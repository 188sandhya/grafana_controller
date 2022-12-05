//nolint:dupl
package api

import (
	"net/http"
	"strconv"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
)

type RecommendationVoteAPI struct {
	Service   service.IRecommendationVoteService
	Validator validator.ITranslatedValidator
	Log       logrus.FieldLogger
}

const (
	cannotGetVotes   = "Cannot get votes"
	cannotCreateVote = "Cannot create vote"
	cannotDeleteVote = "Cannot delete vote"
)

// @Summary Get Recommendation Votes for current user
// @Description Returns array of Recommendation Votes
// @Tags vote
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param orgID query int false "filter by organization name"
// @Success 200 {array} model.RecommendationVote
// @Router /recommendation_vote [get]
func (api *RecommendationVoteAPI) Get(c *gin.Context) {
	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage(cannotGetVotes).Create(), api.Log)
		return
	}

	var orgID *int64
	if c.Query("orgId") != "" {
		parsed, err := strconv.ParseInt(c.Query("orgId"), 10, 64)
		if err != nil {
			setErrorResponse(c, errory.ParseErrors.Builder().Wrap(err).WithMessage("orgId parameter cannot be parsed").Create(), api.Log)
			return
		}
		orgID = &parsed
	}

	votes, err := api.Service.Get(userContext.ID, orgID)
	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage(cannotGetVotes).Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, votes)
}

// @Summary Create Recommendation Vote
// @Description Returns Vote ID
// @Tags vote
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param vote body model.RecommendationVote true "RecommendationVote"
// @Success 201 {object} api.ID
// @Router /recommendation_vote [post]
func (api *RecommendationVoteAPI) Create(c *gin.Context) {
	vote, err := extractAndValidateRecommendationVote(c, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage(cannotCreateVote).Create(), api.Log)
		return
	}

	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage(cannotCreateVote).Create(), api.Log)
		return
	}
	vote.UserID = userContext.ID

	if err := api.Service.Create(vote); err != nil {
		setErrorResponse(c, errory.OnCreateErrors.Builder().Wrap(err).WithMessage(cannotCreateVote).Create(), api.Log)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": vote.ID,
	})
}

// @Summary Delete Recommendation Vote
// @Description Returns Vote ID
// @Tags vote
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token | Basic auth | Cookie grafana_session"
// @Param vote body model.RecommendationVote true "RecommendationVote"
// @Success 200 {object} api.ID
// @Router /recommendation_vote [delete]
func (api *RecommendationVoteAPI) Delete(c *gin.Context) {
	vote, err := extractAndValidateRecommendationVote(c, api.Validator)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage(cannotDeleteVote).Create(), api.Log)
		return
	}

	userContext, err := GetUserContext(c)
	if err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage(cannotDeleteVote).Create(), api.Log)
		return
	}
	vote.UserID = userContext.ID

	if err := api.Service.Delete(vote); err != nil {
		setErrorResponse(c, errory.OnDeleteErrors.Builder().Wrap(err).WithMessage(cannotDeleteVote).Create(), api.Log)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": vote.ID,
	})
}

func extractAndValidateRecommendationVote(c *gin.Context, v validator.ITranslatedValidator) (*model.RecommendationVote, error) {
	var vote model.RecommendationVote
	if err := c.ShouldBindBodyWith(&vote, binding.JSON); err != nil {
		return nil, errory.GetValidationError(err, vote)
	}

	if err := v.Validate(c, vote); err != nil {
		return nil, err
	}

	return &vote, nil
}
