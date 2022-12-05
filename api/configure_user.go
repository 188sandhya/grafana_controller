package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/errory"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/service"
	"github.com/sirupsen/logrus"
)

type ConfigureUserAPI struct {
	UserInfoService service.IUserInfoService
	Log             logrus.FieldLogger
}

func (api *ConfigureUserAPI) ConfigureUser(c *gin.Context) {
	userContext, err := GetUserContext(c)

	if err != nil {
		setErrorResponse(c, errory.OnGetErrors.Builder().Wrap(err).WithMessage("Cannot get user configuration").Create(), api.Log)
		return
	}

	userInfo, err := api.UserInfoService.GetUserInfo(userContext.ID)
	if err != nil {
		setErrorResponse(c, err, api.Log)
		return
	}

	userInfo.Cookie = userContext.Cookie
	c.JSON(http.StatusOK, userInfo)
}
