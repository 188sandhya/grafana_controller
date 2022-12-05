// +build unitTests

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/ctx"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("HappinessMetricAPI", func() {
	var mockController *gomock.Controller
	var happinessMetricAPI *HappinessMetricAPI
	var happinessMetricValidatorMock *validator.MockIHappinessMetricValidator
	var happinessMetricServiceMock *service.MockIHappinessMetricService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error
	var userContext auth.UserContext

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		happinessMetricServiceMock = service.NewMockIHappinessMetricService(mockController)
		happinessMetricValidatorMock = validator.NewMockIHappinessMetricValidator(mockController)
		happinessMetricAPI = &HappinessMetricAPI{
			Validator: happinessMetricValidatorMock,
			Service:   happinessMetricServiceMock,
			Log:       logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContextMiddleware := func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		happinessMetricRoutes := ginEngine.Group("/v1/happiness_metric")
		{
			happinessMetricRoutes.POST("", userContextMiddleware, happinessMetricAPI.Create)
			happinessMetricRoutes.PUT("/:id", userContextMiddleware, happinessMetricAPI.Update)
			happinessMetricRoutes.GET("/:id", happinessMetricAPI.Get)
			happinessMetricRoutes.DELETE("/:id", userContextMiddleware, happinessMetricAPI.Delete)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("When url ID param couldn't be parsed", func() {
		var requestMethod string
		const invalidID string = "abc"
		const requestBody string = `{
			"userId": 12,
			"orgId": 2,
			"happiness": 2,
			"safety": 3,
			"date": "2020-01-07T00:00:00.000Z"
		}`
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest(requestMethod, fmt.Sprintf("/v1/happiness_metric/%s", invalidID), bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})
		Context("when http methods is PUT", func() {
			BeforeEach(func() {
				requestMethod = "PUT"
				expErr = errory.ParseErrors.Builder().WithPayload("metric id", 12).Create()
			})
			It("returns 400 code with proper", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update happiness metric (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_update_error: Cannot update happiness metric, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
		Context("when http methods is GET", func() {
			BeforeEach(func() {
				requestMethod = "GET"
				expErr = errory.ParseErrors.Builder().WithPayload("metric id", 7).Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get happiness metric (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get happiness metric, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
		Context("when http methods is DELETE", func() {
			BeforeEach(func() {
				requestMethod = "DELETE"
				expErr = errory.ParseErrors.Builder().WithPayload("metric id", 66).Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete happiness metric (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete happiness metric, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
	})

	Describe("Create()", func() {
		var (
			err         error
			metric      model.HappinessMetric
			requestBody string
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/v1/happiness_metric", bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			const (
				createdMetricID int64 = 9
				userId          int64 = 10
			)
			BeforeEach(func() {
				requestBody = `{
					"userId": 10,
					"orgId": 2,
					"happiness": 2,
					"safety": 3,
					"date": "2020-01-07T00:00:00.000Z"
				}`
				err = json.Unmarshal([]byte(requestBody), &metric)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				happinessMetricValidatorMock.EXPECT().Validate(ct, metric).Times(1)
				happinessMetricServiceMock.EXPECT().Create(&userContext, &metric).Times(1).Do(func(userContext *auth.UserContext, metric *model.HappinessMetric) {
					metric.ID = createdMetricID
				})
				expErr = nil
			})

			It("returns 201 code and id of created metric", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusCreated))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, createdMetricID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})

		})

		Context("when happinessMetricService returns an error", func() {
			BeforeEach(func() {
				requestBody = `{
					"userId": 10,
					"orgId": 2,
					"happiness": 2,
					"safety": 3,
					"date": "2020-01-07T00:00:00.000Z"
				}`
				someErr := fmt.Errorf("some error")
				expErr = errory.OnCreateErrors.Builder().Wrap(someErr).Create()
				err = json.Unmarshal([]byte(requestBody), &metric)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				happinessMetricValidatorMock.EXPECT().Validate(ct, metric).Times(1)
				happinessMetricServiceMock.EXPECT().Create(&userContext, &metric).Times(1).Return(someErr)
			})

			It("returns 500 code and error message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create happiness metric", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create happiness metric, cause: some error`)
			})
		})

		Context("when userId is not set", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 2,
					"happiness": 2,
					"safety": 3,
					"date": "2020-01-07T00:00:00.000Z"
				}`
				expErr = errory.ValidationErrors.Builder().WithMessage("name must not be null").Create()
				var tmpMetric model.HappinessMetric
				err = json.Unmarshal([]byte(requestBody), &tmpMetric)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				happinessMetricValidatorMock.EXPECT().Validate(ct, tmpMetric).Times(1).Return(expErr)
			})

			It("returns 400 code and proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create happiness metric (name must not be null)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_create_error: Cannot create happiness metric, cause: ebt.validation_error: name must not be null")
			})
		})
	})

	Describe("Update()", func() {
		const metricID int64 = 33
		var requestBody string
		var create = false
		BeforeEach(func() {
			requestBody = `{
				"userId": 1,
				"orgId": 2,
				"happiness": 2,
				"safety": 3,
				"date": "2020-01-07T00:00:00.000Z"
			}`
		})
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("PUT", fmt.Sprintf("/v1/happiness_metric/%d", metricID), bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
		var metric = model.HappinessMetric{
			ID:        metricID,
			UserID:    1,
			OrgID:     2,
			Happiness: 2,
			Safety:    3,
			Date:      t,
		}
		var err error
		Context("when the request succeeds", func() {
			BeforeEach(func() {
				happinessMetricServiceMock.EXPECT().Update(&userContext, &model.HappinessMetric{
					ID:        metricID,
					UserID:    1,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}).Times(1)

				ct := context.WithValue(context.Background(), ctx.Create, create)
				happinessMetricValidatorMock.EXPECT().Validate(ct, model.HappinessMetric{
					ID:        metricID,
					UserID:    1,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}).Times(1)
				expErr = nil
			})

			It("returns 200 code and id of updated metric", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, metricID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when happinessMetricService returns 'NotUnique' error", func() {
			BeforeEach(func() {
				expErr = errory.NotUniqueErrors.Builder().WithMessage("not unique metric").WithPayload("metric name", "name").Create()
				happinessMetricServiceMock.EXPECT().Update(&userContext, &model.HappinessMetric{
					ID:        metricID,
					UserID:    1,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}).Times(1).Return(expErr)

				ct := context.WithValue(context.Background(), ctx.Create, false)
				happinessMetricValidatorMock.EXPECT().Validate(ct, model.HappinessMetric{
					ID:        metricID,
					UserID:    1,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}).Times(1)
			})

			It("returns 409 code", func() {
				Expect(w.Code).To(Equal(http.StatusConflict))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update happiness metric (not unique metric; details [metric name: name])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_update_error: Cannot update happiness metric, cause: ebt.not_unique: not unique metric")
			})
		})
		Context("when happinessMetricService returns an error", func() {
			BeforeEach(func() {
				requestBody = `<xml>{}`
				expErr = errory.ValidationErrors.Builder().WithMessage(`invalid character '<' looking for beginning of value`).Create()
			})

			It("returns 400 code and error message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), `Cannot update happiness metric (error validating data)`, expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_update_error: Cannot update happiness metric, cause: ebt.validation_error: error validating data, cause: invalid character '<' looking for beginning of value`)
			})
		})

		Context("when happinessMetricService returns an error", func() {

			BeforeEach(func() {
				requestBody = `{
					"userId": 1,
					"orgId": 2,
					"happiness": 2,
					"safety": 3,
					"date": "2020-01-07T00:00:00.000Z"
				}`
				someErr := fmt.Errorf("some error")
				expErr = errory.OnUpdateErrors.Wrap(someErr)
				err = json.Unmarshal([]byte(requestBody), &metric)
				ct := context.WithValue(context.Background(), ctx.Create, false)
				happinessMetricValidatorMock.EXPECT().Validate(ct, metric).Times(1)
				happinessMetricServiceMock.EXPECT().Update(&userContext, &metric).Times(1).Return(someErr)
			})

			It("returns 500 code and error message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update happiness metric", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_update_error: Cannot update happiness metric, cause: some error")
			})
		})
	})

	Describe("Get()", func() {
		const (
			metricID int64 = 33
			userID   int64 = 34
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/happiness_metric/%d", metricID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
		Context("when the request succeeds", func() {
			foundMetric := model.HappinessMetric{
				ID:        metricID,
				UserID:    34,
				OrgID:     2,
				Happiness: 2,
				Safety:    3,
				Date:      t,
			}
			BeforeEach(func() {
				expErr = nil
				happinessMetricServiceMock.EXPECT().Get(metricID).Times(1).Return(&foundMetric, nil)
			})

			It("returns 200 code with proper message", func() {
				var metric model.HappinessMetric
				err := json.Unmarshal(w.Body.Bytes(), &metric)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(metric.ID).To(Equal(metricID))
				Expect(metric.UserID).To(Equal(userID))
				Expect(metric.Happiness).To(Equal(float64(2)))
				Expect(metric.Safety).To(Equal(float64(3)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when happinessMetricService returns NotFound error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("metric not found").WithPayload("metric", metricID).Create()
				happinessMetricServiceMock.EXPECT().Get(metricID).Times(1).Return(nil, expErr)
			})
			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get happiness metric (metric not found; details [metric: 33])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get happiness metric, cause: ebt.not_exist: metric not found")
			})
		})

		Context("when happinessMetricService returns error", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("expected error")
				expErr = errory.OnGetErrors.Wrap(someErr)
				happinessMetricServiceMock.EXPECT().Get(metricID).Times(1).Return(nil, someErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get happiness metric (expected error)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get happiness metric, cause: ebt.provider_error: expected error")
			})
		})
	})

	Describe("Delete()", func() {
		const (
			metricID int64 = 33
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("DELETE", fmt.Sprintf("/v1/happiness_metric/%d", metricID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				happinessMetricServiceMock.EXPECT().Delete(&userContext, metricID).Times(1).Return(nil)
				expErr = nil
			})

			It("returns 200 code", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when happinessMetricService returns NotFound error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("metric not found").WithPayload("metric", metricID).Create()
				happinessMetricServiceMock.EXPECT().Delete(&userContext, metricID).Times(1).Return(expErr)
			})
			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete happiness metric (metric not found; details [metric: 33])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_delete_error: Cannot delete happiness metric, cause: ebt.not_exist: metric not found")
			})
		})

		Context("when happinessMetricService returns DeleteLastMetricForbiddenErrors error", func() {

			BeforeEach(func() {
				someErr := errory.DeleteLastMetricForbiddenErrors.New("tried to delete last happiness metric")
				expErr = errory.OnDeleteErrors.Wrap(someErr)
				happinessMetricServiceMock.EXPECT().Delete(&userContext, metricID).Times(1).Return(someErr)
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete happiness metric (tried to delete last happiness metric)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete happiness metric, cause: ebt.api_error.on_delete_error.delete_last_metric_forbidden: tried to delete last happiness metric`)
			})
		})

		Context("when happinessMetricService returns other error", func() {

			BeforeEach(func() {
				someErr := fmt.Errorf("some error")
				expErr = errory.OnDeleteErrors.Wrap(someErr)
				happinessMetricServiceMock.EXPECT().Delete(&userContext, metricID).Times(1).Return(someErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete happiness metric", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_delete_error: Cannot delete happiness metric, cause: some error")
			})
		})
	})
})
