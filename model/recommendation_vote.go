package model

import (
	"time"
)

type Vote string

const (
	Like    Vote = "like"
	Dislike Vote = "dislike"
)

type RecommendationVote struct {
	ID                 int64     `json:"id" db:"id" binding:"-" validate:"isdefault"`
	UserID             int64     `json:"userId" db:"user_id" binding:"-" validate:"isdefault"`
	OrgID              int64     `json:"orgId" db:"org_id" binding:"-" validate:"required"`
	RecommendationType string    `json:"recommendationType" db:"recommendation_type" binding:"-" validate:"required"`
	Vote               Vote      `json:"vote" db:"vote" binding:"-" validate:"required,oneof=like dislike"`
	Date               time.Time `json:"date" db:"date" binding:"-" validate:"isdefault"`
}
