// Code generated by MockGen. DO NOT EDIT.
// Source: ../view.go
//
// Generated by this command:
//
//	mockgen -source=../view.go -destination=./view.go
//

// Package mock_domain is a generated GoMock package.
package mock_domain

import (
	reflect "reflect"

	entity "github.com/canpok1/ai-feed/internal/domain/entity"
	gomock "go.uber.org/mock/gomock"
)

// MockViewer is a mock of Viewer interface.
type MockViewer struct {
	ctrl     *gomock.Controller
	recorder *MockViewerMockRecorder
	isgomock struct{}
}

// MockViewerMockRecorder is the mock recorder for MockViewer.
type MockViewerMockRecorder struct {
	mock *MockViewer
}

// NewMockViewer creates a new mock instance.
func NewMockViewer(ctrl *gomock.Controller) *MockViewer {
	mock := &MockViewer{ctrl: ctrl}
	mock.recorder = &MockViewerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockViewer) EXPECT() *MockViewerMockRecorder {
	return m.recorder
}

// ViewArticles mocks base method.
func (m *MockViewer) ViewArticles(arg0 []entity.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ViewArticles", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ViewArticles indicates an expected call of ViewArticles.
func (mr *MockViewerMockRecorder) ViewArticles(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ViewArticles", reflect.TypeOf((*MockViewer)(nil).ViewArticles), arg0)
}

// ViewRecommend mocks base method.
func (m *MockViewer) ViewRecommend(arg0 *entity.Recommend) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ViewRecommend", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ViewRecommend indicates an expected call of ViewRecommend.
func (mr *MockViewerMockRecorder) ViewRecommend(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ViewRecommend", reflect.TypeOf((*MockViewer)(nil).ViewRecommend), arg0)
}
