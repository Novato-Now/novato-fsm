// Code generated by MockGen. DO NOT EDIT.
// Source: state_handler.go
//
// Generated by this command:
//
//	mockgen -destination=../mocks/mock_state_handler.go -package=mocks -source=state_handler.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	novato_errors "github.com/Novato-Now/novato-utils/errors"
	gomock "go.uber.org/mock/gomock"
)

// MockStateHandler is a mock of StateHandler interface.
type MockStateHandler struct {
	ctrl     *gomock.Controller
	recorder *MockStateHandlerMockRecorder
}

// MockStateHandlerMockRecorder is the mock recorder for MockStateHandler.
type MockStateHandlerMockRecorder struct {
	mock *MockStateHandler
}

// NewMockStateHandler creates a new mock instance.
func NewMockStateHandler(ctrl *gomock.Controller) *MockStateHandler {
	mock := &MockStateHandler{ctrl: ctrl}
	mock.recorder = &MockStateHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStateHandler) EXPECT() *MockStateHandlerMockRecorder {
	return m.recorder
}

// Revisit mocks base method.
func (m *MockStateHandler) Revisit(ctx context.Context, jID string, journeyData any) (any, any, *novato_errors.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Revisit", ctx, jID, journeyData)
	ret0, _ := ret[0].(any)
	ret1, _ := ret[1].(any)
	ret2, _ := ret[2].(*novato_errors.Error)
	return ret0, ret1, ret2
}

// Revisit indicates an expected call of Revisit.
func (mr *MockStateHandlerMockRecorder) Revisit(ctx, jID, journeyData any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Revisit", reflect.TypeOf((*MockStateHandler)(nil).Revisit), ctx, jID, journeyData)
}

// Visit mocks base method.
func (m *MockStateHandler) Visit(ctx context.Context, jID string, journeyData, data any) (any, any, string, *novato_errors.Error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Visit", ctx, jID, journeyData, data)
	ret0, _ := ret[0].(any)
	ret1, _ := ret[1].(any)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(*novato_errors.Error)
	return ret0, ret1, ret2, ret3
}

// Visit indicates an expected call of Visit.
func (mr *MockStateHandlerMockRecorder) Visit(ctx, jID, journeyData, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Visit", reflect.TypeOf((*MockStateHandler)(nil).Visit), ctx, jID, journeyData, data)
}
