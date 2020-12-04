// Code generated by MockGen. DO NOT EDIT.
// Source: smtp.go

// Package mock_smtp is a generated GoMock package.
package mock_smtp

import (
	gomock "github.com/golang/mock/gomock"
	smtp "net/smtp"
	reflect "reflect"
)

// MockMailSender is a mock of MailSender interface
type MockMailSender struct {
	ctrl     *gomock.Controller
	recorder *MockMailSenderMockRecorder
}

// MockMailSenderMockRecorder is the mock recorder for MockMailSender
type MockMailSenderMockRecorder struct {
	mock *MockMailSender
}

// NewMockMailSender creates a new mock instance
func NewMockMailSender(ctrl *gomock.Controller) *MockMailSender {
	mock := &MockMailSender{ctrl: ctrl}
	mock.recorder = &MockMailSenderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMailSender) EXPECT() *MockMailSenderMockRecorder {
	return m.recorder
}

// SendMail mocks base method
func (m *MockMailSender) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMail", addr, a, from, to, msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMail indicates an expected call of SendMail
func (mr *MockMailSenderMockRecorder) SendMail(addr, a, from, to, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMail", reflect.TypeOf((*MockMailSender)(nil).SendMail), addr, a, from, to, msg)
}
