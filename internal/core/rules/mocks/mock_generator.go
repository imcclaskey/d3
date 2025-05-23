// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/imcclaskey/d3/internal/core/rules (interfaces: Generator)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockGenerator is a mock of Generator interface.
type MockGenerator struct {
	ctrl     *gomock.Controller
	recorder *MockGeneratorMockRecorder
}

// MockGeneratorMockRecorder is the mock recorder for MockGenerator.
type MockGeneratorMockRecorder struct {
	mock *MockGenerator
}

// NewMockGenerator creates a new mock instance.
func NewMockGenerator(ctrl *gomock.Controller) *MockGenerator {
	mock := &MockGenerator{ctrl: ctrl}
	mock.recorder = &MockGeneratorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGenerator) EXPECT() *MockGeneratorMockRecorder {
	return m.recorder
}

// GenerateCoreContent mocks base method.
func (m *MockGenerator) GenerateCoreContent(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateCoreContent", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateCoreContent indicates an expected call of GenerateCoreContent.
func (mr *MockGeneratorMockRecorder) GenerateCoreContent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateCoreContent", reflect.TypeOf((*MockGenerator)(nil).GenerateCoreContent), arg0, arg1)
}

// GeneratePhaseContent mocks base method.
func (m *MockGenerator) GeneratePhaseContent(arg0, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneratePhaseContent", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GeneratePhaseContent indicates an expected call of GeneratePhaseContent.
func (mr *MockGeneratorMockRecorder) GeneratePhaseContent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneratePhaseContent", reflect.TypeOf((*MockGenerator)(nil).GeneratePhaseContent), arg0, arg1)
}

// GeneratePrefix mocks base method.
func (m *MockGenerator) GeneratePrefix(arg0, arg1 string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneratePrefix", arg0, arg1)
	ret0, _ := ret[0].(string)
	return ret0
}

// GeneratePrefix indicates an expected call of GeneratePrefix.
func (mr *MockGeneratorMockRecorder) GeneratePrefix(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneratePrefix", reflect.TypeOf((*MockGenerator)(nil).GeneratePrefix), arg0, arg1)
}
