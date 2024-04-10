// Code generated by MockGen. DO NOT EDIT.
// Source: journey_store.go
//
// Generated by this command:
//
//	mockgen -destination=../mocks/mock_journey_store.go -package=mocks -source=journey_store.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	errors "github.com/thevibegod/fsm/errors"
	model "github.com/thevibegod/fsm/model"
	gomock "go.uber.org/mock/gomock"
)

// MockJourneyStore is a mock of JourneyStore interface.
type MockJourneyStore[T any] struct {
	ctrl     *gomock.Controller
	recorder *MockJourneyStoreMockRecorder[T]
}

// MockJourneyStoreMockRecorder is the mock recorder for MockJourneyStore.
type MockJourneyStoreMockRecorder[T any] struct {
	mock *MockJourneyStore[T]
}

// NewMockJourneyStore creates a new mock instance.
func NewMockJourneyStore[T any](ctrl *gomock.Controller) *MockJourneyStore[T] {
	mock := &MockJourneyStore[T]{ctrl: ctrl}
	mock.recorder = &MockJourneyStoreMockRecorder[T]{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJourneyStore[T]) EXPECT() *MockJourneyStoreMockRecorder[T] {
	return m.recorder
}

// Create mocks base method.
func (m *MockJourneyStore[T]) Create(ctx context.Context) (model.Journey[T], *errors.FsmError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx)
	ret0, _ := ret[0].(model.Journey[T])
	ret1, _ := ret[1].(*errors.FsmError)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockJourneyStoreMockRecorder[T]) Create(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockJourneyStore[T])(nil).Create), ctx)
}

// Delete mocks base method.
func (m *MockJourneyStore[T]) Delete(ctx context.Context, jID string) *errors.FsmError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, jID)
	ret0, _ := ret[0].(*errors.FsmError)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockJourneyStoreMockRecorder[T]) Delete(ctx, jID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockJourneyStore[T])(nil).Delete), ctx, jID)
}

// Get mocks base method.
func (m *MockJourneyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], *errors.FsmError) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, jID)
	ret0, _ := ret[0].(model.Journey[T])
	ret1, _ := ret[1].(*errors.FsmError)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockJourneyStoreMockRecorder[T]) Get(ctx, jID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockJourneyStore[T])(nil).Get), ctx, jID)
}

// Save mocks base method.
func (m *MockJourneyStore[T]) Save(ctx context.Context, journey model.Journey[T]) *errors.FsmError {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, journey)
	ret0, _ := ret[0].(*errors.FsmError)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockJourneyStoreMockRecorder[T]) Save(ctx, journey any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockJourneyStore[T])(nil).Save), ctx, journey)
}
