package generate

import (
	"context"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/remote"
)

type (
	GetRemoteInfoMock struct {
		OutDomain string
		OutPath   string
		OutError  error
	}

	TagsMock struct {
		OutTags  git.Tags
		OutError error
	}

	CommitMock struct {
		InHash    string
		OutCommit git.Commit
		OutError  error
	}

	MockGitRepo struct {
		GetRemoteInfoIndex int
		GetRemoteInfoMocks []GetRemoteInfoMock

		TagsIndex int
		TagsMocks []TagsMock

		CommitIndex int
		CommitMocks []CommitMock
	}
)

func (m *MockGitRepo) GetRemoteInfo() (string, string, error) {
	i := m.GetRemoteInfoIndex
	m.GetRemoteInfoIndex++
	return m.GetRemoteInfoMocks[i].OutDomain, m.GetRemoteInfoMocks[i].OutPath, m.GetRemoteInfoMocks[i].OutError
}

func (m *MockGitRepo) Tags() (git.Tags, error) {
	i := m.TagsIndex
	m.TagsIndex++
	return m.TagsMocks[i].OutTags, m.TagsMocks[i].OutError
}

func (m *MockGitRepo) Commit(hash string) (git.Commit, error) {
	i := m.CommitIndex
	m.CommitIndex++
	return m.CommitMocks[i].OutCommit, m.CommitMocks[i].OutError
}

type (
	FetchIssuesAndMergesMock struct {
		InContext context.Context
		InSince   time.Time
		OutIssues remote.Changes
		OutMerges remote.Changes
		OutError  error
	}

	FetchAllTagsMock struct {
		InContext context.Context
		OutTags   remote.Tags
		OutError  error
	}

	MockRemoteRepo struct {
		FetchAllTagsIndex int
		FetchAllTagsMocks []FetchAllTagsMock

		FetchIssuesAndMergesIndex int
		FetchIssuesAndMergesMocks []FetchIssuesAndMergesMock
	}
)

func (m *MockRemoteRepo) FetchAllTags(ctx context.Context) (remote.Tags, error) {
	i := m.FetchAllTagsIndex
	m.FetchAllTagsIndex++
	m.FetchAllTagsMocks[i].InContext = ctx
	return m.FetchAllTagsMocks[i].OutTags, m.FetchAllTagsMocks[i].OutError
}

func (m *MockRemoteRepo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Changes, remote.Changes, error) {
	i := m.FetchIssuesAndMergesIndex
	m.FetchIssuesAndMergesIndex++
	m.FetchIssuesAndMergesMocks[i].InContext = ctx
	m.FetchIssuesAndMergesMocks[i].InSince = since
	return m.FetchIssuesAndMergesMocks[i].OutIssues, m.FetchIssuesAndMergesMocks[i].OutMerges, m.FetchIssuesAndMergesMocks[i].OutError
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
