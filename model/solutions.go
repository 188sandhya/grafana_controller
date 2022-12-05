package model

type ProductScopeType string

const (
	ProductScopeTypeDefault         = "Unknown"
	ProductScopeTypeDifferentiating = "Differentiating"
	ProductScopeTypeElementary      = "Elementary"
)

type Solution struct {
	OrgName           string  `db:"org_name" json:"orgName"`
	ID                int64   `db:"solution_id" json:"solutionId"`
	Name              string  `db:"solution_name" json:"solutionName"`
	ProductID         *int64  `db:"product_id" json:"productId"`
	ProductName       string  `db:"product_name" json:"productName"`
	Featured          bool    `db:"featured" json:"featured"`
	ProductScope      string  `db:"product_scope" json:"productScope"`
	ServiceClass      *string `db:"service_class" json:"serviceClass"`
	ServiceClassLevel string  `db:"service_class_level" json:"serviceClassLevel"`
}
