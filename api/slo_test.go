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

var _ = Describe("SloAPI", func() {
	var mockController *gomock.Controller
	var sloAPI *SloAPI
	var sloServiceMock *service.MockISloService
	var validatorMock *validator.MockISLOValidator
	var createScope context.Context
	var updateScope context.Context

	var userContext auth.UserContext
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		sloServiceMock = service.NewMockISloService(mockController)
		validatorMock = validator.NewMockISLOValidator(mockController)
		sloAPI = &SloAPI{
			SloService: sloServiceMock,
			Validator:  validatorMock,
			Log:        logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContext = auth.UserContext{
			ID:     3,
			Cookie: "test cookie",
		}
		userContextMiddleware := func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		sloRoutes := ginEngine.Group("/v1/slo")
		{
			sloRoutes.POST("", userContextMiddleware, sloAPI.Create)
			sloRoutes.GET("", userContextMiddleware, sloAPI.GetDetailed)
			sloRoutes.PUT("/:id", userContextMiddleware, sloAPI.Update)
			sloRoutes.GET("/:id", sloAPI.Get)
			sloRoutes.DELETE("/:id", userContextMiddleware, sloAPI.Delete)
			sloRoutes.DELETE("/:id/history", userContextMiddleware, sloAPI.DeleteSloHistory)
		}
		createScope = context.WithValue(context.Background(), ctx.Create, true)
		updateScope = context.WithValue(context.Background(), ctx.Create, false)
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
			"orgId": 12,
			"name": "test",
			"successRateExpAvailability": "99.9",
			"complianceExpAvailability": "99.99"
		}`
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest(requestMethod, fmt.Sprintf("/v1/slo/%s", invalidID), bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})
		Context("when http methods is PUT", func() {
			BeforeEach(func() {
				requestMethod = "PUT"
				expErr = errory.ParseErrors.Builder().WithMessage("SLO id").Create()
			})
			It("returns 400 code with proper", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update SLO (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_update_error: Cannot update SLO, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
		Context("when http methods is GET", func() {
			BeforeEach(func() {
				requestMethod = "GET"
				expErr = errory.ParseErrors.Builder().WithMessage("SLO id").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLO (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get SLO, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
		Context("when http methods is DELETE", func() {
			BeforeEach(func() {
				requestMethod = "DELETE"
				expErr = errory.ParseErrors.Builder().WithMessage("SLO id").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete SLO (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete SLO, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "abc": invalid syntax`)
			})
		})
	})

	Describe("Create()", func() {
		var (
			slo         model.Slo
			err         error
			requestBody string
		)

		AfterEach(func() {
			slo = model.Slo{}
		})
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("POST", "/v1/slo", bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			const (
				createdSloID int64 = 9
				orgID        int64 = 10
			)

			Context("user generated SLO", func() {
				BeforeEach(func() {
					requestBody = fmt.Sprintf(`{
						"orgId": %d,
						"name": "test",
						"successRateExpAvailability": "99.9",
						"complianceExpAvailability": "99.99"
					}`, orgID)

					err = json.Unmarshal([]byte(requestBody), &slo)

					validatorMock.EXPECT().Validate(createScope, slo).Times(1)
					sloServiceMock.EXPECT().Create(&userContext, &slo).Times(1).Do(func(userContext *auth.UserContext, slo *model.Slo) {
						slo.ID = createdSloID
						Expect(userContext.ID).To(Equal(int64(3)))
						Expect(userContext.Cookie).To(Equal("test cookie"))
					})
					expErr = nil
				})

				It("returns 201 code and id of created SLO", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusCreated))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})
		})

		Context("when slo name (required field) is not set", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"successRateExpAvailability": "99.9",
					"complianceExpAvailability": "99.99"
				}`
				expErr = errory.ValidationErrors.Builder().WithMessage("validation failed").Create()
				validatorMock.EXPECT().Validate(createScope, gomock.Any()).Times(1).DoAndReturn(func(ct context.Context, slo model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("99.9"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					return expErr
				})
			})

			It("returns 400 code and proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create SLO (validation failed)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_create_error: Cannot create SLO, cause: ebt.validation_error: validation failed")
			})
		})

		Context("when validator returns 'BadRequest' error", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"name": "name with forbidden character #"
				}`
				var slo model.Slo
				expErr = errory.ValidationErrors.Builder().WithMessage("validation failed").Create()
				err = json.Unmarshal([]byte(requestBody), &slo)
				validatorMock.EXPECT().Validate(createScope, slo).Times(1).Return(expErr)
			})

			It("returns 400 code with proper message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create SLO (validation failed)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_create_error: Cannot create SLO, cause: ebt.validation_error: validation failed")
			})
		})

		Context("when sloService returns 'NotUnique' error", func() {
			const orgID int64 = 22
			BeforeEach(func() {
				requestBody = fmt.Sprintf(`{
					"orgId": %d,
					"name": "test",
					"successRateExpAvailability": "99.9",
					"complianceExpAvailability": "99.99"
				}`, orgID)
				err = json.Unmarshal([]byte(requestBody), &slo)
				expErr = errory.NotUniqueErrors.Builder().WithMessage("slo with the same name already exists within given org").WithPayload("Slo name", "name").Create()
				validatorMock.EXPECT().Validate(createScope, slo)
				sloServiceMock.EXPECT().Create(&userContext, &slo).Times(1).Return(expErr)
			})

			It("returns 409 code and error message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusConflict))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create SLO (slo with the same name already exists within given org; details [Slo name: name])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_create_error: Cannot create SLO, cause: ebt.not_unique: slo with the same name already exists within given org")
			})
		})
		Context("when sloService returns not 'special' error", func() {
			const orgID int64 = 22
			BeforeEach(func() {
				requestBody = fmt.Sprintf(`{
					"orgId": %d,
					"name": "test",
					"successRateExpAvailability": "99.9",
					"complianceExpAvailability": "99.99"
				}`, orgID)
				err = json.Unmarshal([]byte(requestBody), &slo)
				someErr := errory.ProviderErrors.New("test error")
				expErr = errory.OnCreateErrors.Builder().Wrap(someErr).Create()
				validatorMock.EXPECT().Validate(createScope, slo)
				sloServiceMock.EXPECT().Create(&userContext, &slo).Times(1).Return(someErr)
			})

			It("returns 500 code and error api error message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create SLO (test error)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create SLO, cause: ebt.provider_error: test error`)
			})
		})
	})

	Describe("Update()", func() {
		var (
			requestBody string
			err         error
		)
		const (
			sloID        int64 = 33
			createdSloID int64 = 9
			create       bool  = false
		)
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("PUT", fmt.Sprintf("/v1/slo/%d", sloID), bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"name": "test",
					"successRateExpAvailability": "99.9",
					"complianceExpAvailability": "99.99"
				}`

				validatorMock.EXPECT().Validate(updateScope, gomock.Any()).DoAndReturn(func(ct context.Context, slo model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("99.9"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					return nil
				})

				sloServiceMock.EXPECT().Update(&userContext, gomock.Any()).Times(1).DoAndReturn(func(userContext *auth.UserContext, slo *model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("99.9"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					Expect(userContext.ID).To(Equal(int64(3)))
					Expect(userContext.Cookie).To(Equal("test cookie"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					return nil
				})

				expErr = nil
			})

			It("returns 200 code and id of updated SLO", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, sloID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when the request has fields to be normalized", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"name": "test",
					"successRateExpAvailability": "1e1",
					"complianceExpAvailability": "0094.9900"
				}`

				validatorMock.EXPECT().Validate(updateScope, gomock.Any()).DoAndReturn(func(ct context.Context, slo model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("1e1"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("0094.9900"))
					return nil
				})

				sloServiceMock.EXPECT().Update(&userContext, gomock.Any()).Times(1).DoAndReturn(func(userContext *auth.UserContext, slo *model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("10"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("94.99"))
					Expect(userContext.ID).To(Equal(int64(3)))
					Expect(userContext.Cookie).To(Equal("test cookie"))
					return nil
				})
				expErr = nil
			})

			It("returns 200 code and id of updated SLO", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, sloID)))
				Expect(w.Code).To(Equal(http.StatusOK))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when validator returns 'BadRequest' error", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"name": "name with forbidden character #"
				}`
				slo := model.Slo{
					ID:    sloID,
					OrgID: int64(1),
					Name:  "name with forbidden character #",
				}
				expErr = errory.ValidationErrors.Builder().WithMessage("validation failed").Create()
				validatorMock.EXPECT().Validate(updateScope, slo).Times(1).Return(expErr)
			})

			It("returns 400 code with proper message", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update SLO (validation failed)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_update_error: Cannot update SLO, cause: ebt.validation_error: validation failed")
			})
		})

		Context("when sloService returns untyped error", func() {
			BeforeEach(func() {
				requestBody = `{
					"orgId": 1,
					"name": "test",
					"successRateExpAvailability": "99.9",
					"complianceExpAvailability": "99.99"
				}`
				someErr := errory.ProviderErrors.New("Error while updating")
				expErr = errory.OnUpdateErrors.Builder().Wrap(someErr).Create()
				validatorMock.EXPECT().Validate(updateScope, gomock.Any()).DoAndReturn(func(ct context.Context, slo model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("99.9"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					return nil
				})
				sloServiceMock.EXPECT().Update(&userContext, gomock.Any()).Times(1).DoAndReturn(func(userContext *auth.UserContext, slo *model.Slo) error {
					Expect(slo.OrgID).To(Equal(int64(1)))
					Expect(slo.Name).To(Equal("test"))
					Expect(slo.SuccessRateExpectedAvailability).To(Equal("99.9"))
					Expect(slo.ComplianceExpectedAvailability).To(Equal("99.99"))
					Expect(userContext.ID).To(Equal(int64(3)))
					Expect(userContext.Cookie).To(Equal("test cookie"))
					return someErr
				})
			})

			It("returns 500 code and error message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot update SLO (Error while updating)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_update_error: Cannot update SLO, cause: ebt.provider_error: Error while updating")
			})
		})

	})

	Describe("Delete()", func() {
		const sloID int64 = 33
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("DELETE", fmt.Sprintf("/v1/slo/%d", sloID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				sloServiceMock.EXPECT().Delete(&userContext, sloID).Times(1)
				expErr = nil
			})

			It("returns 200 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, sloID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when slo service returns 'NotFound' error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("slo not found").WithPayload("Slo", 1).Create()
				sloServiceMock.EXPECT().Delete(&userContext, sloID).Times(1).Return(expErr)
			})

			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete SLO (slo not found; details [Slo: 1])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_delete_error: Cannot delete SLO, cause: ebt.not_exist: slo not found")
			})
		})

		Context("when slo service returns not 'special' error", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("test slo")
				expErr = errory.OnDeleteErrors.Builder().Wrap(someErr).Create()
				sloServiceMock.EXPECT().Delete(&userContext, sloID).Times(1).Return(someErr)
			})

			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete SLO (test slo)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_delete_error: Cannot delete SLO, cause: ebt.provider_error: test slo")
			})
		})
	})

	Describe("DeleteSloHistory()", func() {
		const sloID int64 = 33
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("DELETE", fmt.Sprintf("/v1/slo/%d/history", sloID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				sloServiceMock.EXPECT().DeleteSloHistory(sloID).Times(1)
				expErr = nil
			})

			It("returns 200 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, sloID)))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when slo service returns not 'special' error", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("test slo")
				expErr = errory.OnDeleteErrors.Builder().Wrap(someErr).Create()
				sloServiceMock.EXPECT().DeleteSloHistory(sloID).Times(1).Return(someErr)
			})

			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot delete SLO History (test slo)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_delete_error: Cannot delete SLO History, cause: ebt.provider_error: test slo")
			})
		})
	})

	Describe("Get()", func() {
		const sloID int64 = 33
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/slo/%d", sloID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			foundSLO := model.Slo{
				ID:                              sloID,
				Name:                            "test2",
				ComplianceExpectedAvailability:  "88.8",
				SuccessRateExpectedAvailability: "77.77",
			}
			BeforeEach(func() {
				sloServiceMock.EXPECT().Get(sloID).Times(1).Return(&foundSLO, nil)
				expErr = nil
			})

			It("returns 200 code with proper message", func() {
				var slo model.Slo
				err := json.Unmarshal(w.Body.Bytes(), &slo)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(slo.ID).To(Equal(sloID))
				Expect(slo.Name).To(Equal("test2"))
				Expect(slo.ComplianceExpectedAvailability).To(Equal("88.8"))
				Expect(slo.SuccessRateExpectedAvailability).To(Equal("77.77"))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when slo service returns NotFound error", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("slo not found").WithPayload("SLO", 1).Create()
				sloServiceMock.EXPECT().Get(sloID).Times(1).Return(nil, expErr)
			})
			It("returns 404 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLO (slo not found; details [SLO: 1])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SLO, cause: ebt.not_exist: slo not found")
			})
		})
		Context("when slo service returns an error", func() {
			BeforeEach(func() {
				someErr := errory.ProviderErrors.New("test error")
				expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
				sloServiceMock.EXPECT().Get(sloID).Times(1).Return(nil, someErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLO (test error)", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SLO, cause: ebt.provider_error: test error")
			})
		})
	})

	Describe("GetDetailed()", func() {
		var path string

		BeforeEach(func() {
			path = "/v1/slo"
		})

		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", path, nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			BeforeEach(func() {
				foundSLOs := []*model.DetailedSlo{
					{
						Slo: model.Slo{
							ID:                              int64(12),
							Name:                            "test2",
							ComplianceExpectedAvailability:  "88.8",
							SuccessRateExpectedAvailability: "77.77",
						},
						OrgName: "test-org",
					},
					{
						Slo: model.Slo{
							ID:                              int64(45),
							Name:                            "test4",
							ComplianceExpectedAvailability:  "11",
							SuccessRateExpectedAvailability: "12",
						},
						OrgName: "test-org2",
					},
				}
				sloServiceMock.EXPECT().GetDetailedSlos("", "").Times(1).Return(foundSLOs, nil)
			})

			It("returns 200 code with proper message", func() {
				var slos []*model.DetailedSlo
				err := json.Unmarshal(w.Body.Bytes(), &slos)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(len(slos)).To(Equal(2))
				Expect(slos[0].ID).To(Equal(int64(12)))
				Expect(slos[0].Name).To(Equal("test2"))
				Expect(slos[0].ComplianceExpectedAvailability).To(Equal("88.8"))
				Expect(slos[0].SuccessRateExpectedAvailability).To(Equal("77.77"))
				Expect(slos[0].OrgName).To(Equal("test-org"))

				Expect(slos[1].ID).To(Equal(int64(45)))
				Expect(slos[1].Name).To(Equal("test4"))
				Expect(slos[1].ComplianceExpectedAvailability).To(Equal("11"))
				Expect(slos[1].SuccessRateExpectedAvailability).To(Equal("12"))
				Expect(slos[1].OrgName).To(Equal("test-org2"))
			})
		})

		Context("when the request is for filtered result", func() {
			BeforeEach(func() {
				path = "/v1/slo?name=test2&orgName=test-org"
				foundSLOs := []*model.DetailedSlo{
					{
						Slo: model.Slo{
							ID:                              int64(12),
							Name:                            "test2",
							ComplianceExpectedAvailability:  "88.8",
							SuccessRateExpectedAvailability: "77.77",
						},
						OrgName: "test-org",
					},
				}
				sloServiceMock.EXPECT().GetDetailedSlos("test2", "test-org").Times(1).Return(foundSLOs, nil)
			})

			It("returns 200 code with proper message", func() {
				var slos []*model.DetailedSlo
				err := json.Unmarshal(w.Body.Bytes(), &slos)
				Expect(err).ToNot(HaveOccurred())
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(len(slos)).To(Equal(1))
				Expect(slos[0].ID).To(Equal(int64(12)))
				Expect(slos[0].Name).To(Equal("test2"))
				Expect(slos[0].ComplianceExpectedAvailability).To(Equal("88.8"))
				Expect(slos[0].SuccessRateExpectedAvailability).To(Equal("77.77"))
				Expect(slos[0].OrgName).To(Equal("test-org"))
			})
		})

		Context("when slo service returns error", func() {
			var expErr error
			BeforeEach(func() {
				expErr = errory.ProviderErrors.Builder().WithMessage("database error").Create()
				sloServiceMock.EXPECT().GetDetailedSlos("", "").Times(1).Return(nil, expErr)
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})
