package model

import (
	"time"
)

type ExternalSloType string

const (
	ExternalSloTypeNoType  ExternalSloType = ""
	ExternalSloTypeMetric  ExternalSloType = "metric"
	ExternalSloTypeMonitor ExternalSloType = "monitor"
)

//nolint:lll
type Slo struct {
	ID                              int64           `db:"id" json:"id" binding:"-" validate:"required_for_update"`
	OrgID                           int64           `db:"org_id" json:"orgId" binding:"-" validate:"required"`
	Name                            string          `db:"name" json:"name" binding:"-" validate:"required,max=188"`
	SuccessRateExpectedAvailability string          `db:"success_rate_exp_availability" json:"successRateExpAvailability" binding:"-" validate:"required,float_string,gte_val=0,lte_val=100,max_precision=3"`
	ComplianceExpectedAvailability  string          `db:"compliance_exp_availability" json:"complianceExpAvailability" binding:"-" validate:"required,float_string,gte_val=0,lte_val=100,max_precision=3"`
	Autogen                         bool            `db:"autogen" json:"autogen"`
	CreationDate                    time.Time       `db:"creation_date" json:"creationDate" binding:"-"`
	Critical                        bool            `db:"critical" json:"critical" binding:"-" validate:"required_with=ExternalSLA"`
	DatasourceID                    int64           `db:"ds_id" json:"datasourceId" binding:"-" validate:"gte=0"`
	ExternalID                      string          `db:"external_id" json:"externalId" binding:"-" validate:"omitempty,max=32"`
	ExternalSLA                     string          `db:"external_sla" json:"externalSla" binding:"-" validate:"omitempty,float_string,gte_val=0,lte_val=100,max_precision=3"`
	ExternalType                    ExternalSloType `db:"external_type" json:"externalType" binding:"-" validate:"omitempty,oneof=metric monitor"`
}

type DetailedSlo struct {
	Slo
	SolutionID *int64 `db:"solution_id" json:"solutionId" binding:"-"`
	OrgName    string `db:"org_name" json:"orgName" binding:"-"`
}

type MetricType string
type DatasourceType string

const (
	MetricTypeSuccessRate  MetricType = "successrate"
	MetricTypeCompliance   MetricType = "compliance"
	MetricTypeNoMetricType MetricType = ""
)

type SloQueryParams struct {
	OrgID          int64          `json:"-" validate:"required"`
	MetricType     MetricType     `json:"metricType" validate:"omitempty,oneof=successrate compliance"`
	DatasourceType DatasourceType `json:"datasourceType" validate:"omitempty,oneof=elasticsearch prometheus datadog"`
}
