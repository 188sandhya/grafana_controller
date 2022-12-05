//  +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test Solution Slo service", func() {
	var mockController *gomock.Controller
	var mockFSProvider *provider.MockISolutionSloProvider
	var fsService *service.SolutionSloService

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockFSProvider = provider.NewMockISolutionSloProvider(mockController)

		fsService = &service.SolutionSloService{
			SloProvider: mockFSProvider,
		}

	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("GetSolutionSlo", func() {
		Context("When provider finds data", func() {
			It("Should not return an error", func() {
				data := model.SolutionSlo{
					OrgID:          1,
					NoCriticalSlos: 3,
					DashboardPath:  "/some/path",
				}
				mockFSProvider.EXPECT().GetSolutionSlo("errorbudget").Times(1).Return(&data, nil)

				result, err := fsService.GetSolutionSlo("errorbudget")

				Expect(err).ToNot(HaveOccurred())
				Expect(result.OrgID).To(Equal(int64(1)))
				Expect(result.NoCriticalSlos).To(Equal(int64(3)))
				Expect(result.DashboardPath).To(Equal("/some/path"))
			})
		})

		Context("When provider DOES NOT find data", func() {
			It("Should return an error", func() {
				mockFSProvider.EXPECT().GetSolutionSlo("").Times(1).Return(nil, errory.NotFoundErrors.New("Data not found"))

				result, err := fsService.GetSolutionSlo("")

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})
	})
})
