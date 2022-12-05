package model

type UsageMetric struct {
	Name  string `db:"name" json:"name"`
	Value string `db:"value" json:"value"`
}
