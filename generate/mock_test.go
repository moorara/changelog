package generate

import (
	"context"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/remote"
)

type (
	GetRemoteMock struct {
		OutDomain string
		OutPath   string
		OutError  error
	}

	MockGitRepo struct {
		GetRemoteIndex int
		GetRemoteMocks []GetRemoteMock
	}
)

func (m *MockGitRepo) GetRemote() (string, string, error) {
	i := m.GetRemoteIndex
	m.GetRemoteIndex++
	return m.GetRemoteMocks[i].OutDomain, m.GetRemoteMocks[i].OutPath, m.GetRemoteMocks[i].OutError
}

type (
	FutureTagMock struct {
		InName string
		OutTag remote.Tag
	}

	CompareURLMock struct {
		InBase    string
		InHead    string
		OutString string
	}

	CheckPermissionsMock struct {
		InContext context.Context
		OutError  error
	}

	FetchFirstCommitMock struct {
		InContext context.Context
		OutCommit remote.Commit
		OutError  error
	}

	FetchBranchMock struct {
		InContext context.Context
		InName    string
		OutBranch remote.Branch
		OutError  error
	}

	FetchDefaultBranchMock struct {
		InContext context.Context
		OutBranch remote.Branch
		OutError  error
	}

	FetchTagsMock struct {
		InContext context.Context
		OutTags   remote.Tags
		OutError  error
	}

	FetchIssuesAndMergesMock struct {
		InContext context.Context
		InSince   time.Time
		OutIssues remote.Issues
		OutMerges remote.Merges
		OutError  error
	}

	FetchParentCommitsMock struct {
		InContext  context.Context
		InHash     string
		OutCommits remote.Commits
		OutError   error
	}

	MockRemoteRepo struct {
		FutureTagIndex int
		FutureTagMocks []FutureTagMock

		CompareURLIndex int
		CompareURLMocks []CompareURLMock

		CheckPermissionsIndex int
		CheckPermissionsMocks []CheckPermissionsMock

		FetchFirstCommitIndex int
		FetchFirstCommitMocks []FetchFirstCommitMock

		FetchBranchIndex int
		FetchBranchMocks []FetchBranchMock

		FetchDefaultBranchIndex int
		FetchDefaultBranchMocks []FetchDefaultBranchMock

		FetchTagsIndex int
		FetchTagsMocks []FetchTagsMock

		FetchIssuesAndMergesIndex int
		FetchIssuesAndMergesMocks []FetchIssuesAndMergesMock

		FetchParentCommitsIndex int
		FetchParentCommitsMocks []FetchParentCommitsMock
	}
)

func (m *MockRemoteRepo) FutureTag(name string) remote.Tag {
	i := m.FutureTagIndex
	m.FutureTagIndex++
	m.FutureTagMocks[i].InName = name
	return m.FutureTagMocks[i].OutTag
}

func (m *MockRemoteRepo) CompareURL(base, head string) string {
	i := m.CompareURLIndex
	m.CompareURLIndex++
	m.CompareURLMocks[i].InBase = base
	m.CompareURLMocks[i].InHead = head
	return m.CompareURLMocks[i].OutString
}

func (m *MockRemoteRepo) CheckPermissions(ctx context.Context) error {
	i := m.CheckPermissionsIndex
	m.CheckPermissionsIndex++
	m.CheckPermissionsMocks[i].InContext = ctx
	return m.CheckPermissionsMocks[i].OutError
}

func (m *MockRemoteRepo) FetchFirstCommit(ctx context.Context) (remote.Commit, error) {
	i := m.FetchFirstCommitIndex
	m.FetchFirstCommitIndex++
	m.FetchFirstCommitMocks[i].InContext = ctx
	return m.FetchFirstCommitMocks[i].OutCommit, m.FetchFirstCommitMocks[i].OutError
}

func (m *MockRemoteRepo) FetchBranch(ctx context.Context, name string) (remote.Branch, error) {
	i := m.FetchBranchIndex
	m.FetchBranchIndex++
	m.FetchBranchMocks[i].InContext = ctx
	m.FetchBranchMocks[i].InName = name
	return m.FetchBranchMocks[i].OutBranch, m.FetchBranchMocks[i].OutError
}

func (m *MockRemoteRepo) FetchDefaultBranch(ctx context.Context) (remote.Branch, error) {
	i := m.FetchDefaultBranchIndex
	m.FetchDefaultBranchIndex++
	m.FetchDefaultBranchMocks[i].InContext = ctx
	return m.FetchDefaultBranchMocks[i].OutBranch, m.FetchDefaultBranchMocks[i].OutError
}

func (m *MockRemoteRepo) FetchTags(ctx context.Context) (remote.Tags, error) {
	i := m.FetchTagsIndex
	m.FetchTagsIndex++
	m.FetchTagsMocks[i].InContext = ctx
	return m.FetchTagsMocks[i].OutTags, m.FetchTagsMocks[i].OutError
}

func (m *MockRemoteRepo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Issues, remote.Merges, error) {
	i := m.FetchIssuesAndMergesIndex
	m.FetchIssuesAndMergesIndex++
	m.FetchIssuesAndMergesMocks[i].InContext = ctx
	m.FetchIssuesAndMergesMocks[i].InSince = since
	return m.FetchIssuesAndMergesMocks[i].OutIssues, m.FetchIssuesAndMergesMocks[i].OutMerges, m.FetchIssuesAndMergesMocks[i].OutError
}

func (m *MockRemoteRepo) FetchParentCommits(ctx context.Context, hash string) (remote.Commits, error) {
	i := m.FetchParentCommitsIndex
	m.FetchParentCommitsIndex++
	m.FetchParentCommitsMocks[i].InContext = ctx
	m.FetchParentCommitsMocks[i].InHash = hash
	return m.FetchParentCommitsMocks[i].OutCommits, m.FetchParentCommitsMocks[i].OutError
}

type (
	ParseMock struct {
		InParseOptions changelog.ParseOptions
		OutChangelog   *changelog.Changelog
		OutError       error
	}

	RenderMock struct {
		InChangelog *changelog.Changelog
		OutContent  string
		OutError    error
	}

	MockChangelogProcessor struct {
		ParseIndex int
		ParseMocks []ParseMock

		RenderIndex int
		RenderMocks []RenderMock
	}
)

func (m *MockChangelogProcessor) Parse(opts changelog.ParseOptions) (*changelog.Changelog, error) {
	i := m.ParseIndex
	m.ParseIndex++
	m.ParseMocks[i].InParseOptions = opts
	return m.ParseMocks[i].OutChangelog, m.ParseMocks[i].OutError
}

func (m *MockChangelogProcessor) Render(chlog *changelog.Changelog) (string, error) {
	i := m.RenderIndex
	m.RenderIndex++
	m.RenderMocks[i].InChangelog = chlog
	return m.RenderMocks[i].OutContent, m.RenderMocks[i].OutError
}
