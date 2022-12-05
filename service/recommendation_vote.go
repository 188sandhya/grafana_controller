package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IRecommendationVoteService interface {
	Get(userID int64, orgID *int64) ([]*model.RecommendationVote, error)
	Create(metric *model.RecommendationVote) error
	Delete(metric *model.RecommendationVote) error
}

type RecommendationVoteService struct {
	Provider provider.IRecommendationVoteProvider
	Log      logrus.FieldLogger
}

func (s *RecommendationVoteService) Get(userID int64, orgID *int64) ([]*model.RecommendationVote, error) {
	return s.Provider.GetRecommendationVotes(userID, orgID)
}

func (s *RecommendationVoteService) Create(vote *model.RecommendationVote) error {
	return s.Provider.UpsertRecommendationVote(vote)
}

func (s *RecommendationVoteService) Delete(vote *model.RecommendationVote) error {
	return s.Provider.DeleteRecommendationVote(vote)
}
