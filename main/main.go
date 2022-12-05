package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/elastic"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/client/idam"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	pluginService "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service/plugin"
	"github.com/spf13/viper"
)

var logger *logrus.Logger

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap:        logrus.FieldMap{logrus.FieldKeyTime: "@timestamp"},
	})
	initConfig()

	cruiser, err := initApp()

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Fatal(cruiser.Run())
}

func newGrafanaClient() *grafana.Client {
	var (
		baseAPIURL = viper.GetString("grafana_base_api_url")
	)
	grafanaClient, err := grafana.New(baseAPIURL)

	if err != nil {
		return nil
	}

	return grafanaClient
}

func newSDAElasticClient(pluginCfg *model.PluginConfig, log logrus.FieldLogger) *elastic.Client {
	if pluginCfg.SDA == nil {
		return nil
	}

	elasticURL := viper.GetString("ELASTICSEARCH_DS_URL")

	elasticClient, err := elastic.New(elasticURL, log)
	if err != nil {
		return nil
	}
	return elasticClient
}

func newDBConnection(log logrus.FieldLogger) *sqlx.DB {
	postgresURL := viper.GetString("postgres_url")
	db, err := sqlx.Connect("postgres", postgresURL)
	if err != nil {
		log.WithError(err).WithField("ENV_VARIABLE", "POSTGRES_URL").Error("Could not connect to Database")
		return nil
	}

	db.SetMaxIdleConns(20)
	return db
}

func newAuthProvider(sql *sqlx.DB, log logrus.FieldLogger) *provider.AuthProvider {
	grafanaSecret := viper.GetString("grafana_secret")
	return provider.NewAuthProvider(sql, log, grafanaSecret)
}

func newDashboardService(log logrus.FieldLogger, dp provider.IDatasourceProvider,
	ap provider.IAuthProvider, g grafana.IClient) *service.DashboardService {
	resourcePath := viper.GetString("resource_path")
	return &service.DashboardService{
		DatasourceProvider: dp,
		Grafana:            g,
		ResourcePath:       resourcePath,
		Log:                log,
	}
}
func newPluginService(log logrus.FieldLogger, g grafana.IClient, e elastic.IClient,
	ds provider.IDatasourceProvider, sc provider.ISDAProvider, ap provider.IAuthProvider,
	cfg *model.PluginConfig, dsParser pluginService.IDatasourceParser) *pluginService.Plugin {
	resourcePath := viper.GetString("resource_path")

	return &pluginService.Plugin{
		Elastic:            e,
		Grafana:            g,
		DatasourceProvider: ds,
		SDAProvider:        sc,
		Provider:           ap,
		Config:             cfg,
		ResourcePath:       resourcePath,
		Log:                log,
		DSParser:           dsParser,
	}
}

func newDSParser() *pluginService.DatasourceParser {
	elasticURL := viper.GetString("ELASTICSEARCH_DS_URL")
	ddAPIKey := viper.GetString("datadog_api_key")
	ddAppKey := viper.GetString("datadog_app_key")
	a, b := abc()
	fmt.Println(a, b)
	return &pluginService.DatasourceParser{
		ElasticsearchDatasourceURL: elasticURL,
		DatadogAPIKey:              ddAPIKey,
		DatadogAPPKey:              ddAppKey,
	}

}

func xx(a, b int) bool {
	return true
}

func abc() (a int, b int) {
	x := 00
	var y int = 4

	z := x + y

	return z, b
}

func newIDAMClient(log logrus.FieldLogger) *idam.IDAMRestClient {
	idamURL := viper.GetString("idam_base_url")
	idamRestClient := idam.NewIDAMClient(idamURL, logger, "")
	idamRestClient.Startup()
	return idamRestClient
}

func newAuthenticator(idamClient idam.IIDAMClient, p provider.IAuthProvider, g grafana.IClient, log logrus.FieldLogger) *auth.Authenticator {
	return &auth.Authenticator{
		IDAMClient: idamClient,
		Provider:   p,
		Log:        log,
		Grafana:    g,
	}
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetDefault("grafana_base_api_url", "http://grafana:3000/")
	viper.SetDefault("elasticsearch_ds_url", "http://elastic:9200")
	viper.SetDefault("resource_path", "/resource/")
	viper.SetDefault("datadog_address", "https://datadoghq.com")
	viper.SetDefault("postgres_url", "postgres://postgres:example@postgres:5432/grafana?sslmode=disable")
	viper.SetDefault("host_url", "localhost")
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("slo_value_precision", 3)
	viper.SetDefault("service_version", "latest")
	viper.SetDefault("allowed_origins", []string{"*"})

	viper.SetConfigType("yaml")
	viper.SetConfigName("cruiser")
	viper.AddConfigPath("/config")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Printf("error reading cruiser config file: %s \n", err.Error())
	} else {
		viper.WatchConfig()
		viper.OnConfigChange(onConfigChange)
	}
}

func readInPluginConfig() *model.PluginConfig {
	viper.SetConfigType("yaml")
	viper.SetConfigName("plugin")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Printf("error reading plugin config file: %s \n", err.Error())
	}
	config := &model.PluginConfig{}
	err = viper.Unmarshal(config)
	if err != nil {
		logrus.Printf("error unmarshaling plugin config file: %s \n", err.Error())
	}

	return config
}

func onConfigChange(e fsnotify.Event) {
	logrus.Println("Config file changed:", e.Name)
	loggingLevel := viper.GetString("logger.level")
	setLoggingLevel(loggingLevel)
}

func setLoggingLevel(level string) {
	envLoggingLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.WithError(err).Println("error parsing logging level")
	}
	logrus.WithField("new_level", level).Println("Logrus level changed")
	logger.SetLevel(envLoggingLevel)
}

func createLoggerWithStandardFields() logrus.FieldLogger {
	logger = logrus.New()
	logger.Formatter = &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap:        logrus.FieldMap{logrus.FieldKeyTime: "@timestamp"},
	}

	withFields := logger.WithFields(logrus.Fields{
		middleware.LogfieldOrganization:   "errorbudget",
		middleware.LogfieldServiceName:    "grafana-controller",
		middleware.LogfieldHostname:       "grafana-controller",
		middleware.LogfieldServiceVersion: viper.GetString("service_version"),
		middleware.LogfieldRetention:      "technical",
		middleware.LogfieldType:           "service",
	})

	loggingLevel := viper.GetString("logger.level")
	setLoggingLevel(loggingLevel)

	return withFields
}

type Cors gin.HandlerFunc

func createCors(log logrus.FieldLogger) Cors {
	allowedOrigins := viper.GetStringSlice("allowed_origins")

	log.WithFields(logrus.Fields{
		"allowed-origins": &allowedOrigins,
	}).Info("creating cors middleware")
	return Cors(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Encoding", "Content-Type"},
		AllowCredentials: true,
		AllowWildcard:    true,
		MaxAge:           12 * time.Hour,
	}))
}
