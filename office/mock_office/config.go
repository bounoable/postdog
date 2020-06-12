// Code generated by MockGen. DO NOT EDIT.
// Source: config.go

// Package mock_office is a generated GoMock package.
package mock_office

import (
	context "context"
	letter "github.com/bounoable/postdog/letter"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockMiddleware is a mock of Middleware interface
type MockMiddleware struct {
	ctrl     *gomock.Controller
	recorder *MockMiddlewareMockRecorder
}

// MockMiddlewareMockRecorder is the mock recorder for MockMiddleware
type MockMiddlewareMockRecorder struct {
	mock *MockMiddleware
}

// NewMockMiddleware creates a new mock instance
func NewMockMiddleware(ctrl *gomock.Controller) *MockMiddleware {
	mock := &MockMiddleware{ctrl: ctrl}
	mock.recorder = &MockMiddlewareMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMiddleware) EXPECT() *MockMiddlewareMockRecorder {
	return m.recorder
}

// Handle mocks base method
func (m *MockMiddleware) Handle(ctx context.Context, let letter.Letter) (letter.Letter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Handle", ctx, let)
	ret0, _ := ret[0].(letter.Letter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Handle indicates an expected call of Handle
func (mr *MockMiddlewareMockRecorder) Handle(ctx, let interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Handle", reflect.TypeOf((*MockMiddleware)(nil).Handle), ctx, let)
}

// MockLogger is a mock of Logger interface
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Log mocks base method
func (m *MockLogger) Log(v ...interface{}) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range v {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Log", varargs...)
}

// Log indicates an expected call of Log
func (mr *MockLoggerMockRecorder) Log(v ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockLogger)(nil).Log), v...)
}
