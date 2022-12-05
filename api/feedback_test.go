// +build unitTests

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

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

var _ = Describe("FeedbackAPI", func() {
	var mockController *gomock.Controller
	var feedbackAPI *FeedbackAPI
	var feedbackValidatorMock *validator.MockIFeedbackValidator
	var feedbackServiceMock *service.MockIFeedbackService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error
	var userContext auth.UserContext

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		feedbackServiceMock = service.NewMockIFeedbackService(mockController)
		feedbackValidatorMock = validator.NewMockIFeedbackValidator(mockController)
		feedbackAPI = &FeedbackAPI{
			Validator: feedbackValidatorMock,
			Service:   feedbackServiceMock,
			Log:       logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContextMiddleware := func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		feedbackRoutes := ginEngine.Group("/v1/feedback")
		{
			feedbackRoutes.POST("", userContextMiddleware, feedbackAPI.Create)
			feedbackRoutes.GET("/:id", feedbackAPI.Get)
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
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest(requestMethod, fmt.Sprintf("/v1/feedback/%s", invalidID), nil)
			ginEngine.ServeHTTP(w, req)
		})
		Context("when http method is GET", func() {
			BeforeEach(func() {
				requestMethod = "GET"
				expErr = errory.ParseErrors.Builder().WithPayload("fb id", 7).Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get feedback (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get feedback, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
	})

	Describe("Create()", func() {
		var (
			err         error
			fb          model.Feedback
			requestBody string
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/v1/feedback", bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			const (
				createdFeedbackID int64 = 9
			)
			BeforeEach(func() {
				requestBody = `{
					"feedbackDate": "2020-01-07T00:00:00.000Z",
					"givingUserId": 3,
					"receivingUserId": 4,
					"orgId": 2
				}`
				err = json.Unmarshal([]byte(requestBody), &fb)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				feedbackValidatorMock.EXPECT().Validate(ct, fb).Times(1)
				feedbackServiceMock.EXPECT().Create(&fb).Times(1).Do(func(fb *model.Feedback) {
					fb.ID = createdFeedbackID
				})
				expErr = nil
			})

			It("returns 200 code and id of created fb", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusCreated))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, createdFeedbackID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})

		})

		Context("when feedbackService returns an error", func() {
			BeforeEach(func() {
				requestBody = `{
					"feedbackDate": "2020-01-07T00:00:00.000Z",
					"givingUserId": 3,
					"receivingUserId": 4,
					"orgId": 2
				}`
				someErr := fmt.Errorf("some error")
				expErr = errory.OnCreateErrors.Builder().Wrap(someErr).Create()
				err = json.Unmarshal([]byte(requestBody), &fb)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				feedbackValidatorMock.EXPECT().Validate(ct, fb).Times(1)
				feedbackServiceMock.EXPECT().Create(&fb).Times(1).Return(someErr)
			})

			It("returns 500 code and error message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create feedback", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create feedback, cause: some error`)
			})
		})

		Context("when receiving_user_id is not set", func() {
			BeforeEach(func() {
				requestBody = `{
					"feedbackDate": "2020-01-07T00:00:00.000Z",
					"givingUserId": 3,
					"orgId": 2
				}`
				expErr = errory.ValidationErrors.Builder().WithMessage("receiving_user_id must not be null").Create()
				var tmpMetric model.Feedback
				err = json.Unmarshal([]byte(requestBody), &tmpMetric)
				ct := context.WithValue(context.Background(), ctx.Create, true)
				feedbackValidatorMock.EXPECT().Validate(ct, tmpMetric).Times(1).Return(expErr)
			})

			It("returns 400 code and proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create feedback (receiving_user_id must not be null)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_create_error: Cannot create feedback, cause: ebt.validation_error: receiving_user_id must not be null")
			})
		})
	})

	Describe("Get()", func() {
		const (
			orgID           int64 = 2
			feedbackID      int64 = 33
			receivingUserID int64 = 34
			givingUserID    int64 = 3
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/feedback/%d", orgID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			foundMetric := []*model.Feedback{
				{
					ID:              feedbackID,
					ReceivingUserID: receivingUserID,
					GivingUserID:    givingUserID,
					OrgID:           orgID,
				},
			}
			BeforeEach(func() {
				expErr = nil
				feedbackServiceMock.EXPECT().GetByOrgID(orgID).Times(1).Return(foundMetric, nil)
			})

			It("returns 200 code with proper message", func() {
				var fb []*model.Feedback
				err := json.Unmarshal(w.Body.Bytes(), &fb)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(fb[0].ID).To(Equal(feedbackID))
				Expect(fb[0].ReceivingUserID).To(Equal(receivingUserID))
				Expect(fb[0].GivingUserID).To(Equal(givingUserID))
				Expect(fb[0].OrgID).To(Equal(orgID))
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when feedbackService returns NotFound error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("feedback not found").WithPayload("fb", feedbackID).Create()
				feedbackServiceMock.EXPECT().GetByOrgID(orgID).Times(1).Return(nil, expErr)
			})
			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get feedback (feedback not found; details [fb: 33])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get feedback, cause: ebt.not_exist: feedback not found")
			})
		})

		Context("when feedbackService returns NotFound error", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("expected error")
				expErr = errory.OnGetErrors.Wrap(someErr)
				feedbackServiceMock.EXPECT().GetByOrgID(orgID).Times(1).Return(nil, someErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get feedback (expected error)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get feedback, cause: ebt.provider_error: expected error")
			})
		})
	})
})
