// +build unitTests

package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	grafanaModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("OrgAPI", func() {
	var mockController *gomock.Controller
	var orgAPI *OrgAPI
	var sloServiceMock *service.MockISloService
	var dataSourceServiceMock *service.MockIDatasourceService
	var orgServiceMock *service.MockIOrganizationService
	var happinessMetricServiceMock *service.MockIHappinessMetricService
	var validatorMock *validator.MockITranslatedValidator
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var expErr error
	var userContext auth.UserContext

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		sloServiceMock = service.NewMockISloService(mockController)
		dataSourceServiceMock = service.NewMockIDatasourceService(mockController)
		orgServiceMock = service.NewMockIOrganizationService(mockController)
		happinessMetricServiceMock = service.NewMockIHappinessMetricService(mockController)
		validatorMock = validator.NewMockITranslatedValidator(mockController)
		orgAPI = &OrgAPI{
			SloService:        sloServiceMock,
			OrgService:        orgServiceMock,
			DatasourceService: dataSourceServiceMock,
			HappinessService:  happinessMetricServiceMock,
			Validator:         validatorMock,
			Log:               logger,
		}

		userContext = auth.UserContext{
			ID:     34,
			Cookie: "",
		}

		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContextMiddleware := func(c *gin.Context) {
			c.Set("UserContext", &userContext)
			c.Next()
		}

		ginEngine.GET("/v1/org/:id", orgAPI.GetOrg)
		ginEngine.GET("/v1/org/:id/slo", orgAPI.GetSlos)
		ginEngine.POST("/v1/org/:id/slo", orgAPI.FindSlos)
		ginEngine.GET("/v1/org/:id/datasource", userContextMiddleware, orgAPI.GetDatasources)
		ginEngine.GET("/v1/org/:id/user_happiness", userContextMiddleware, orgAPI.GetAllHappinessMetricsForUser)
		ginEngine.GET("/v1/org/:id/team_happiness", userContextMiddleware, orgAPI.GetAllHappinessMetricsForTeam)
		ginEngine.POST("/v1/org/:id/team_happiness/average", userContextMiddleware, orgAPI.SaveTeamAverage)
		ginEngine.GET("/v1/org/:id/team_happiness/missing", userContextMiddleware, orgAPI.GetUsersMissingInput)

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("GetSlos()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%s/slo", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("Org id").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs for organization (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get SLOs for organization, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const correctOrgID int64 = 99

			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d/slo", correctOrgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			Context("when slo service returns an error", func() {
				BeforeEach(func() {
					someErr := fmt.Errorf(`{"message":"some test error"}`)
					sloServiceMock.EXPECT().GetByOrgID(correctOrgID).Times(1).Return(nil, someErr)
					expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs for organization", expErr)
					assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get SLOs for organization, cause: {"message":"some test error"}`)
				})
			})
			Context("when no slo was found", func() {
				BeforeEach(func() {
					sloServiceMock.EXPECT().GetByOrgID(correctOrgID).Times(1).Return([]*model.Slo{}, nil)
					expErr = nil
				})
				It("returns 200 code with empty JSON array", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(w.Body.String()).To(Equal(`[]`))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})
			Context("when slos were found", func() {
				foundSlos := []*model.Slo{
					{
						ID:                              1,
						OrgID:                           2,
						Name:                            "Test",
						SuccessRateExpectedAvailability: "99.9",
						ComplianceExpectedAvailability:  "99.9",
						CreationDate:                    time.Now(),
					},
					{
						ID:                              2,
						OrgID:                           2,
						Name:                            "Test2",
						SuccessRateExpectedAvailability: "99.9",
						ComplianceExpectedAvailability:  "99.9",
						CreationDate:                    time.Now(),
					},
				}
				foundSlosJSON, _ := json.Marshal(foundSlos)

				BeforeEach(func() {
					sloServiceMock.EXPECT().GetByOrgID(correctOrgID).Times(1).Return(foundSlos, nil)
					expErr = nil
				})
				It("returns 200 code with found slos", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(w.Body.String()).To(BeEquivalentTo(foundSlosJSON))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})
		})
	})

	Describe("GetDatasources()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%s/datasource", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("Org id").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get datasources for organization (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get datasources for organization, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const correctOrgID int64 = 99

			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d/datasource", correctOrgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			Context("when datasource service returns an error", func() {
				BeforeEach(func() {
					someErr := fmt.Errorf(`some test error`)
					dataSourceServiceMock.EXPECT().GetDatasourcesByOrganizationID(correctOrgID).Times(1).Return(nil, someErr)
					expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot get datasources for organization", expErr)
					assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get datasources for organization, cause: some test error")
				})
			})
			Context("when no datasource was found", func() {
				BeforeEach(func() {
					dataSourceServiceMock.EXPECT().GetDatasourcesByOrganizationID(correctOrgID).Times(1).Return([]*grafanaModel.Datasource{}, nil)
					expErr = nil
				})
				It("returns 200 code with empty JSON array", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(w.Body.String()).To(Equal(`[]`))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})
			Context("when datasources were found", func() {
				foundDatasources := []*grafanaModel.Datasource{
					{
						ID:   22,
						Name: "Elastic MCC4",
						Type: "elasticsearch",
					},
					{
						ID:   23,
						Name: "Prometheus 5",
						Type: "prometheus",
					},
				}
				foundDatasourcesJSON, _ := json.Marshal(foundDatasources)

				BeforeEach(func() {
					dataSourceServiceMock.EXPECT().GetDatasourcesByOrganizationID(correctOrgID).Times(1).Return(foundDatasources, nil)
					expErr = nil
				})
				It("returns 200 code with found datasources", func() {
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(w.Body.String()).To(BeEquivalentTo(foundDatasourcesJSON))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})
		})
	})

	Describe("FindSlos()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/org/%s/slo", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("ID parameter cannot be parsed").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get SLOs, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const correctOrgID int64 = 99
			var (
				params      model.SloQueryParams
				requestBody string
				err         error
			)
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/org/%d/slo", correctOrgID), bytes.NewBufferString(requestBody))
				ginEngine.ServeHTTP(w, req)
			})

			Context("when slo service returns slos", func() {
				var expectedSlos = []*model.Slo{
					{
						ID:   int64(1),
						Name: "name 1",
					}, {
						ID:   int64(2),
						Name: "name 2",
					},
				}
				BeforeEach(func() {
					requestBody = fmt.Sprintf(`{
						"datasourceType": "elasticsearch",
						"metricType": "%v"
					}`, model.MetricTypeSuccessRate)

					err = json.Unmarshal([]byte(requestBody), &params)

					validatorMock.EXPECT().Validate(context.Background(), gomock.Any()).Times(1)
					sloServiceMock.EXPECT().FindSlos(&model.SloQueryParams{
						DatasourceType: "elasticsearch",
						MetricType:     model.MetricTypeSuccessRate,
						OrgID:          correctOrgID,
					}).Times(1).Return(expectedSlos, nil)
					expErr = nil
				})
				It("returns 200 code with slos", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					bytes, err := json.Marshal(&expectedSlos)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Body.String()).To(Equal(string(bytes)))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})

			Context("when wrong format of body is passed", func() {
				BeforeEach(func() {
					requestBody = `{`

					err = json.Unmarshal([]byte(requestBody), &params)
					errMsg := "unexpected EOF"
					expErr = errory.ValidationErrors.New(errMsg)
				})
				It("returns 400 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs (error validating data)", expErr)
					assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SLOs, cause: ebt.validation_error: error validating data, cause: unexpected EOF")
				})
			})

			Context("when validator returns 'BadRequest' error", func() {
				BeforeEach(func() {
					requestBody = fmt.Sprintf(`{
						"datasourceType": "unknown metric type",
						"metricType": "%v"
					}`, model.MetricTypeSuccessRate)
					var slo model.Slo
					err = json.Unmarshal([]byte(requestBody), &slo)
					expErr = errory.ValidationErrors.Builder().WithMessage("validation failed").Create()
					validatorMock.EXPECT().Validate(context.Background(), gomock.Any()).Times(1).Return(expErr)
				})

				It("returns 400 code with proper message", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs (validation failed)", expErr)
					assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SLOs, cause: ebt.validation_error: validation failed")
				})
			})

			Context("when slo service returns an error", func() {
				BeforeEach(func() {
					requestBody = fmt.Sprintf(`{
						"datasourceType": "elasticsearch",
						"metricType": "%v"
					}`, model.MetricTypeSuccessRate)

					err = json.Unmarshal([]byte(requestBody), &params)

					validatorMock.EXPECT().Validate(context.Background(), gomock.Any()).Times(1)
					someErr := fmt.Errorf(`some error`)
					sloServiceMock.EXPECT().FindSlos(gomock.Any()).Times(1).Return([]*model.Slo{}, someErr)
					expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot get SLOs", expErr)
					assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get SLOs, cause: some error")
				})
			})
		})
	})

	Describe("GetAllHappinessMetricsForUser()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%s/user_happiness", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("ID parameter cannot be parsed").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get happiness metrics (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get happiness metrics, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const (
				orgID  int64 = 2
				userID int64 = 34
			)
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d/user_happiness", orgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
			Context("when the request succeeds", func() {
				foundMetric := []*model.HappinessMetric{{
					ID:        1,
					UserID:    34,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}}
				BeforeEach(func() {
					expErr = nil
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForUser(orgID, userID).Times(1).Return(foundMetric, nil)
				})

				It("returns 200 code with proper message", func() {
					var metric []*model.HappinessMetric
					err := json.Unmarshal(w.Body.Bytes(), &metric)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(metric[0].ID).To(Equal(int64(1)))
					Expect(metric[0].UserID).To(Equal(userID))
					Expect(metric[0].Happiness).To(Equal(float64(2)))
					Expect(metric[0].Safety).To(Equal(float64(3)))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})

			Context("when happinessMetricService returns NotFound error", func() {
				BeforeEach(func() {
					expErr = errory.NotFoundErrors.Builder().WithMessage("nothing found").WithPayload("userID", userID).Create()
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForUser(orgID, userID).Times(1).Return(nil, expErr)
				})
				It("returns 404 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusNotFound))
					assertions.AssertAPIResponse(w.Body.String(), "nothing found; details [userID: 34]", expErr)
					assertions.AssertLogger(logHook, "ebt.not_exist: nothing found")
				})
			})

			Context("when happinessMetricService returns error", func() {
				BeforeEach(func() {
					someErr := errory.ProviderErrors.New("expected error")
					expErr = errory.OnGetErrors.Wrap(someErr)
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForUser(orgID, userID).Times(1).Return(nil, someErr)
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "expected error", expErr)
					assertions.AssertLogger(logHook, "ebt.provider_error: expected error")
				})
			})
		})
	})

	Describe("GetAllHappinessMetricsForTeam()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%s/team_happiness", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("ID parameter cannot be parsed").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get happiness metrics (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get happiness metrics, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const (
				orgID int64 = 2
			)
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d/team_happiness", orgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
			Context("when the request succeeds", func() {
				foundMetric := []*model.HappinessMetric{{
					ID:        1,
					OrgID:     2,
					Happiness: 2,
					Safety:    3,
					Date:      t,
				}}
				BeforeEach(func() {
					expErr = nil
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForTeam(orgID).Times(1).Return(foundMetric, nil)
				})

				It("returns 200 code with proper message", func() {
					var metric []*model.HappinessMetric
					err := json.Unmarshal(w.Body.Bytes(), &metric)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(metric[0].ID).To(Equal(int64(1)))
					Expect(metric[0].UserID).To(Equal(int64(0)))
					Expect(metric[0].Happiness).To(Equal(float64(2)))
					Expect(metric[0].Safety).To(Equal(float64(3)))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})

			Context("when happinessMetricService returns NotFound error", func() {
				BeforeEach(func() {
					expErr = errory.NotFoundErrors.Builder().WithMessage("nothing found").WithPayload("orgID", orgID).Create()
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForTeam(orgID).Times(1).Return(nil, expErr)
				})
				It("returns 404 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusNotFound))
					assertions.AssertAPIResponse(w.Body.String(), "nothing found; details [orgID: 2]", expErr)
					assertions.AssertLogger(logHook, "ebt.not_exist: nothing found")
				})
			})

			Context("when happinessMetricService returns error", func() {
				BeforeEach(func() {
					someErr := errory.ProviderErrors.New("expected error")
					expErr = errory.OnGetErrors.Wrap(someErr)
					happinessMetricServiceMock.EXPECT().GetAllHappinessMetricsForTeam(orgID).Times(1).Return(nil, someErr)
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "expected error", expErr)
					assertions.AssertLogger(logHook, "ebt.provider_error: expected error")
				})
			})
		})
	})

	Describe("SaveTeamAverage()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/org/%s/team_happiness/average", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("ID parameter cannot be parsed").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot create happiness metrics average (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create happiness metrics average, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const (
				orgID int64 = 2
			)
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("POST", fmt.Sprintf("/v1/org/%d/team_happiness/average", orgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			Context("when the request succeeds", func() {
				BeforeEach(func() {
					expErr = nil
					happinessMetricServiceMock.EXPECT().SaveTeamAverage(orgID).Times(1).Return(int64(5), nil)
				})

				It("returns 201 code and id of created metric", func() {
					Expect(w.Code).To(Equal(http.StatusCreated))
					Expect(w.Body.String()).To(Equal(fmt.Sprintf(`{"id":%d}`, 5)))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})

			Context("when happinessMetricService returns error", func() {
				BeforeEach(func() {
					someErr := errory.ProviderErrors.New("expected error")
					expErr = errory.OnGetErrors.Wrap(someErr)
					happinessMetricServiceMock.EXPECT().SaveTeamAverage(orgID).Times(1).Return(int64(5), someErr)
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "expected error", expErr)
					assertions.AssertLogger(logHook, "ebt.provider_error: expected error")
				})
			})
		})
	})

	Describe("GetUsersMissingInput()", func() {
		Context("when orgID param is in wrong format", func() {
			const incorrectOrgID = "testId99"
			BeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%s/team_happiness/missing", incorrectOrgID), nil)
				ginEngine.ServeHTTP(w, req)
				expErr = errory.ParseErrors.Builder().WithMessage("ID parameter cannot be parsed").Create()
			})
			It("returns 400 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get missing users inputs (ID parameter cannot be parsed)", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get missing users inputs, cause: ebt.parsing_error: ID parameter cannot be parsed, cause: strconv.ParseInt: parsing "testId99": invalid syntax`)
			})
		})

		Context("when orgID param is in correct format", func() {
			const (
				orgID int64 = 2
			)
			JustBeforeEach(func() {
				w = httptest.NewRecorder()
				req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d/team_happiness/missing", orgID), nil)
				ginEngine.ServeHTTP(w, req)
			})

			Context("when the request succeeds", func() {
				foundUser := []*model.UserMissingInput{{
					UserID: 1,
					Login:  "x-man",
				}}
				BeforeEach(func() {
					expErr = nil
					happinessMetricServiceMock.EXPECT().GetUsersMissingInput(orgID).Times(1).Return(foundUser, nil)
				})

				It("returns 200 code with proper message", func() {
					var metric []*model.UserMissingInput
					err := json.Unmarshal(w.Body.Bytes(), &metric)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(metric[0].UserID).To(Equal(int64(1)))
					Expect(metric[0].Login).To(Equal("x-man"))
					assertions.AssertAPIResponse(w.Body.String(), "", expErr)
					assertions.AssertLogger(logHook, "")
				})
			})

			Context("when happinessMetricService returns NotFound error", func() {
				BeforeEach(func() {
					expErr = errory.NotFoundErrors.Builder().WithMessage("nothing found").WithPayload("orgID", orgID).Create()
					happinessMetricServiceMock.EXPECT().GetUsersMissingInput(orgID).Times(1).Return(nil, expErr)
				})
				It("returns 404 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusNotFound))
					assertions.AssertAPIResponse(w.Body.String(), "nothing found; details [orgID: 2]", expErr)
					assertions.AssertLogger(logHook, "ebt.not_exist: nothing found")
				})
			})

			Context("when happinessMetricService returns error", func() {
				BeforeEach(func() {
					someErr := errory.ProviderErrors.New("expected error")
					expErr = errory.OnGetErrors.Wrap(someErr)
					happinessMetricServiceMock.EXPECT().GetUsersMissingInput(orgID).Times(1).Return(nil, someErr)
				})
				It("returns 500 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))
					assertions.AssertAPIResponse(w.Body.String(), "expected error", expErr)
					assertions.AssertLogger(logHook, "ebt.provider_error: expected error")
				})
			})
		})
	})

	Describe("GetOrg()", func() {
		const correctOrgID int64 = 99

		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", fmt.Sprintf("/v1/org/%d", correctOrgID), nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when the request succeeds", func() {
			orgDetails := &grafanaModel.Organization{
				ID:           correctOrgID,
				Name:         "test-org-name",
				ProductID:    sql.NullInt64{Int64: 9, Valid: true},
				SolutionID:   sql.NullInt64{Int64: 11, Valid: true},
				ProductName:  sql.NullString{String: "product-0", Valid: true},
				SolutionName: sql.NullString{String: "solution-11", Valid: true},
				Featured:     sql.NullBool{Bool: true, Valid: true},
			}
			orgDetailsJSON, _ := json.Marshal(orgDetails)

			BeforeEach(func() {
				orgServiceMock.EXPECT().GetOrganizationByID(correctOrgID).Times(1).Return(orgDetails, nil)
				expErr = nil
			})
			It("returns 200 code with found slos", func() {
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(BeEquivalentTo(orgDetailsJSON))
				assertions.AssertAPIResponse(w.Body.String(), "", expErr)
				assertions.AssertLogger(logHook, "")
			})
		})

		Context("when org service returns an error", func() {
			BeforeEach(func() {
				someErr := fmt.Errorf(`{"message":"some test error"}`)
				orgServiceMock.EXPECT().GetOrganizationByID(correctOrgID).Times(1).Return(nil, someErr)
				expErr = errory.OnGetErrors.Builder().Wrap(someErr).Create()
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get organization details", expErr)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get organization details, cause: {"message":"some test error"}`)
			})
		})

		Context("when org was not found", func() {
			BeforeEach(func() {
				expErr = errory.NotFoundErrors.Builder().WithMessage("org not found").WithPayload("ID", correctOrgID).Create()
				orgServiceMock.EXPECT().GetOrganizationByID(correctOrgID).Times(1).Return(nil, expErr)
			})
			It("returns 200 code with empty JSON array", func() {
				Expect(w.Code).To(Equal(http.StatusNotFound))
				assertions.AssertAPIResponse(w.Body.String(), "Cannot get organization details (org not found; details [ID: 99])", expErr)
				assertions.AssertLogger(logHook, "ebt.api_error.on_get_error: Cannot get organization details, cause: ebt.not_exist: org not found")
			})
		})

	})
})
