// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/auth (interfaces: IAuthenticator)

// Package auth is a generated GoMock package.
package auth

import (
	reflect "reflect"

	gin "github.com/gin-gonic/gin"
	gomock "github.com/golang/mock/gomock"
	auth "github.com/metro-digital-inner-source/errorbudget-grafana-controller/grafana-controller/model/auth"
)

// MockIAuthenticator is a mock of IAuthenticator interface.
type MockIAuthenticator struct {
	ctrl     *gomock.Controller
	recorder *MockIAuthenticatorMockRecorder
}

// MockIAuthenticatorMockRecorder is the mock recorder for MockIAuthenticator.
type MockIAuthenticatorMockRecorder struct {
	mock *MockIAuthenticator
}

// NewMockIAuthenticator creates a new mock instance.
func NewMockIAuthenticator(ctrl *gomock.Controller) *MockIAuthenticator {
	mock := &MockIAuthenticator{ctrl: ctrl}
	mock.recorder = &MockIAuthenticatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAuthenticator) EXPECT() *MockIAuthenticatorMockRecorder {
	return m.recorder
}

// Authenticate mocks base method.
func (m *MockIAuthenticator) Authenticate(arg0 *gin.Context) (*auth.UserContext, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authenticate", arg0)
	ret0, _ := ret[0].(*auth.UserContext)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Authenticate indicates an expected call of Authenticate.
func (mr *MockIAuthenticatorMockRecorder) Authenticate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authenticate", reflect.TypeOf((*MockIAuthenticator)(nil).Authenticate), arg0)
}
