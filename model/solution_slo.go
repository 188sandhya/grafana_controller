package model

type SolutionSlo struct {
	OrgID          int64  `db:"org_id" json:"orgId" binding:"-"`
	NoCriticalSlos int64  `db:"no_critical_slo" json:"noCriticalSLOs" binding:"-"`
	SolutionName   string `db:"solution_name" json:"-" binding:"-"`
	DashboardPath  string `db:"-" json:"dashboardPath" binding:"-"`
}
