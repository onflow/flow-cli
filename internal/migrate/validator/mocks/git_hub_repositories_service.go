// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	github "github.com/google/go-github/github"

	mock "github.com/stretchr/testify/mock"
)

// GitHubRepositoriesService is an autogenerated mock type for the GitHubRepositoriesService type
type GitHubRepositoriesService struct {
	mock.Mock
}

// DownloadContents provides a mock function with given fields: ctx, owner, repo, filepath, opt
func (_m *GitHubRepositoriesService) DownloadContents(ctx context.Context, owner string, repo string, filepath string, opt *github.RepositoryContentGetOptions) (io.ReadCloser, error) {
	ret := _m.Called(ctx, owner, repo, filepath, opt)

	if len(ret) == 0 {
		panic("no return value specified for DownloadContents")
	}

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) (io.ReadCloser, error)); ok {
		return rf(ctx, owner, repo, filepath, opt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) io.ReadCloser); ok {
		r0 = rf(ctx, owner, repo, filepath, opt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) error); ok {
		r1 = rf(ctx, owner, repo, filepath, opt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContents provides a mock function with given fields: ctx, owner, repo, path, opt
func (_m *GitHubRepositoriesService) GetContents(ctx context.Context, owner string, repo string, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	ret := _m.Called(ctx, owner, repo, path, opt)

	if len(ret) == 0 {
		panic("no return value specified for GetContents")
	}

	var r0 *github.RepositoryContent
	var r1 []*github.RepositoryContent
	var r2 *github.Response
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)); ok {
		return rf(ctx, owner, repo, path, opt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) *github.RepositoryContent); ok {
		r0 = rf(ctx, owner, repo, path, opt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.RepositoryContent)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) []*github.RepositoryContent); ok {
		r1 = rf(ctx, owner, repo, path, opt)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]*github.RepositoryContent)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) *github.Response); ok {
		r2 = rf(ctx, owner, repo, path, opt)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*github.Response)
		}
	}

	if rf, ok := ret.Get(3).(func(context.Context, string, string, string, *github.RepositoryContentGetOptions) error); ok {
		r3 = rf(ctx, owner, repo, path, opt)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// NewGitHubRepositoriesService creates a new instance of GitHubRepositoriesService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGitHubRepositoriesService(t interface {
	mock.TestingT
	Cleanup(func())
}) *GitHubRepositoriesService {
	mock := &GitHubRepositoriesService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}