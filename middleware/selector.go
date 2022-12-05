package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/validator"
)

type IDSelector func(c *gin.Context) (int64, error)

type OrgID struct {
	ID int64 `json:"orgId" binding:"required"`
}

func OrgIDInStruct(c *gin.Context) (int64, error) {
	var id OrgID

	if err := c.ShouldBindBodyWith(&id, binding.JSON); err != nil {
		return 0, errory.ParseErrors.Builder().Wrap(err).WithMessage("Required parameter Org ID is missing").Create()
	}
	return id.ID, validator.ValidateID(id.ID)
}

type SloID struct {
	ID int64 `json:"sloId" binding:"required"`
}

func SloIDInStruct(c *gin.Context) (int64, error) {
	var id SloID

	if err := c.ShouldBindBodyWith(&id, binding.JSON); err != nil {
		return 0, errory.ParseErrors.Builder().Wrap(err).WithMessage("Required parameter SLO ID is missing").Create()
	}
	return id.ID, validator.ValidateID(id.ID)
}
