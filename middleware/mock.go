// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/middleware (interfaces: IAuthorizer)

// Package middleware is a generated GoMock package.
package middleware

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	auth "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
)

// MockIAuthorizer is a mock of IAuthorizer interface.
type MockIAuthorizer struct {
	ctrl     *gomock.Controller
	recorder *MockIAuthorizerMockRecorder
}

// MockIAuthorizerMockRecorder is the mock recorder for MockIAuthorizer.
type MockIAuthorizerMockRecorder struct {
	mock *MockIAuthorizer
}

// NewMockIAuthorizer creates a new mock instance.
func NewMockIAuthorizer(ctrl *gomock.Controller) *MockIAuthorizer {
	mock := &MockIAuthorizer{ctrl: ctrl}
	mock.recorder = &MockIAuthorizerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAuthorizer) EXPECT() *MockIAuthorizerMockRecorder {
	return m.recorder
}

// AuthorizeForDatasource mocks base method.
func (m *MockIAuthorizer) AuthorizeForDatasource(arg0, arg1 int64, arg2 auth.Permission) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeForDatasource", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthorizeForDatasource indicates an expected call of AuthorizeForDatasource.
func (mr *MockIAuthorizerMockRecorder) AuthorizeForDatasource(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeForDatasource", reflect.TypeOf((*MockIAuthorizer)(nil).AuthorizeForDatasource), arg0, arg1, arg2)
}

// AuthorizeForHappinessMetric mocks base method.
func (m *MockIAuthorizer) AuthorizeForHappinessMetric(arg0, arg1 int64, arg2 auth.Permission) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeForHappinessMetric", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthorizeForHappinessMetric indicates an expected call of AuthorizeForHappinessMetric.
func (mr *MockIAuthorizerMockRecorder) AuthorizeForHappinessMetric(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeForHappinessMetric", reflect.TypeOf((*MockIAuthorizer)(nil).AuthorizeForHappinessMetric), arg0, arg1, arg2)
}

// AuthorizeForOrganization mocks base method.
func (m *MockIAuthorizer) AuthorizeForOrganization(arg0, arg1 int64, arg2 auth.Permission) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeForOrganization", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthorizeForOrganization indicates an expected call of AuthorizeForOrganization.
func (mr *MockIAuthorizerMockRecorder) AuthorizeForOrganization(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeForOrganization", reflect.TypeOf((*MockIAuthorizer)(nil).AuthorizeForOrganization), arg0, arg1, arg2)
}

// AuthorizeForSLO mocks base method.
func (m *MockIAuthorizer) AuthorizeForSLO(arg0, arg1 int64, arg2 auth.Permission) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeForSLO", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthorizeForSLO indicates an expected call of AuthorizeForSLO.
func (mr *MockIAuthorizerMockRecorder) AuthorizeForSLO(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeForSLO", reflect.TypeOf((*MockIAuthorizer)(nil).AuthorizeForSLO), arg0, arg1, arg2)
}

// AuthorizeForTeam mocks base method.
func (m *MockIAuthorizer) AuthorizeForTeam(arg0, arg1 int64, arg2 auth.Permission) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthorizeForTeam", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthorizeForTeam indicates an expected call of AuthorizeForTeam.
func (mr *MockIAuthorizerMockRecorder) AuthorizeForTeam(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthorizeForTeam", reflect.TypeOf((*MockIAuthorizer)(nil).AuthorizeForTeam), arg0, arg1, arg2)
}
