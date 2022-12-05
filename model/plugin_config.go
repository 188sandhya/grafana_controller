package model

type PluginConfig struct {
	Datasources map[string]string `json:"datasources"`
	PluginName  string            `json:"pluginName"`
	SDA         *SDAConfig        `json:"sda"`
}

type SDAConfig struct {
	ElasticTemplate   string               `json:"elasticTemplate"`
	ElasticTimeFields map[IndexType]string `json:"elasticTimeFields"`
}

type IndexType string

type SDAFeatures struct {
	Git       bool `db:"git"`
	CICD      bool `db:"cicd"`
	Jira      bool `db:"jira"`
	Sonarqube bool `db:"sonarqube"`
}
