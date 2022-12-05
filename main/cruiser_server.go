//nolint:funlen,gocritic,unparam,nolintlint
package main

import (
	"net/http"
	"time"

	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"

	_ "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/docs"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	authModel "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CruiserServer struct {
	OrgAPI                 *api.OrgAPI
	SloAPI                 *api.SloAPI
	FeedbackAPI            *api.FeedbackAPI
	HappinessMetricAPI     *api.HappinessMetricAPI
	RecommendationVoteAPI  *api.RecommendationVoteAPI
	HealthAPI              *api.HealthAPI
	ConfigureUserAPI       *api.ConfigureUserAPI
	Log                    logrus.FieldLogger
	Authenticator          auth.IAuthenticator
	DatasourceAPI          *api.DatasourceAPI
	PluginAPI              *api.PluginAPI
	SDAAPI                 *api.SDAAPI
	SolutionsAPI           *api.SolutionsAPI
	SolutionSloAPI         *api.SolutionSloAPI
	ProductsStatusAPI      *api.ProductsStatusAPI
	Authorizer             middleware.IAuthorizer
	Cors                   Cors
	ParamExistCheckService service.IParamExistCheckService
}

var prometheus *middleware.Prometheus

type Group struct {
	*gin.RouterGroup
}

func addEndpoint(fullPath, method string) {
	prometheus.Add(fullPath, method)
}

func (group *Group) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	addEndpoint(group.BasePath()+relativePath, http.MethodGet)
	return group.RouterGroup.
		GET(relativePath, handlers...)
}

func (group *Group) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	addEndpoint(group.BasePath()+relativePath, http.MethodPost)
	return group.RouterGroup.
		POST(relativePath, handlers...)
}

func (group *Group) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	addEndpoint(group.BasePath()+relativePath, http.MethodPut)
	return group.RouterGroup.
		PUT(relativePath, handlers...)
}

func (group *Group) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	addEndpoint(group.BasePath()+relativePath, http.MethodDelete)
	return group.RouterGroup.
		DELETE(relativePath, handlers...)
}

// @title OMA Tool API
// @version 1.0
// @description Manage your SLOs & Metrics

// @contact.name OMA Team
// @contact.url
// @license.name ​
// @host oma.metro.digital

// @BasePath /ebt/v1
func (s *CruiserServer) Run() error {
	authenticate := middleware.Authenticate(s.Authenticator, s.Log)
	checkContentType := middleware.ContentType()
	timeout := 120 * time.Second

	authorize := middleware.Authorize
	idParam := api.GetIDParam
	paramExistChecker := middleware.CheckIfExist

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	prometheus = middleware.NewPrometheus("/metrics", "cruiser", s.Log)
	prometheus.Use(r)

	r.UseRawPath = true
	r.Use(middleware.Tracing)

	r.Use(gin.HandlerFunc(s.Cors))

	health := Group{r.Group("/api")}
	health.Use(middleware.Recovery(s.Log))
	health.GET("/health", s.HealthAPI.HealthCheck)

	r.Use(middleware.LoggerMiddleware(s.Log, timeout))

	r.StaticFile("/grafana_source", "/doc/oma_grafana_source.zip")

	docs := Group{r.Group("/openapi")}
	docs.Use(middleware.AppJSONHeader())
	docs.StaticFile("/doc.json", "/doc/openapi.json")

	ar := r.Group("")
	ar.Use(gzip.Gzip(gzip.DefaultCompression))
	ar.Use(middleware.Recovery(s.Log), authenticate)

	ar.GET("/v1/configure_user", s.ConfigureUserAPI.ConfigureUser)

	pluginRoutes := Group{ar.Group("/v1/plugin")}
	{
		pluginRoutes.POST("/:id", authorize(idParam, s.Authorizer.AuthorizeForOrganization, authModel.Admin, s.Log), s.PluginAPI.Enable)
	}

	sdaRoutes := Group{ar.Group("/v1/sda")}
	{
		sdaRoutes.GET("", s.SDAAPI.GetConfig)
	}

	fsRoutes := Group{ar.Group("/v1/solutions")}
	{
		fsRoutes.GET("", s.SolutionsAPI.GetSolutions)
	}

	solRoutes := Group{ar.Group("/v1/solutionSlo")}
	{
		solRoutes.GET("", s.SolutionSloAPI.GetSolutionSlo)
	}

	psRoutes := Group{ar.Group("/v1/products_status")}
	{
		psRoutes.GET("", s.ProductsStatusAPI.GetProductsStatus)
	}

	dsRoutes := Group{ar.Group("/v1/datasource")}
	{
		dsRoutes.GET("/:id", authorize(idParam, s.Authorizer.AuthorizeForDatasource, authModel.Viewer, s.Log), s.DatasourceAPI.Get)
	}

	organizationRoutes := Group{ar.Group("/v1/org")}
	{
		organizationRoutes.GET("/:id", authorize(idParam, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log), s.OrgAPI.GetOrg)
		organizationRoutes.GET("/:id/slo", authorize(idParam, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log),
			paramExistChecker(idParam, s.ParamExistCheckService, service.Organization), s.OrgAPI.GetSlos)
		organizationRoutes.POST("/:id/slo", checkContentType, authorize(idParam, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log),
			paramExistChecker(idParam, s.ParamExistCheckService, service.Organization), s.OrgAPI.FindSlos)
		organizationRoutes.GET("/:id/datasource", authorize(idParam, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log),
			paramExistChecker(idParam, s.ParamExistCheckService, service.Organization), s.OrgAPI.GetDatasources)

		organizationRoutes.GET("/:id/user_happiness", s.OrgAPI.GetAllHappinessMetricsForUser)
		organizationRoutes.GET("/:id/team_happiness", s.OrgAPI.GetAllHappinessMetricsForTeam)
		organizationRoutes.POST("/:id/team_happiness/average", authorize(idParam, s.Authorizer.AuthorizeForTeam, authModel.Viewer, s.Log), s.OrgAPI.SaveTeamAverage)
		organizationRoutes.GET("/:id/team_happiness/missing", authorize(idParam, s.Authorizer.AuthorizeForTeam, authModel.Viewer, s.Log), s.OrgAPI.GetUsersMissingInput)
	}

	sloRoutes := Group{ar.Group("/v1/slo")}
	{
		sloRoutes.GET("", s.SloAPI.GetDetailed)
		sloRoutes.POST("", checkContentType, authorize(middleware.OrgIDInStruct, s.Authorizer.AuthorizeForOrganization, authModel.Editor, s.Log), s.SloAPI.Create)
		sloRoutes.PUT("/:id", checkContentType, authorize(idParam, s.Authorizer.AuthorizeForSLO, authModel.Editor, s.Log), s.SloAPI.Update)
		sloRoutes.GET("/:id", s.SloAPI.Get)
		sloRoutes.DELETE("/:id", authorize(idParam, s.Authorizer.AuthorizeForSLO, authModel.Editor, s.Log), s.SloAPI.Delete)
		sloRoutes.DELETE("/:id/history", authorize(idParam, s.Authorizer.AuthorizeForSLO, authModel.Editor, s.Log), s.SloAPI.DeleteSloHistory)
	}

	feedbackRoutes := Group{ar.Group("/v1/feedback")}
	{
		feedbackRoutes.POST("", checkContentType,
			authorize(middleware.OrgIDInStruct, s.Authorizer.AuthorizeForTeam, authModel.Editor, s.Log), s.FeedbackAPI.Create)
		feedbackRoutes.GET("/:id", s.FeedbackAPI.Get)
	}

	happinessMetricRoutes := Group{ar.Group("/v1/happiness_metric")}
	{
		happinessMetricRoutes.POST("", checkContentType,
			authorize(middleware.OrgIDInStruct, s.Authorizer.AuthorizeForTeam, authModel.Editor, s.Log), s.HappinessMetricAPI.Create)
		happinessMetricRoutes.PUT("/:id", checkContentType,
			authorize(idParam, s.Authorizer.AuthorizeForHappinessMetric, authModel.Editor, s.Log), s.HappinessMetricAPI.Update)
		happinessMetricRoutes.DELETE("/:id",
			authorize(idParam, s.Authorizer.AuthorizeForHappinessMetric, authModel.Editor, s.Log), s.HappinessMetricAPI.Delete)
		happinessMetricRoutes.GET("/:id", s.HappinessMetricAPI.Get)
	}

	recommendationVoteRoutes := Group{ar.Group("/v1/recommendation_vote")}
	{
		recommendationVoteRoutes.GET("", s.RecommendationVoteAPI.Get)
		recommendationVoteRoutes.POST("", checkContentType,
			authorize(middleware.OrgIDInStruct, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log), s.RecommendationVoteAPI.Create)
		recommendationVoteRoutes.DELETE("", checkContentType,
			authorize(middleware.OrgIDInStruct, s.Authorizer.AuthorizeForOrganization, authModel.Viewer, s.Log), s.RecommendationVoteAPI.Delete)
	}

	srv := &http.Server{
		Addr:              ":8000",
		Handler:           middleware.TimeoutLogHandler(http.TimeoutHandler(r, timeout, "Timeout Exceeded"), s.Log, timeout),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      130 * time.Second,
	}

	return srv.ListenAndServe()
}
