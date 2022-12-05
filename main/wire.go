//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/api"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/elastic"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	pluginService "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service/plugin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
)

var configsSet = wire.NewSet(
	readInPluginConfig,
)

var clientsSet = wire.NewSet(
	newGrafanaClient, wire.Bind(new(grafana.IClient), new(*grafana.Client)),
	newSDAElasticClient, wire.Bind(new(elastic.IClient), new(*elastic.Client)),
	newIDAMClient, wire.Bind(new(idam.IIDAMClient), new(*idam.IDAMRestClient)),
)

var servicesSet = wire.NewSet(
	wire.Struct(new(service.SloService), "*"), wire.Bind(new(service.ISloService), new(*service.SloService)),
	wire.Struct(new(service.DatasourceService), "*"), wire.Bind(new(service.IDatasourceService), new(*service.DatasourceService)),
	wire.Struct(new(service.OrganizationService), "*"), wire.Bind(new(service.IOrganizationService), new(*service.OrganizationService)),
	newDashboardService, wire.Bind(new(service.IDashboardService), new(*service.DashboardService)),
	wire.Struct(new(service.ParamExistCheckService), "*"), wire.Bind(new(service.IParamExistCheckService), new(*service.ParamExistCheckService)),
	newPluginService, wire.Bind(new(pluginService.IPluginService), new(*pluginService.Plugin)),
	newDSParser, wire.Bind(new(pluginService.IDatasourceParser), new(*pluginService.DatasourceParser)),
	wire.Struct(new(service.SDAService), "*"), wire.Bind(new(service.ISDAService), new(*service.SDAService)),
	wire.Struct(new(service.HealthService), "*"), wire.Bind(new(service.IHealthService), new(*service.HealthService)),
	wire.Struct(new(service.FeedbackService), "*"), wire.Bind(new(service.IFeedbackService), new(*service.FeedbackService)),
	wire.Struct(new(service.HappinessMetricService), "*"), wire.Bind(new(service.IHappinessMetricService), new(*service.HappinessMetricService)),
	wire.Struct(new(service.UserInfoService), "*"), wire.Bind(new(service.IUserInfoService), new(*service.UserInfoService)),
	wire.Struct(new(service.SolutionsService), "*"), wire.Bind(new(service.ISolutionsService), new(*service.SolutionsService)),
	wire.Struct(new(service.SolutionSloService), "*"), wire.Bind(new(service.ISolutionSloService), new(*service.SolutionSloService)),
	wire.Struct(new(service.RecommendationVoteService), "*"), wire.Bind(new(service.IRecommendationVoteService), new(*service.RecommendationVoteService)),
	wire.Struct(new(service.ProductsStatusService), "*"), wire.Bind(new(service.IProductsStatusService), new(*service.ProductsStatusService)),
)

var validatorsSet = wire.NewSet(
	validator.NewSLOValidator, wire.Bind(new(validator.ISLOValidator), new(*validator.SLOValidator)),
	validator.NewFeedbackValidator, wire.Bind(new(validator.IFeedbackValidator), new(*validator.FeedbackValidator)),
	validator.NewHappinessMetricValidator, wire.Bind(new(validator.IHappinessMetricValidator), new(*validator.HappinessMetricValidator)),
	validator.NewValidator, wire.Bind(new(validator.ITranslatedValidator), new(*validator.TranslatedValidator)),
)

var apisSet = wire.NewSet(
	wire.Struct(new(api.OrgAPI), "*"),
	wire.Struct(new(api.SloAPI), "*"),
	wire.Struct(new(api.HealthAPI), "*"),
	wire.Struct(new(api.ConfigureUserAPI), "*"),
	wire.Struct(new(api.DatasourceAPI), "*"),
	wire.Struct(new(api.PluginAPI), "*"),
	wire.Struct(new(api.SDAAPI), "*"),
	wire.Struct(new(api.SolutionsAPI), "*"),
	wire.Struct(new(api.FeedbackAPI), "*"),
	wire.Struct(new(api.HappinessMetricAPI), "*"),
	wire.Struct(new(api.SolutionSloAPI), "*"),
	wire.Struct(new(api.RecommendationVoteAPI), "*"),
	wire.Struct(new(api.ProductsStatusAPI), "*"),
)

var providerSet = wire.NewSet(
	newDBConnection,
	wire.Struct(new(provider.SQL), "*"),
	wire.Bind(new(provider.IOrganizationProvider), new(*provider.SQL)),
	wire.Bind(new(provider.ISLOProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IDatasourceProvider), new(*provider.SQL)),
	wire.Bind(new(provider.ISDAProvider), new(*provider.SQL)),
	wire.Bind(new(provider.ISolutionsProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IHealthProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IFeedbackProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IRecommendationVoteProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IHappinessMetricProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IUserInfoProvider), new(*provider.SQL)),
	newAuthProvider, wire.Bind(new(provider.IAuthProvider), new(*provider.AuthProvider)),
	wire.Bind(new(middleware.IAuthorizer), new(*provider.AuthProvider)),
	wire.Bind(new(provider.ISolutionSloProvider), new(*provider.SQL)),
	wire.Bind(new(provider.IProductsStatusProvider), new(*provider.SQL)),
)

var othersSet = wire.NewSet(
	newAuthenticator, wire.Bind(new(auth.IAuthenticator), new(*auth.Authenticator)),
	createCors,
	createLoggerWithStandardFields,
)

func initApp() (*CruiserServer, error) {
	wire.Build(configsSet, clientsSet, servicesSet, validatorsSet, apisSet, othersSet, providerSet, wire.Struct(new(CruiserServer), "*"))

	return &CruiserServer{}, nil
}
