//  +build unitTests

package service_test

import (
	"github.com/golang/mock/gomock"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Test SDA service", func() {
	var mockController *gomock.Controller
	var mockISDAProvider *provider.MockISDAProvider
	var sdaService *service.SDAService
	logger, logHook := logrustest.NewNullLogger()

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockISDAProvider = provider.NewMockISDAProvider(mockController)

		sdaService = &service.SDAService{
			Provider: mockISDAProvider,
			Log:      logger,
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("GetConfig", func() {
		Context("When provider finds configuration", func() {
			It("Should not return an error", func() {
				jsonData := `[{"errorbudget": {"git": [{"url": "git@github.com:metro-digital-inner-source/errorbudget-grafana-controller.git"},{"url": "git@github.com:metro-digital-inner-source/errorbudget-sda.git"},{"url": "git@github.com:metro-digital-inner-source/errorbudget-susie.git"},{"url": "git@github.com:metro-digital-inner-source/errorbudget-autoslo.git"}],"cicd": [{"url": "https://errorbudget.cip.metronom.com","jobs_scoring_whitelist": ["autoslo-prod","errorbudget-v2-prod","sda-prod","susie-prod"],"jobs_scraping_whitelist": []}],"jira": [{"url": "https://jira.metrosystems.net","project": "OPEB"}]}}]`
				mockISDAProvider.EXPECT().GetConfig().Times(1).Return(jsonData, nil)

				result, err := sdaService.GetConfig()

				Expect(err).ToNot(HaveOccurred())
				Expect(result[0]).To(HaveKey("errorbudget"))
			})
		})

		Context("When provider DOES NOT find configuration", func() {
			It("Should return an error", func() {
				mockISDAProvider.EXPECT().GetConfig().Times(1).Return("", errory.NotFoundErrors.New("Configuration not found"))

				result, err := sdaService.GetConfig()

				Expect(err).To(HaveOccurred())
				Expect(len(result)).To(Equal(0))
			})
		})
	})
})
