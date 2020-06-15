// Code generated by MockGen. DO NOT EDIT.
// Source: markdown.go

// Package mock_markdown is a generated GoMock package.
package mock_markdown

import (
	gomock "github.com/golang/mock/gomock"
	io "io"
	reflect "reflect"
)

// MockConverter is a mock of Converter interface
type MockConverter struct {
	ctrl     *gomock.Controller
	recorder *MockConverterMockRecorder
}

// MockConverterMockRecorder is the mock recorder for MockConverter
type MockConverterMockRecorder struct {
	mock *MockConverter
}

// NewMockConverter creates a new mock instance
func NewMockConverter(ctrl *gomock.Controller) *MockConverter {
	mock := &MockConverter{ctrl: ctrl}
	mock.recorder = &MockConverterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockConverter) EXPECT() *MockConverterMockRecorder {
	return m.recorder
}

// Convert mocks base method
func (m *MockConverter) Convert(src []byte, w io.Writer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Convert", src, w)
	ret0, _ := ret[0].(error)
	return ret0
}

// Convert indicates an expected call of Convert
func (mr *MockConverterMockRecorder) Convert(src, w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Convert", reflect.TypeOf((*MockConverter)(nil).Convert), src, w)
}
