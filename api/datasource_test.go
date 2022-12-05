// +build unitTests

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("DatasourceAPI", func() {
	var mockController *gomock.Controller
	var datasourceAPI *DatasourceAPI
	var dataSourceServiceMock *service.MockIDatasourceService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		dataSourceServiceMock = service.NewMockIDatasourceService(mockController)
		datasourceAPI = &DatasourceAPI{
			DatasourceService: dataSourceServiceMock,
			Log:               logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()
		ginEngine.GET("/v1/datasource/:id", datasourceAPI.Get)
		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("Get()", func() {
		const datasourceID int64 = 33
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/datasource/%d", datasourceID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			foundDatasource := grafanaModel.Datasource{
				ID:   datasourceID,
				Name: "test2",
			}
			BeforeEach(func() {
				dataSourceServiceMock.EXPECT().GetDatasourceByID(datasourceID).Times(1).Return(&foundDatasource, nil)
				expErr = nil
			})

			It("returns 200 code with proper message", func() {
				var datasource grafanaModel.Datasource
				err := json.Unmarshal(w.Body.Bytes(), &datasource)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(datasource.ID).To(Equal(datasourceID))
				Expect(datasource.Name).To(Equal("test2"))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when datasource service returns NotFound error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithPayload("datasource", 1).Create()
				dataSourceServiceMock.EXPECT().GetDatasourceByID(datasourceID).Times(1).Return(nil, expErr)
			})
			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get datasource (data not found error; details [datasource: 1])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get datasource, cause: ebt.not_exist: data not found error")
			})
		})

		Context("when datasource service returns unexpected  error", func() {
			BeforeEach(func() {
				someError := errory.ProviderErrors.New("datasource unexepected error")
				expErr = errory.OnGetErrors.Builder().Wrap(someError).Create()
				dataSourceServiceMock.EXPECT().GetDatasourceByID(datasourceID).Times(1).Return(nil, someError)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get datasource (datasource unexepected error)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get datasource, cause: ebt.provider_error: datasource unexepected error")
			})
		})
	})
})
