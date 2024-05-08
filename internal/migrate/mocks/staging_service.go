// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import (
	context "context"

	common "github.com/onflow/cadence/runtime/common"

	mock "github.com/stretchr/testify/mock"

	project "github.com/onflow/flowkit/v2/project"
)

// StagingService is an autogenerated mock type for the stagingService type
type StagingService struct {
	mock.Mock
}

// StageContracts provides a mock function with given fields: ctx, filter
func (_m *StagingService) StageContracts(ctx context.Context, filter func(*project.Contract) bool) (map[common.AddressLocation]error, error) {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for StageContracts")
	}

	var r0 map[common.AddressLocation]error
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, func(*project.Contract) bool) (map[common.AddressLocation]error, error)); ok {
		return rf(ctx, filter)
	}
	if rf, ok := ret.Get(0).(func(context.Context, func(*project.Contract) bool) map[common.AddressLocation]error); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[common.AddressLocation]error)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, func(*project.Contract) bool) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewStagingService creates a new instance of StagingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStagingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *StagingService {
	mock := &StagingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
