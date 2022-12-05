package model

import (
	"time"
)

type HappinessMetric struct {
	ID            int64     `json:"id" db:"id" binding:"-" validate:"required_for_update"`
	UserID        int64     `json:"userId" db:"user_id" binding:"-"`
	OrgID         int64     `json:"orgId" db:"org_id" binding:"-" validate:"required"`
	Happiness     float64   `json:"happiness" db:"happiness" binding:"-" validate:"required,gte_val=0,lte_val=10"`
	Safety        float64   `json:"safety" db:"safety"  binding:"-" validate:"required,gte_val=0,lte_val=10"`
	SafetyOutlier int       `json:"safetyOutlier" db:"safety_min_outlier" binding:"-"`
	Date          time.Time `json:"date" db:"date" binding:"-" validate:"required"`
	Enabled       bool      `json:"enabled" db:"enabled" binding:"-"`
}

type UserMissingInput struct {
	UserID int64  `json:"userId" db:"user_id" binding:"-"`
	Login  string `json:"login" db:"login" binding:"-"`
}
