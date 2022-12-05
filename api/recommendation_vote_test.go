// +build unitTests

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
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

var _ = Describe("RecommendationVoteAPI", func() {
	var mockController *gomock.Controller
	var recommendationVoteAPI *RecommendationVoteAPI
	var recommendationVoteValidatorMock *validator.MockITranslatedValidator
	var recommendationVoteServiceMock *service.MockIRecommendationVoteService
	logger, logHook := logrustest.NewNullLogger()

	var ginEngine *gin.Engine
	var w *httptest.ResponseRecorder
	var req *http.Request
	var userContext *auth.UserContext

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		recommendationVoteServiceMock = service.NewMockIRecommendationVoteService(mockController)
		recommendationVoteValidatorMock = validator.NewMockITranslatedValidator(mockController)
		recommendationVoteAPI = &RecommendationVoteAPI{
			Validator: recommendationVoteValidatorMock,
			Service:   recommendationVoteServiceMock,
			Log:       logger,
		}
		gin.SetMode(gin.TestMode)
		ginEngine = gin.New()

		userContextMiddleware := func(c *gin.Context) {
			if userContext != nil {
				c.Set("UserContext", userContext)
			}
			c.Next()
		}

		recommendationVoteRoutes := ginEngine.Group("/v1/recommendation_vote")
		{
			recommendationVoteRoutes.POST("", userContextMiddleware, recommendationVoteAPI.Create)
			recommendationVoteRoutes.GET("", userContextMiddleware, recommendationVoteAPI.Get)
			recommendationVoteRoutes.DELETE("", userContextMiddleware, recommendationVoteAPI.Delete)
		}

		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
		mockController = gomock.NewController(GinkgoT())
	})

	Describe("Get()", func() {
		var orgID string
		JustBeforeEach(func() {
			w = httptest.NewRecorder()
			url := "/v1/recommendation_vote"
			if orgID != "" {
				url = url + "?orgId=" + orgID
			}
			req, _ = http.NewRequest("GET", url, nil)
			ginEngine.ServeHTTP(w, req)
		})

		Context("when there is no user context", func() {
			BeforeEach(func() {
				userContext = nil
			})
			It("returns 500 code with proper message", func() {
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
				expError := errory.OnGetErrors.Builder().Wrap(errory.ProcessingErrors.New("could not extract UserContext")).WithMessage("Cannot get votes").Create()
				assertions.AssertAPIResponse(w.Body.String(), "", expError)
				assertions.AssertLogger(logHook, `ebt.api_error.on_get_error: Cannot get votes, cause: ebt.processing_error: could not extract UserContext`)
			})
		})

		Context("when user is in context", func() {
			BeforeEach(func() {
				userContext = &auth.UserContext{
					ID:     2,
					Cookie: "terefere",
				}
			})

			Context("when no orgID is provided and service succeeds", func() {
				t1, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
				t2, _ := time.Parse(time.RFC3339, "2020-01-08T00:00:00Z")
				foundVotes := []*model.RecommendationVote{
					{
						ID:                 15,
						UserID:             2,
						OrgID:              2,
						RecommendationType: "jiic",
						Vote:               model.Like,
						Date:               t1,
					},
					{
						ID:                 16,
						UserID:             2,
						OrgID:              3,
						RecommendationType: "jiic",
						Vote:               model.Dislike,
						Date:               t2,
					},
				}
				BeforeEach(func() {
					orgID = ""
					recommendationVoteServiceMock.EXPECT().Get(int64(2), nil).Times(1).Return(foundVotes, nil)
				})

				It("returns all votes for user", func() {
					Expect(w.Code).To(Equal(http.StatusOK))

					var metrics []*model.RecommendationVote
					err := json.Unmarshal(w.Body.Bytes(), &metrics)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(metrics).To(HaveLen(2))
					Expect(metrics).To(Equal(foundVotes))
				})
			})
			Context("when no orgID is provided and service fails", func() {
				BeforeEach(func() {
					orgID = ""
					recommendationVoteServiceMock.EXPECT().Get(int64(2), nil).Times(1).Return(nil, errory.ProviderErrors.New("error when getting votes"))
				})

				It("returns proper error", func() {
					Expect(w.Code).To(Equal(http.StatusInternalServerError))

					expError := errory.OnGetErrors.Builder().Wrap(errory.ProviderErrors.New("error when getting votes")).WithMessage("Cannot get votes").Create()
					assertions.AssertAPIResponse(w.Body.String(), "", expError)
				})
			})

			Context("when invalid orgID is provided", func() {
				BeforeEach(func() {
					orgID = "errorbudget"
				})

				It("returns proper error", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					expError := errory.OnGetErrors.New("orgId parameter cannot be parsed")
					assertions.AssertAPIResponse(w.Body.String(), "", expError)
					assertions.AssertLogger(logHook, `ebt.parsing_error: orgId parameter cannot be parsed, cause: strconv.ParseInt: parsing "errorbudget": invalid syntax`)
				})
			})

			Context("when correct orgID is provided", func() {
				t, _ := time.Parse(time.RFC3339, "2020-01-07T00:00:00Z")
				foundVotes := []*model.RecommendationVote{
					{
						ID:                 15,
						UserID:             2,
						OrgID:              12,
						RecommendationType: "jiic",
						Vote:               model.Like,
						Date:               t,
					},
				}
				BeforeEach(func() {
					orgID = "12"
					orgIDParam := int64(12)
					recommendationVoteServiceMock.EXPECT().Get(int64(2), &orgIDParam).Times(1).Return(foundVotes, nil)
				})

				It("returns votes from org for user", func() {
					Expect(w.Code).To(Equal(http.StatusOK))

					var metrics []*model.RecommendationVote
					err := json.Unmarshal(w.Body.Bytes(), &metrics)
					Expect(err).ToNot(HaveOccurred())
					Expect(w.Code).To(Equal(http.StatusOK))
					Expect(metrics).To(HaveLen(1))
					Expect(metrics).To(Equal(foundVotes))
				})
			})
		})
	})

	Describe("Create() and Delete()", func() {
		var requestBody string
		var method string
		JustBeforeEach(func() {
			w = httptest.NewRecorder()

			req, _ = http.NewRequest(method, "/v1/recommendation_vote", bytes.NewBufferString(requestBody))
			ginEngine.ServeHTTP(w, req)
		})

		Context("when body cannot be parsed", func() {
			BeforeEach(func() {
				requestBody = "I like this recommendation"
			})
			Context("POST", func() {
				BeforeEach(func() {
					method = "POST"
				})
				It("returns 400 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot create vote (error validating data)", errory.OnCreateErrors.New(""))
					assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create vote, cause: ebt.validation_error: error validating data, cause: invalid character 'I' looking for beginning of value`)
				})
			})
			Context("DELETE", func() {
				BeforeEach(func() {
					method = "DELETE"
				})
				It("returns 400 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot delete vote (error validating data)", errory.OnCreateErrors.New(""))
					assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete vote, cause: ebt.validation_error: error validating data, cause: invalid character 'I' looking for beginning of value`)
				})
			})
		})

		Context("when body can be parsed but is not valid", func() {
			BeforeEach(func() {
				requestBody = "{}"
				recommendationVoteValidatorMock.EXPECT().Validate(gomock.Any(), model.RecommendationVote{}).Times(1).Return(errory.ValidationErrors.New("orgId is required field"))
			})
			Context("POST", func() {
				BeforeEach(func() {
					method = "POST"
				})
				It("returns 400 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot create vote (orgId is required field)", errory.OnCreateErrors.New(""))
					assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create vote, cause: ebt.validation_error: orgId is required field`)
				})
			})
			Context("DELETE", func() {
				BeforeEach(func() {
					method = "DELETE"
				})
				It("returns 400 code with proper message", func() {
					Expect(w.Code).To(Equal(http.StatusBadRequest))
					assertions.AssertAPIResponse(w.Body.String(), "Cannot delete vote (orgId is required field)", errory.OnCreateErrors.New(""))
					assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete vote, cause: ebt.validation_error: orgId is required field`)
				})
			})
		})

		Context("when body can be parsed and is valid", func() {
			var recommendationVote model.RecommendationVote
			BeforeEach(func() {
				requestBody = `{
					"orgId": 2,
					"recommendationType": "jiic",
					"vote": "like"
				}`
				recommendationVote = model.RecommendationVote{
					OrgID:              2,
					RecommendationType: "jiic",
					Vote:               model.Like,
				}
				recommendationVoteValidatorMock.EXPECT().Validate(gomock.Any(), recommendationVote).Times(1).Return(nil)
			})

			Context("when there is no user context", func() {
				Context("POST", func() {
					BeforeEach(func() {
						userContext = nil
						method = "POST"
					})
					It("returns 500 code with proper message", func() {
						Expect(w.Code).To(Equal(http.StatusInternalServerError))
						expError := errory.OnCreateErrors.Builder().Wrap(errory.ProcessingErrors.New("could not extract UserContext")).WithMessage("Cannot create vote").Create()
						assertions.AssertAPIResponse(w.Body.String(), "", expError)
						assertions.AssertLogger(logHook, `ebt.api_error.on_create_error: Cannot create vote, cause: ebt.processing_error: could not extract UserContext`)
					})
				})
				Context("DELETE", func() {
					BeforeEach(func() {
						userContext = nil
						method = "DELETE"
					})
					It("returns 500 code with proper message", func() {
						Expect(w.Code).To(Equal(http.StatusInternalServerError))
						expError := errory.OnDeleteErrors.Builder().Wrap(errory.ProcessingErrors.New("could not extract UserContext")).WithMessage("Cannot delete vote").Create()
						assertions.AssertAPIResponse(w.Body.String(), "", expError)
						assertions.AssertLogger(logHook, `ebt.api_error.on_delete_error: Cannot delete vote, cause: ebt.processing_error: could not extract UserContext`)
					})
				})
			})

			Context("when user context is provided", func() {
				BeforeEach(func() {
					userContext = &auth.UserContext{
						ID:     2,
						Cookie: "terefere",
					}
				})

				Context("POST", func() {
					BeforeEach(func() {
						method = "POST"
					})

					Context("when service fails", func() {
						BeforeEach(func() {
							recommendationVote.UserID = int64(2)
							recommendationVoteServiceMock.EXPECT().Create(&recommendationVote).Times(1).Return(errory.ProviderErrors.New("provider error when creating vote"))
						})

						It("returns 500 code with proper message", func() {
							Expect(w.Code).To(Equal(http.StatusInternalServerError))

							expError := errory.OnGetErrors.Builder().Wrap(errory.ProviderErrors.New("provider error when creating vote")).WithMessage("Cannot create vote").Create()
							assertions.AssertAPIResponse(w.Body.String(), "", expError)
						})
					})

					Context("when service succeeds", func() {
						BeforeEach(func() {
							recommendationVote.UserID = int64(2)
							recommendationVoteServiceMock.EXPECT().Create(&recommendationVote).Times(1).DoAndReturn(func(rv *model.RecommendationVote) error {
								rv.ID = 77
								return nil
							})
						})

						It("returns 201 code with proper message", func() {
							Expect(w.Code).To(Equal(http.StatusCreated))
							Expect(w.Body.String()).To(Equal(`{"id":77}`))
						})
					})
				})

				Context("DELETE", func() {
					BeforeEach(func() {
						method = "DELETE"
					})

					Context("when service fails", func() {
						BeforeEach(func() {
							recommendationVote.UserID = int64(2)
							recommendationVoteServiceMock.EXPECT().Delete(&recommendationVote).Times(1).Return(errory.ProviderErrors.New("provider error when deleting vote"))
						})

						It("returns 500 code with proper message", func() {
							Expect(w.Code).To(Equal(http.StatusInternalServerError))

							expError := errory.OnGetErrors.Builder().Wrap(errory.ProviderErrors.New("provider error when deleting vote")).WithMessage("Cannot delete vote").Create()
							assertions.AssertAPIResponse(w.Body.String(), "", expError)
						})
					})

					Context("when service succeeds", func() {
						BeforeEach(func() {
							recommendationVote.UserID = int64(2)
							recommendationVoteServiceMock.EXPECT().Delete(&recommendationVote).Times(1).DoAndReturn(func(rv *model.RecommendationVote) error {
								rv.ID = 77
								return nil
							})
						})

						It("returns 200 code with proper message", func() {
							Expect(w.Code).To(Equal(http.StatusOK))
							Expect(w.Body.String()).To(Equal(`{"id":77}`))
						})
					})
				})
			})
		})
	})
})
