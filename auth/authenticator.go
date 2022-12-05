package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/sirupsen/logrus"
)

const bearerPrefix = "Bearer "

type IAuthenticator interface {
	Authenticate(c *gin.Context) (*auth.UserContext, error)
}

type Authenticator struct {
	IDAMClient idam.IIDAMClient
	Provider   provider.IAuthProvider
	Log        logrus.FieldLogger
	Grafana    grafana.IClient
}

func (a *Authenticator) Authenticate(c *gin.Context) (userContext *auth.UserContext, err error) {
	var userID int64

	cookie, username, password, token, err := extractCredentials(c)
	if err != nil {
		return
	}

	switch {
	case cookie != "":
		userID, err = a.Provider.AuthenticateUser(cookie)

	case username != "" && password != "":
		cookie, err = a.Grafana.Login(username, password)
		if err != nil {
			return
		}

		userID, err = a.Provider.AuthenticateUser(cookie)

	case token != "":
		userID, cookie, err = a.processToken(token)
		if err != nil {
			return
		}

	default:
		return nil, errory.AuthErrors.New("authorization credenitals incorrect")
	}

	return &auth.UserContext{ID: userID, Cookie: cookie}, err
}

func extractCredentials(c *gin.Context) (cookie, username, password, token string, err error) {
	if c.Request.Header.Get("Authorization") != "" {
		var ok bool
		username, password, ok = c.Request.BasicAuth()
		if ok {
			return
		} else if authorization := c.Request.Header.Get("Authorization"); len(authorization) >= len(bearerPrefix) &&
			strings.EqualFold(authorization[:len(bearerPrefix)], bearerPrefix) {
			token = authorization[len(bearerPrefix):]
			return
		}
		err = errory.AuthErrors.New("authorization header incorrect")
		return
	}

	cookie, err = c.Cookie("grafana_session")
	if err != nil {
		err = errory.AuthErrors.Builder().Wrap(err).WithMessage("grafana_session cookie not found").Create()
	}

	return
}

func (a *Authenticator) processToken(token string) (userID int64, cookie string, err error) {
	claims, username, err := a.validateToken(token)
	if err != nil {
		return
	}

	userID, err = a.configureUser(claims, username)
	if err != nil {
		return
	}

	cookie, err = a.Provider.FindOrCreateSession(userID, username)
	if err != nil {
		return
	}

	err = a.Provider.CreateFakeOauthLogin(userID)

	a.Log.Infof("User %s logged in with token", username)

	return
}

func (a *Authenticator) validateToken(token string) (*idam.StandardAndIdamClaims, string, error) {
	claims, ok, message := a.IDAMClient.TokenAuthenticator(token)
	if !ok {
		return nil, "", errory.AuthErrors.New("bearer token incorrect: %s", message)
	}

	var username string
	if claims.UserType == idam.UserTypeClient {
		username = claims.Subject + "@" + claims.Realm
	} else {
		username = claims.UserPrincipalName
	}

	return claims, username, nil
}

func (a *Authenticator) configureUser(claims *idam.StandardAndIdamClaims, username string) (int64, error) {
	userID, created, err := a.Provider.FindOrCreateUser(username)
	if err != nil {
		return 0, err
	}

	orgAccess, err := a.getOrgAccess(claims.Authorization, userID)
	if err != nil {
		return 0, err
	}

	if created {
		err = a.setupNewUser(userID, orgAccess)
		if err != nil {
			return 0, err
		}
	}

	err = a.configureUserOrgs(orgAccess, userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (a *Authenticator) configureUserOrgs(orgAccess auth.OrgAccess, userID int64) error {
	allVerticals, verticals, currentRoles, omaAdmin := orgAccess.AllVerticals, orgAccess.Verticals, orgAccess.Roles, orgAccess.GrafanaAdmin

	var err error

	for org, orgRole := range currentRoles {
		switch {
		case orgRole.Role == "" && omaAdmin:
			err = a.Provider.CreateUserRole(userID, auth.OrgRole{OrgID: orgRole.OrgID, Role: auth.GFAdminRole})
		case orgRole.Role != auth.GFAdminRole && omaAdmin:
			err = a.Provider.UpdateUserRole(userID, auth.OrgRole{OrgID: orgRole.OrgID, Role: auth.GFAdminRole})
		case orgRole.Role == "" && (verticals[org]):
			err = a.Provider.CreateUserRole(userID, auth.OrgRole{OrgID: orgRole.OrgID, Role: auth.GFEditorRole})
		case orgRole.Role == auth.GFViewerRole && (verticals[org]):
			err = a.Provider.UpdateUserRole(userID, auth.OrgRole{OrgID: orgRole.OrgID, Role: auth.GFEditorRole})
		case orgRole.Role == "" && (orgRole.OrgID == 1 || allVerticals):
			err = a.Provider.CreateUserRole(userID, auth.OrgRole{OrgID: orgRole.OrgID, Role: auth.GFViewerRole})
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func extractVerticals(roles idam.Roles) (allVerticals bool, verticals map[string]bool) {
	verticals = make(map[string]bool)

	allVerticals, _, _ = roles.GetRole(idam.RoleOMAViewAll)

	found, customers, _ := roles.GetRole(idam.RoleSpecificVerticalFullAccess)
	if !found {
		return
	}

	for _, context := range customers {
		values, exist := context[idam.ContextVertical]
		if exist {
			for _, vertical := range values {
				verticals[vertical] = true
			}
		}
	}

	return
}

func (a *Authenticator) getOrgAccess(roles idam.Roles, userID int64) (auth.OrgAccess, error) {
	orgAccess := auth.OrgAccess{}

	allVerticals, verticals := extractVerticals(roles)
	currentRoles, err := a.Provider.GetOrgRoles(userID)
	if err != nil {
		return orgAccess, err
	}
	grafanaAdmin, _, _ := roles.GetRole(idam.RoleOMASuperAdmin)

	orgAccess.AllVerticals = allVerticals
	orgAccess.Verticals = verticals
	orgAccess.Roles = currentRoles
	orgAccess.GrafanaAdmin = grafanaAdmin

	return orgAccess, nil
}

func (a *Authenticator) setupNewUser(userID int64, orgAccess auth.OrgAccess) error {
	defaultOrgID := int64(1)
	for org, orgRole := range orgAccess.Roles {
		if orgAccess.Verticals[org] {
			defaultOrgID = orgRole.OrgID
			break
		}
	}
	if defaultOrgID != int64(1) || orgAccess.GrafanaAdmin {
		a.Log.Infof("Assigned org id: %d, isAdmin: %t to user %d", defaultOrgID, orgAccess.GrafanaAdmin, userID)
		return a.Provider.UpdateUserDetails(userID, defaultOrgID, orgAccess.GrafanaAdmin)
	}
	return nil
}
