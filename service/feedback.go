package service

import (
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

type IFeedbackService interface {
	Create(metric *model.Feedback) error
	GetByOrgID(id int64) ([]*model.Feedback, error)
}

type FeedbackService struct {
	Provider provider.IFeedbackProvider
	Log      logrus.FieldLogger
}

func (s *FeedbackService) Create(feedback *model.Feedback) error {
	return s.Provider.CreateFeedback(feedback)
}

func (s *FeedbackService) GetByOrgID(id int64) ([]*model.Feedback, error) {
	return s.Provider.GetFeedbacksByOrgID(id)
}
