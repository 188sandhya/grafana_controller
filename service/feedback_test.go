//  +build unitTests

package service_test

import (
	"time"

	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Feedback service test", func() {
	var mockController *gomock.Controller
	var mockIFeedbackProvider *provider.MockIFeedbackProvider
	var feedbackService *service.FeedbackService
	logger, logHook := logrustest.NewNullLogger()
	var fb model.Feedback

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIFeedbackProvider = provider.NewMockIFeedbackProvider(mockController)

		t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
		fb = model.Feedback{
			ID:              12,
			ReceivingUserID: 66,
			GivingUserID:    22,
			OrgID:           12,
			FeedbackDate:    t,
		}

		feedbackService = &service.FeedbackService{
			Provider: mockIFeedbackProvider,
			Log:      logger,
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("Create", func() {
		Context("When provider creates correctly new feedback", func() {
			It("Should not return an error", func() {
				mockIFeedbackProvider.EXPECT().CreateFeedback(&fb).Times(1).Return(nil)
				err := feedbackService.Create(&fb)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When provider does not create new feedback and returns an error", func() {
			It("Should return an error", func() {
				mockIFeedbackProvider.EXPECT().CreateFeedback(&fb).Times(1).Return(errory.ProviderErrors.New("Provider error"))

				err := feedbackService.Create(&fb)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetByOrgID", func() {
		idToFind := int64(12)
		var searchedFeedbacks []*model.Feedback
		searchedFeedbacks = append(searchedFeedbacks, &fb)
		Context("When provider finds correct feedbacks", func() {
			It("Should not return an error", func() {
				mockIFeedbackProvider.EXPECT().GetFeedbacksByOrgID(idToFind).Times(1).Return(searchedFeedbacks, nil)

				result, err := feedbackService.GetByOrgID(idToFind)

				Expect(err).NotTo(HaveOccurred())
				Expect(idToFind).To(Equal(result[0].ID))
			})
		})
		Context("When provider does not find feedbacks", func() {
			It("Should not return an error", func() {
				mockIFeedbackProvider.EXPECT().GetFeedbacksByOrgID(idToFind).Times(1).Return(nil, errory.NotFoundErrors.New("Feedback not found"))

				result, err := feedbackService.GetByOrgID(idToFind)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

	})
})
