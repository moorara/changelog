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

	CommitsMock struct {
		OutCommits git.Commits
		OutError   error
	}

	CommitMock struct {
		InHash    string
		OutCommit git.Commit
		OutError  error
	}

	TagsMock struct {
		OutTags  git.Tags
		OutError error
	}

	TagMock struct {
		InName   string
		OutTag   git.Tag
		OutError error
	}

	CommitsFromRevisionMock struct {
		InRev      string
		OutCommits git.Commits
		OutError   error
	}

	MockGitRepo struct {
		GetRemoteInfoIndex int
		GetRemoteInfoMocks []GetRemoteInfoMock

		CommitsIndex int
		CommitsMocks []CommitsMock

		CommitIndex int
		CommitMocks []CommitMock

		TagsIndex int
		TagsMocks []TagsMock

		TagIndex int
		TagMocks []TagMock

		CommitsFromRevisionIndex int
		CommitsFromRevisionMocks []CommitsFromRevisionMock
	}
)

func (m *MockGitRepo) GetRemoteInfo() (string, string, error) {
	i := m.GetRemoteInfoIndex
	m.GetRemoteInfoIndex++
	return m.GetRemoteInfoMocks[i].OutDomain, m.GetRemoteInfoMocks[i].OutPath, m.GetRemoteInfoMocks[i].OutError
}

func (m *MockGitRepo) Commits() (git.Commits, error) {
	i := m.CommitsIndex
	m.CommitsIndex++
	return m.CommitsMocks[i].OutCommits, m.CommitsMocks[i].OutError
}

func (m *MockGitRepo) Commit(hash string) (git.Commit, error) {
	i := m.CommitIndex
	m.CommitIndex++
	m.CommitMocks[i].InHash = hash
	return m.CommitMocks[i].OutCommit, m.CommitMocks[i].OutError
}

func (m *MockGitRepo) Tags() (git.Tags, error) {
	i := m.TagsIndex
	m.TagsIndex++
	return m.TagsMocks[i].OutTags, m.TagsMocks[i].OutError
}

func (m *MockGitRepo) Tag(name string) (git.Tag, error) {
	i := m.TagIndex
	m.TagIndex++
	m.TagMocks[i].InName = name
	return m.TagMocks[i].OutTag, m.TagMocks[i].OutError
}

func (m *MockGitRepo) CommitsFromRevision(rev string) (git.Commits, error) {
	i := m.CommitsFromRevisionIndex
	m.CommitsFromRevisionIndex++
	m.CommitsFromRevisionMocks[i].InRev = rev
	return m.CommitsFromRevisionMocks[i].OutCommits, m.CommitsFromRevisionMocks[i].OutError
}

type (
	FetchIssuesAndMergesMock struct {
		InContext context.Context
		InSince   time.Time
		OutIssues remote.Changes
		OutMerges remote.Changes
		OutError  error
	}

	FetchTagsMock struct {
		InContext context.Context
		OutTags   remote.Tags
		OutError  error
	}

	MockRemoteRepo struct {
		FetchTagsIndex int
		FetchTagsMocks []FetchTagsMock

		FetchIssuesAndMergesIndex int
		FetchIssuesAndMergesMocks []FetchIssuesAndMergesMock
	}
)

func (m *MockRemoteRepo) FetchTags(ctx context.Context) (remote.Tags, error) {
	i := m.FetchTagsIndex
	m.FetchTagsIndex++
	m.FetchTagsMocks[i].InContext = ctx
	return m.FetchTagsMocks[i].OutTags, m.FetchTagsMocks[i].OutError
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
