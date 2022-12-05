package service

import (
	model "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/grafana"
	"github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/provider"
)

type IUserInfoService interface {
	GetUserInfo(userID int64) (*model.UserInfo, error)
}

type UserInfoService struct {
	UserInfoProvider provider.IUserInfoProvider
}

func (u *UserInfoService) GetUserInfo(userID int64) (*model.UserInfo, error) {
	return u.UserInfoProvider.GetUserInfo(userID)
}
