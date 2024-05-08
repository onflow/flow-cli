// Code generated by mockery v2.40.3. DO NOT EDIT.

package migrate

import mock "github.com/stretchr/testify/mock"

// mockStagingValidator is an autogenerated mock type for the stagingValidator type
type mockStagingValidator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: stagedContracts
func (_m *mockStagingValidator) Validate(stagedContracts []StagedContract) error {
	ret := _m.Called(stagedContracts)

	if len(ret) == 0 {
		panic("no return value specified for Validate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]StagedContract) error); ok {
		r0 = rf(stagedContracts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// newMockStagingValidator creates a new instance of mockStagingValidator. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockStagingValidator(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockStagingValidator {
	mock := &mockStagingValidator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
