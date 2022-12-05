//  +build unitTests

package service_test

import (
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Test metric service test", func() {
	var mockController *gomock.Controller
	var mockIHappinessMetricProvider *provider.MockIHappinessMetricProvider
	var happinessMetricService *service.HappinessMetricService
	logger, logHook := logrustest.NewNullLogger()
	var metric model.HappinessMetric
	var userContext auth.UserContext

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockIHappinessMetricProvider = provider.NewMockIHappinessMetricProvider(mockController)

		userContext = auth.UserContext{
			ID:     3,
			Cookie: "test cookie",
		}

		t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
		metric = model.HappinessMetric{
			ID:        12,
			UserID:    34,
			OrgID:     2,
			Happiness: 2,
			Safety:    3,
			Date:      t,
		}

		happinessMetricService = &service.HappinessMetricService{
			Provider: mockIHappinessMetricProvider,
			Log:      logger,
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("Create(metric *model.HappinessMetric)", func() {
		Context("When provider creates correctly new metric", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().CreateHappinessMetric(&metric).Times(1).Return(nil)
				err := happinessMetricService.Create(&userContext, &metric)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When provider does not create new metric and return an error", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().CreateHappinessMetric(&metric).Times(1).Return(errory.ProviderErrors.New("Provider errory"))

				err := happinessMetricService.Create(&userContext, &metric)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Update(metric *model.HappinessMetric)", func() {
		Context("When provider updates correctly a metric", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().UpdateHappinessMetric(&metric).Times(1).Return(nil)

				err := happinessMetricService.Update(&userContext, &metric)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When provider does not update a metric and return an error", func() {
			It("Should return an error", func() {
				mockIHappinessMetricProvider.EXPECT().UpdateHappinessMetric(&metric).Times(1).Return(errory.ProviderErrors.New("Provider errory"))

				err := happinessMetricService.Update(&userContext, &metric)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Get(id int64)", func() {
		idToFind := int64(5)
		searchedMetric := metric
		Context("When provider finds correct metric", func() {
			It("Should not return an error", func() {
				searchedMetric.ID = idToFind
				mockIHappinessMetricProvider.EXPECT().GetHappinessMetric(idToFind).Times(1).Return(&searchedMetric, nil)

				resultMetric, err := happinessMetricService.Get(idToFind)

				Expect(err).NotTo(HaveOccurred())
				Expect(idToFind).To(Equal(resultMetric.ID))
			})
		})
		Context("When provider DOES NOT find correct metric", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().GetHappinessMetric(idToFind).Times(1).Return(nil, errory.NotFoundErrors.New("Happiness metric not found"))

				resultMetric, err := happinessMetricService.Get(idToFind)

				Expect(err).To(HaveOccurred())
				Expect(resultMetric).To(BeNil())
			})
		})
	})

	Describe("Delete(id int64)", func() {
		metricID := int64(5)

		Context("When everything goes well", func() {
			It("Should not return an error", func() {
				metric.ID = metricID
				mockIHappinessMetricProvider.EXPECT().DeleteHappinessMetric(metricID).Times(1).Return(nil)

				err := happinessMetricService.Delete(&userContext, metricID)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When error is returned by one of functions", func() {
			var expErr error
			BeforeEach(func() {
				expErr = errory.ProviderErrors.New("test error")
			})
			It("returns the error", func() {
				metric.ID = metricID
				mockIHappinessMetricProvider.EXPECT().DeleteHappinessMetric(metricID).Times(1).Return(expErr)

				err := happinessMetricService.Delete(&userContext, metricID)

				Expect(err).To(HaveOccurred())
				AssertErr(err, expErr)
			})
		})
	})

	Describe("Get all for user", func() {
		idToFind := int64(5)
		searchedMetric := []*model.HappinessMetric{&metric}
		const (
			orgID  int64 = 2
			userID int64 = 1
		)
		Context("When provider finds correct metric", func() {
			It("Should not return an error", func() {
				searchedMetric[0].ID = idToFind
				mockIHappinessMetricProvider.EXPECT().GetAllHappinessMetricsForUser(orgID, userID).Times(1).Return(searchedMetric, nil)

				resultMetric, err := happinessMetricService.GetAllHappinessMetricsForUser(orgID, userID)

				Expect(err).NotTo(HaveOccurred())
				Expect(idToFind).To(Equal(resultMetric[0].ID))
			})
		})
		Context("When provider DOES NOT find correct metric", func() {
			It("Should return an error", func() {
				mockIHappinessMetricProvider.EXPECT().GetAllHappinessMetricsForUser(orgID, userID).Times(1).Return(nil, errory.NotFoundErrors.New("Happiness metric not found"))

				resultMetric, err := happinessMetricService.GetAllHappinessMetricsForUser(orgID, userID)

				Expect(err).To(HaveOccurred())
				Expect(resultMetric).To(BeNil())
			})
		})
	})

	Describe("Get all for team", func() {
		idToFind := int64(5)
		searchedMetric := []*model.HappinessMetric{&metric}
		const (
			orgID int64 = 2
		)
		Context("When provider finds correct metric", func() {
			It("Should not return an error", func() {
				searchedMetric[0].ID = idToFind
				mockIHappinessMetricProvider.EXPECT().GetAllHappinessMetricsForTeam(orgID).Times(1).Return(searchedMetric, nil)

				resultMetric, err := happinessMetricService.GetAllHappinessMetricsForTeam(orgID)

				Expect(err).NotTo(HaveOccurred())
				Expect(idToFind).To(Equal(resultMetric[0].ID))
			})
		})
		Context("When provider DOES NOT find correct metric", func() {
			It("Should return an error", func() {
				mockIHappinessMetricProvider.EXPECT().GetAllHappinessMetricsForTeam(orgID).Times(1).Return(nil, errory.NotFoundErrors.New("Happiness metric not found"))

				resultMetric, err := happinessMetricService.GetAllHappinessMetricsForTeam(orgID)

				Expect(err).To(HaveOccurred())
				Expect(resultMetric).To(BeNil())
			})
		})
	})

	Describe("Save average", func() {
		const (
			orgID int64 = 2
		)
		Context("When provider can save the average", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().SaveTeamAverage(orgID).Times(1).Return(int64(5), nil)

				result, err := happinessMetricService.SaveTeamAverage(orgID)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(int64(5)))
			})
		})
		Context("When provider can not save the average", func() {
			It("Should return an error", func() {
				mockIHappinessMetricProvider.EXPECT().SaveTeamAverage(orgID).Times(1).Return(int64(0), errory.NotFoundErrors.New("Not found"))

				result, err := happinessMetricService.SaveTeamAverage(orgID)

				Expect(err).To(HaveOccurred())
				Expect(result).To(Equal(int64(0)))
			})
		})
	})

	Describe("Get missing", func() {
		users := []*model.UserMissingInput{{
			UserID: 2,
			Login:  "x-man",
		}}
		const (
			orgID int64 = 2
		)
		Context("When provider finds missing users", func() {
			It("Should not return an error", func() {
				mockIHappinessMetricProvider.EXPECT().GetUsersMissingInput(orgID).Times(1).Return(users, nil)

				result, err := happinessMetricService.GetUsersMissingInput(orgID)

				Expect(err).NotTo(HaveOccurred())
				Expect(result[0].UserID).To(Equal(int64(2)))
				Expect(result[0].Login).To(Equal("x-man"))
			})
		})
		Context("When provider DOES NOT find missing users", func() {
			It("Should return an error", func() {
				mockIHappinessMetricProvider.EXPECT().GetUsersMissingInput(orgID).Times(1).Return(nil, errory.NotFoundErrors.New("Not found"))

				result, err := happinessMetricService.GetUsersMissingInput(orgID)

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})
})
