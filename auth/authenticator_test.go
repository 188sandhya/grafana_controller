// +build unitTests

package auth_test

import (
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	. "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/assertions"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	authModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

var _ = Describe("Authenticator", func() {
	var mockController *gomock.Controller
	var mockAuthProvider *provider.MockIAuthProvider
	var mockGrafanaClient *grafana.MockIClient
	var mockIDAMClient *idam.MockIIDAMClient
	var authenticator auth.Authenticator
	logger, logHook := logrustest.NewNullLogger()

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAuthProvider = provider.NewMockIAuthProvider(mockController)
		mockGrafanaClient = grafana.NewMockIClient(mockController)
		mockIDAMClient = idam.NewMockIIDAMClient(mockController)
		authenticator = auth.Authenticator{
			Provider:   mockAuthProvider,
			Log:        logger,
			Grafana:    mockGrafanaClient,
			IDAMClient: mockIDAMClient,
		}
		logHook.Reset()
	})

	AfterEach(func() {
		mockController.Finish()
	})

	Describe("Authenticate(c *gin.Context)", func() {
		var header http.Header
		var context *gin.Context
		var userContext *authModel.UserContext
		var token string
		var err error
		var claims idam.StandardAndIdamClaims

		BeforeEach(func() {
			header = make(http.Header, 2)
			request := &http.Request{
				Header: header,
			}
			context = &gin.Context{
				Request: request,
			}
			role := make(map[string]idam.ContextSet)
			role[idam.RoleSpecificVerticalFullAccess] = idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"errorbudget"}}}
			claims = idam.StandardAndIdamClaims{
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
					Issuer:    "ds-test-setup",
					Audience:  "ds-prod",
					IssuedAt:  time.Now().Unix(),
				},
				UserPrincipalName: "test@metronom.com",
				Authorization:     idam.Roles{role},
				UserType:          "EMP",
			}

			token = "jhkfhjkshdheuyfgyud testToken"
		})

		JustBeforeEach(func() {
			userContext, err = authenticator.Authenticate(context)
		})

		Describe("Authenticate Basic auth", func() {
			Context("authoriztion is correctly passed in header", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
					mockGrafanaClient.EXPECT().Login("Aladdin", "open sesame").Times(1).Return("xyz", nil)
					mockAuthProvider.EXPECT().AuthenticateUser("xyz").Times(1).Return(int64(4), nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("xyz"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("authoriztion is incorrectly passed in header", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Basic QHAHHA==")
				})

				It("should not pass", func() {
					AssertErr(err, errory.AuthErrors.New("ebt.auth_error: authorization header incorrect"))
					Expect(err.Error()).To(ContainSubstring("ebt.auth_error: authorization header incorrect"))
				})
			})

			Context("authoriztion is incorrectly passed in header", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Basic YWRtaW46")
				})

				It("should not pass", func() {
					AssertErr(err, errory.AuthErrors.New("ebt.auth_error: authorization credenitals incorrect"))
					Expect(err.Error()).To(ContainSubstring("ebt.auth_error: authorization credenitals incorrect"))
				})
			})

			Context("authoriztion is correctly passed but logging in to grafana returns error", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
					mockGrafanaClient.EXPECT().Login("Aladdin", "open sesame").Times(1).Return("", errory.GrafanaClientAuthErrors.New("grafana client failure"))
				})

				It("should not pass", func() {
					AssertErr(err, errory.GrafanaClientAuthErrors.New("ebt.grafana_auth_error: grafana client failure"))
					Expect(err.Error()).To(ContainSubstring("ebt.grafana_auth_error: grafana client failure"))
				})
			})

			Context("authoriztion is correctly passed but authenticating cookie fails", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
					mockGrafanaClient.EXPECT().Login("Aladdin", "open sesame").Times(1).Return("xyz", nil)
					mockAuthProvider.EXPECT().AuthenticateUser("xyz").Times(1).Return(int64(0), errory.ProviderErrors.New("fatal provider error"))
				})

				It("should not pass", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})
		})

		Describe("Authenticate cookie", func() {
			Context("cookie is correctly passed in header", func() {
				BeforeEach(func() {
					header.Add("Cookie", "grafana_session=ABCDEZ")
					mockAuthProvider.EXPECT().AuthenticateUser("ABCDEZ").Times(1).Return(int64(4), nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("ABCDEZ"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("cookie is incorrectly passed in header", func() {
				BeforeEach(func() {
					header.Add("Cookie", "grapthite_session=ABCDEZ")
				})

				It("should not pass", func() {
					AssertErr(err, errory.AuthErrors.New("ebt.auth_error: grafana_session cookie not found, cause: http: named cookie not present"))
					Expect(err.Error()).To(ContainSubstring("ebt.auth_error: grafana_session cookie not found, cause: http: named cookie not present"))
				})
			})

			Context("cookie is correctly passed but authenticating cookie fails", func() {
				BeforeEach(func() {
					header.Add("Cookie", "grafana_session=ABCDEZ")
					mockAuthProvider.EXPECT().AuthenticateUser("ABCDEZ").Times(1).Return(int64(0), errory.ProviderErrors.New("fatal provider error"))
				})

				It("should not pass", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})
		})

		Describe("Authenticate bearer", func() {
			Context("authoriztion is correctly passed in header", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "test@metronom.com").Times(1).Return("uberCookie", nil)
					mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("uberCookie"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("authoriztion is correctly passed in header and there are multiple orgs", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
					orgRoles["other1"] = authModel.OrgRole{13, ""}
					orgRoles["custo"] = authModel.OrgRole{14, ""}
					orgRoles["default"] = authModel.OrgRole{1, ""}

					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "test@metronom.com").Times(1).Return("uberCookie", nil)
					mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
					mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{1, "Viewer"}).Times(1).Return(nil)
					mockAuthProvider.EXPECT().UpdateUserRole(int64(4), authModel.OrgRole{12, "Editor"}).Times(1).Return(nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("uberCookie"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("client authorization with client token with RoleOMAViewAll", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
					orgRoles["other1"] = authModel.OrgRole{13, ""}
					orgRoles["custo"] = authModel.OrgRole{14, ""}

					claims.UserType = "CLIENT"

					claims.Subject = "errorbudget"
					claims.Realm = "2TR_PENG"
					claims.Authorization = idam.Roles{map[string]idam.ContextSet{idam.RoleOMAViewAll: idam.ContextSet{}}}

					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "errorbudget@2TR_PENG").Times(1).Return("uberCookie", nil)
					mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
					mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{13, "Viewer"}).Times(1).Return(nil)
					mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{14, "Viewer"}).Times(1).Return(nil)
					//mockAuthProvider.EXPECT().UpdateUserRole(int64(4), authModel.OrgRole{12, "Editor"}).Times(1).Return(nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("uberCookie"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("User created", func() {
				Context("client authorization with client token with RoleOMAViewAll and customer to other1", func() {
					BeforeEach(func() {
						orgRoles := make(map[string]authModel.OrgRole)
						orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
						orgRoles["other1"] = authModel.OrgRole{13, ""}
						orgRoles["custo"] = authModel.OrgRole{14, ""}

						claims.UserType = "CLIENT"

						claims.Subject = "errorbudget"
						claims.Realm = "2TR_PENG"

						verticalRole := idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"other1"}}}

						roleMap := map[string]idam.ContextSet{idam.RoleOMAViewAll: idam.ContextSet{},
							idam.RoleSpecificVerticalFullAccess: verticalRole}
						claims.Authorization = idam.Roles{roleMap}

						header.Add("Authorization", "Bearer "+token)
						mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
						mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), true, nil)
						mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
						mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "errorbudget@2TR_PENG").Times(1).Return("uberCookie", nil)
						mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{13, "Editor"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{14, "Viewer"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().UpdateUserDetails(int64(4), int64(13), false).Times(1).Return(nil)
					})

					It("should pass", func() {
						Expect(userContext.ID).To(Equal(int64(4)))
						Expect(userContext.Cookie).To(Equal("uberCookie"))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("client authorization with  customer to other1", func() {
					BeforeEach(func() {
						orgRoles := make(map[string]authModel.OrgRole)
						orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
						orgRoles["other1"] = authModel.OrgRole{13, ""}
						orgRoles["custo"] = authModel.OrgRole{14, ""}

						claims.UserType = "CLIENT"

						claims.Subject = "errorbudget"
						claims.Realm = "2TR_PENG"

						verticalRole := idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"other1"}}}

						roleMap := map[string]idam.ContextSet{
							idam.RoleSpecificVerticalFullAccess: verticalRole}
						claims.Authorization = idam.Roles{roleMap}

						header.Add("Authorization", "Bearer "+token)
						mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
						mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), true, nil)
						mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
						mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "errorbudget@2TR_PENG").Times(1).Return("uberCookie", nil)
						mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{13, "Editor"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().UpdateUserDetails(int64(4), int64(13), false).Times(1).Return(nil)
					})

					It("should pass", func() {
						Expect(userContext.ID).To(Equal(int64(4)))
						Expect(userContext.Cookie).To(Equal("uberCookie"))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("client authorization with customer to other1 and OMA Admin role", func() {
					BeforeEach(func() {
						orgRoles := make(map[string]authModel.OrgRole)
						orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
						orgRoles["other1"] = authModel.OrgRole{13, ""}
						orgRoles["custo"] = authModel.OrgRole{14, ""}

						claims.UserType = "CLIENT"

						claims.Subject = "errorbudget"
						claims.Realm = "2TR_PENG"

						verticalRole := idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"other1"}}}

						roleMap := map[string]idam.ContextSet{
							idam.RoleOMASuperAdmin:              idam.ContextSet{},
							idam.RoleSpecificVerticalFullAccess: verticalRole}
						claims.Authorization = idam.Roles{roleMap}

						header.Add("Authorization", "Bearer "+token)
						mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
						mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), true, nil)
						mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
						mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "errorbudget@2TR_PENG").Times(1).Return("uberCookie", nil)
						mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
						mockAuthProvider.EXPECT().UpdateUserRole(int64(4), authModel.OrgRole{12, "Admin"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{13, "Admin"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{14, "Admin"}).Times(1).Return(nil)
						mockAuthProvider.EXPECT().UpdateUserDetails(int64(4), int64(13), true).Times(1).Return(nil)
					})

					It("should pass", func() {
						Expect(userContext.ID).To(Equal(int64(4)))
						Expect(userContext.Cookie).To(Equal("uberCookie"))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("client authorization with customer to default first org and with no OMA Admin role", func() {
					BeforeEach(func() {
						orgRoles := make(map[string]authModel.OrgRole)
						orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
						orgRoles["other1"] = authModel.OrgRole{1, ""}
						orgRoles["custo"] = authModel.OrgRole{14, ""}

						claims.UserType = "CLIENT"

						claims.Subject = "errorbudget"
						claims.Realm = "2TR_PENG"

						verticalRole := idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"other1"}}}

						roleMap := map[string]idam.ContextSet{
							idam.RoleSpecificVerticalFullAccess: verticalRole}
						claims.Authorization = idam.Roles{roleMap}

						header.Add("Authorization", "Bearer "+token)
						mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
						mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), true, nil)
						mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
						mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "errorbudget@2TR_PENG").Times(1).Return("uberCookie", nil)
						mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
						mockAuthProvider.EXPECT().CreateUserRole(int64(4), authModel.OrgRole{1, "Editor"}).Times(1).Return(nil)
					})

					It("should pass", func() {
						Expect(userContext.ID).To(Equal(int64(4)))
						Expect(userContext.Cookie).To(Equal("uberCookie"))
						Expect(err).ToNot(HaveOccurred())
					})
				})

				Context("client authorization with customer to other1 and OMA Admin role - but provider returns an error", func() {
					BeforeEach(func() {
						orgRoles := make(map[string]authModel.OrgRole)
						orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
						orgRoles["other1"] = authModel.OrgRole{13, ""}
						orgRoles["custo"] = authModel.OrgRole{14, ""}

						claims.UserType = "CLIENT"

						claims.Subject = "errorbudget"
						claims.Realm = "2TR_PENG"

						verticalRole := idam.ContextSet{idam.ContextCombination{"vertical": idam.ContextValues{"other1"}}}

						roleMap := map[string]idam.ContextSet{
							idam.RoleOMASuperAdmin:              idam.ContextSet{},
							idam.RoleSpecificVerticalFullAccess: verticalRole}
						claims.Authorization = idam.Roles{roleMap}

						header.Add("Authorization", "Bearer "+token)
						mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
						mockAuthProvider.EXPECT().FindOrCreateUser("errorbudget@2TR_PENG").Times(1).Return(int64(4), true, nil)
						mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
						mockAuthProvider.EXPECT().UpdateUserDetails(int64(4), int64(13), true).Times(1).Return(errory.ProviderErrors.New("fatal provider error"))
					})

					It("should return correct error", func() {
						AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
						Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
					})
				})
			})

			Context("user authorization with no roles", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}
					orgRoles["other1"] = authModel.OrgRole{13, ""}
					orgRoles["custo"] = authModel.OrgRole{14, ""}

					claims.Authorization = idam.Roles{}

					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "test@metronom.com").Times(1).Return("uberCookie", nil)
					mockAuthProvider.EXPECT().CreateFakeOauthLogin(int64(4)).Times(1).Return(nil)
				})

				It("should pass", func() {
					Expect(userContext.ID).To(Equal(int64(4)))
					Expect(userContext.Cookie).To(Equal("uberCookie"))
					Expect(err).ToNot(HaveOccurred())
				})
			})

			Context("token can't be parsed", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(nil, false, "parsing bearer token failed")
				})

				It("should return correct error", func() {
					AssertErr(err, errory.AuthErrors.New("ebt.auth_error: bearer token incorrect: parsing bearer token failed"))
					Expect(err.Error()).To(ContainSubstring("ebt.auth_error: bearer token incorrect: parsing bearer token failed"))
				})
			})

			Context("provider is returning error", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(0), false, errory.ProviderErrors.New("fatal provider error"))
				})

				It("should pass", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})

			Context("provider is returning error on GetOrgRoles", func() {
				BeforeEach(func() {
					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), true, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(map[string]authModel.OrgRole{}, errory.ProviderErrors.New("fatal provider error"))
				})

				It("should return correct error", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})

			Context("provider is returning error on UpdateUserRole", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					orgRoles["errorbudget"] = authModel.OrgRole{12, "Viewer"}

					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().UpdateUserRole(int64(4), authModel.OrgRole{12, "Editor"}).Times(1).Return(errory.ProviderErrors.New("fatal provider error"))
				})

				It("should return correct error", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})

			Context("provider is returning error on FindOrCreateSession", func() {
				BeforeEach(func() {
					orgRoles := make(map[string]authModel.OrgRole)
					header.Add("Authorization", "Bearer "+token)
					mockIDAMClient.EXPECT().TokenAuthenticator(token).Times(1).Return(&claims, true, "")
					mockAuthProvider.EXPECT().FindOrCreateUser("test@metronom.com").Times(1).Return(int64(4), false, nil)
					mockAuthProvider.EXPECT().GetOrgRoles(int64(4)).Times(1).Return(orgRoles, nil)
					mockAuthProvider.EXPECT().FindOrCreateSession(int64(4), "test@metronom.com").Times(1).Return("", errory.ProviderErrors.New("fatal provider error"))

				})

				It("should return correct error", func() {
					AssertErr(err, errory.ProviderErrors.New("ebt.provider_error: fatal provider error"))
					Expect(err.Error()).To(ContainSubstring("ebt.provider_error: fatal provider error"))
				})
			})

		})
	})
})
