//go:build unitTests
// +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Test Solutions service", func() {
	var mockController *gomock.Controller
	var mockFSProvider *provider.MockISolutionsProvider
	var fsService *service.SolutionsService
	logger, logHook := logrustest.NewNullLogger()

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockFSProvider = provider.NewMockISolutionsProvider(mockController)

		fsService = &service.SolutionsService{
			Provider: mockFSProvider,
			Log:      logger,
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("GetSolutions", func() {
		Context("When provider finds data", func() {
			It("Should not return an error", func() {
				data := []*model.Solution{
					{
						ID:          1,
						Name:        "M|COMPANION",
						ProductID:   new(int64),
						ProductName: "M|COMPANION",
					},
					{
						ID:          7,
						Name:        "M|CREDIT",
						ProductID:   new(int64),
						ProductName: "Customer Credit Limit Management System",
					},
				}
				mockFSProvider.EXPECT().GetSolutions(true, false, "").Times(1).Return(data, nil)

				result, err := fsService.GetSolutions(true, false, "")

				Expect(err).ToNot(HaveOccurred())
				Expect(result[0].ID).To(Equal(int64(1)))
				Expect(result[0].Name).To(Equal("M|COMPANION"))
			})
		})

		Context("When provider DOES NOT find data", func() {
			It("Should return an error", func() {
				mockFSProvider.EXPECT().GetSolutions(false, false, "").Times(1).Return(nil, errory.NotFoundErrors.New("Data not found"))

				result, err := fsService.GetSolutions(false, false, "")

				Expect(err).To(HaveOccurred())
				Expect(len(result)).To(Equal(0))
			})
		})
	})
})
