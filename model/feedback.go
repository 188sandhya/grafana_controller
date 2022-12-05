package model

import (
	"time"
)

const TeamName string = "Dev Performance Team"
const FolderName string = "Team's dashboards"

type Feedback struct {
	ID              int64     `json:"id" db:"id" binding:"-" validate:"required_for_update"`
	ReceivingUserID int64     `json:"receivingUserId" db:"receiving_user_id" binding:"-" validate:"required"`
	GivingUserID    int64     `json:"givingUserId" db:"giving_user_id" binding:"-" validate:"required"`
	OrgID           int64     `json:"orgId" db:"org_id" binding:"-" validate:"required"`
	FeedbackDate    time.Time `json:"feedbackDate" db:"feedback_date" binding:"-" validate:"required"`
}
