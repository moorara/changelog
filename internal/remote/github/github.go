package github

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/pkg/log"
	"github.com/moorara/go-github"
)

const pageSize = 100

type (
	githubService interface {
		EnsureScopes(context.Context, ...github.Scope) error
	}

	usersService interface {
		Get(context.Context, string) (*github.User, *github.Response, error)
	}

	repoService interface {
		Get(context.Context) (*github.Repository, *github.Response, error)
		Commit(context.Context, string) (*github.Commit, *github.Response, error)
		Commits(context.Context, int, int) ([]github.Commit, *github.Response, error)
		Branch(context.Context, string) (*github.Branch, *github.Response, error)
		Tags(context.Context, int, int) ([]github.Tag, *github.Response, error)
		Issues(context.Context, int, int, github.IssuesParams) ([]github.Issue, *github.Response, error)
		Events(context.Context, int, int, int) ([]github.Event, *github.Response, error)
	}
)

// repo implements the remote.Repo interface for GitHub.
type repo struct {
	logger log.Logger
	owner  string
	repo   string
	stores struct {
		users   *store
		commits *store
	}
	services struct {
		github githubService
		users  usersService
		repo   repoService
	}
}

// NewRepo creates a new GitHub repository.
func NewRepo(logger log.Logger, repoOwner, repoName, accessToken string) remote.Repo {
	client := github.NewClient(accessToken)

	r := &repo{
		logger: logger,
		owner:  repoOwner,
		repo:   repoName,
	}

	r.stores.users = newStore()
	r.stores.commits = newStore()
	r.services.github = client
	r.services.users = client.Users
	r.services.repo = client.Repo(repoOwner, repoName)

	return r
}

func (r *repo) getUser(ctx context.Context, username string) (github.User, error) {
	// First, check the cache
	if v, ok := r.stores.users.Load(username); ok {
		u := v.(github.User)
		return u, nil
	}

	u, _, err := r.services.users.Get(ctx, username)
	if err != nil {
		return github.User{}, err
	}

	// Update the cache
	r.stores.users.Save(u.Login, *u)

	return *u, nil
}

func (r *repo) getCommit(ctx context.Context, ref string) (github.Commit, error) {
	// First, check the cache
	if v, ok := r.stores.commits.Load(ref); ok {
		c := v.(github.Commit)
		return c, nil
	}

	c, _, err := r.services.repo.Commit(ctx, ref)
	if err != nil {
		return github.Commit{}, err
	}

	// Update the cache
	r.stores.commits.Save(c.SHA, *c)

	return *c, nil
}

func (r *repo) getParentCommits(ctx context.Context, ref string) (remote.Commits, error) {
	commits := remote.Commits{}

	c, err := r.getCommit(ctx, ref)
	if err != nil {
		return nil, err
	}

	commits = append(commits, toCommit(c))

	for _, parent := range c.Parents {
		parentCommits, err := r.getParentCommits(ctx, parent.SHA)
		if err != nil {
			return nil, err
		}

		commits = append(commits, parentCommits...)
	}

	return commits, nil
}

func (r *repo) findEvent(ctx context.Context, num int, name string) (github.Event, error) {
	for p := 1; p > 0; {
		events, resp, err := r.services.repo.Events(ctx, num, pageSize, p)
		if err != nil {
			return github.Event{}, err
		}

		for _, e := range events {
			if e.Event == name {
				r.logger.Debugf("Found %s event for issue %d", name, num)
				return e, nil
			}
		}

		// resp.Pages.Next == 0 is not a valid page number and causes the loop to exit
		p = resp.Pages.Next
	}

	return github.Event{}, nil
}

// FutureTag returns a tag that does not exist yet for a GitHub repository.
func (r *repo) FutureTag(name string) remote.Tag {
	return remote.Tag{
		Name:   name,
		Time:   time.Now(),
		WebURL: fmt.Sprintf("https://github.com/%s/%s/tree/%s", r.owner, r.repo, name),
	}
}

// CompareURL returns a URL for comparing two revisions for a GitHub repository.
func (r *repo) CompareURL(base, head string) string {
	return fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s", r.owner, r.repo, base, head)
}

// CheckPermissions ensures the client has all the required permissions for a GitHub repository.
func (r *repo) CheckPermissions(ctx context.Context) error {
	err := r.services.github.EnsureScopes(ctx, github.ScopeRepo)
	if err != nil {
		return err
	}

	r.logger.Debugf("GitHub token scopes verified: %s", github.ScopeRepo)

	return nil
}

// FetchFirstCommit retrieves the firist/initial commit for a GitHub repository.
func (r *repo) FetchFirstCommit(ctx context.Context) (remote.Commit, error) {
	r.logger.Debug("Fetching the first GitHub commit ...")

	var c github.Commit

	for p := 1; p > 0; {
		commits, resp, err := r.services.repo.Commits(ctx, pageSize, p)
		if err != nil {
			return remote.Commit{}, err
		}

		// Add commits to commit store
		for _, c := range commits {
			r.stores.commits.Save(c.SHA, c)
		}

		if l := len(commits); l > 0 {
			c = commits[l-1]
		}

		// Fetch the last page if there more pages
		// resp.Pages.Last == 0 is not a valid page number and causes the loop to exit
		p = resp.Pages.Last
	}

	commit := toCommit(c)

	r.logger.Debugf("Fetched the first GitHub commit: %s", commit)

	return commit, nil
}

// FetchBranch retrieves a branch by name for a GitHub repository.
func (r *repo) FetchBranch(ctx context.Context, name string) (remote.Branch, error) {
	b, _, err := r.services.repo.Branch(ctx, name)
	if err != nil {
		return remote.Branch{}, err
	}

	branch := toBranch(*b)

	r.logger.Debugf("Fetched GitHub branch: %s", name)

	return branch, nil
}

// FetchDefaultBranch retrieves the default branch for a GitHub repository.
func (r *repo) FetchDefaultBranch(ctx context.Context) (remote.Branch, error) {
	rp, _, err := r.services.repo.Get(ctx)
	if err != nil {
		return remote.Branch{}, err
	}

	b, _, err := r.services.repo.Branch(ctx, rp.DefaultBranch)
	if err != nil {
		return remote.Branch{}, err
	}

	branch := toBranch(*b)

	r.logger.Debugf("Fetched GitHub default branch: %s", b.Name)

	return branch, nil
}

// FetchTags retrieves all tags for a GitHub repository.
func (r *repo) FetchTags(ctx context.Context) (remote.Tags, error) {
	r.logger.Debug("Fetching GitHub tags ...")

	// ==============================> FETCH TAGS <==============================

	tagStore := newStore()

	// Fetch tags
	r.logger.Debug("Fetched GitHub tags page 1 ...")
	gitHubTags, resp, err := r.services.repo.Tags(ctx, pageSize, 1)
	if err != nil {
		return nil, err
	}
	for _, t := range gitHubTags {
		tagStore.Save(t.Name, t)
	}

	g1, ctx1 := errgroup.WithContext(ctx)

	// Fetch more tags if any
	for p := 2; p <= resp.Pages.Last; p++ {
		p := p // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			r.logger.Debugf("Fetched GitHub tags page %d ...", p)
			gitHubTags, _, err := r.services.repo.Tags(ctx1, pageSize, p)
			if err != nil {
				return err
			}
			for _, t := range gitHubTags {
				tagStore.Save(t.Name, t)
			}
			return nil
		})
	}

	if err := g1.Wait(); err != nil {
		return nil, err
	}

	// ==============================> FETCH TAG COMMITS <==============================

	r.logger.Debug("Fetching GitHub commits for tags ...")

	g2, ctx2 := errgroup.WithContext(ctx)

	// Fetch commits for tags
	_ = tagStore.ForEach(func(_, v interface{}) error {
		g2.Go(func() error {
			tag := v.(github.Tag)
			_, err := r.getCommit(ctx2, tag.Commit.SHA)
			return err
		})
		return nil
	})

	if err := g2.Wait(); err != nil {
		return nil, err
	}

	// ==============================> JOINING TAGS & COMMITS <==============================

	tags := resolveTags(tagStore, r.stores.commits, r.owner, r.repo)

	r.logger.Debugf("GitHub tags are fetched: %s", tags.Map(func(t remote.Tag) string {
		return t.Name
	}))

	return tags, nil
}

// FetchIssuesAndMerges retrieves all closed issues and merged pull requests for a GitHub repository.
func (r *repo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Issues, remote.Merges, error) {
	if since.IsZero() {
		r.logger.Info("Fetching GitHub issues since the beginning ...")
	} else {
		r.logger.Infof("Fetching GitHub issues since %s ...", since.Format(time.RFC3339))
	}

	// ==============================> FETCH ISSUES <==============================

	issueStore := newStore()
	opts := github.IssuesParams{
		State: "closed",
		Since: since,
	}

	// Fetch closed issues
	r.logger.Debug("Fetched GitHub issues page 1 ...")
	gitHubIssues, resp, err := r.services.repo.Issues(ctx, pageSize, 1, opts)
	if err != nil {
		return nil, nil, err
	}
	for _, i := range gitHubIssues {
		issueStore.Save(i.Number, i)
	}

	g1, ctx1 := errgroup.WithContext(ctx)

	// Fetch more closed issues if any
	for p := 2; p <= resp.Pages.Last; p++ {
		p := p // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			r.logger.Debugf("Fetched GitHub issues page %d ...", p)
			gitHubIssues, _, err := r.services.repo.Issues(ctx1, pageSize, p, opts)
			if err != nil {
				return err
			}
			for _, i := range gitHubIssues {
				issueStore.Save(i.Number, i)
			}
			return nil
		})
	}

	if err := g1.Wait(); err != nil {
		return nil, nil, err
	}

	r.logger.Debugf("Fetched GitHub issues: %d", issueStore.Len())

	// ==============================> FETCH EVENTS & COMMITS <==============================

	r.logger.Debug("Fetching GitHub events and commits for issues and pull requests ...")

	eventStore := newStore()

	g2, ctx2 := errgroup.WithContext(ctx)

	// Fetch and search events
	_ = issueStore.ForEach(func(k, v interface{}) error {
		num := k.(int)
		issue := v.(github.Issue)

		g2.Go(func() error {
			// Issue
			if issue.PullURLs == nil {
				e, err := r.findEvent(ctx2, num, "closed")
				if err != nil {
					return err
				}
				eventStore.Save(num, e)
				return nil
			}

			// Pull Request
			e, err := r.findEvent(ctx2, num, "merged")
			if err != nil {
				return err
			}

			// Ensure the eevnt is not empty/zero
			// If it is empty/zero, the desired event has not been found
			if e.CommitID != "" {
				eventStore.Save(num, e)
				if _, err := r.getCommit(ctx2, e.CommitID); err != nil {
					return err
				}
			}

			return nil
		})

		return nil
	})

	if err := g2.Wait(); err != nil {
		return nil, nil, err
	}

	// ==============================> FETCH USERS <==============================

	r.logger.Debug("Fetching GitHub users for issues and pull requests ...")

	// Fetch author users for issues and pull requests
	err = issueStore.ForEach(func(k, v interface{}) error {
		num := k.(int)
		issue := v.(github.Issue)

		// Only fetch the user of the issue is closed or the pull request is merged
		if _, ok := eventStore.Load(num); ok {
			_, err := r.getUser(ctx, issue.User.Login)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Fetch closer/merger users
	err = eventStore.ForEach(func(k, v interface{}) error {
		e := v.(github.Event)
		_, err := r.getUser(ctx, e.Actor.Login)
		return err
	})

	if err != nil {
		return nil, nil, err
	}

	// ==============================> JOINING ISSUES, PULLS, EVENTS, COMMITS, & USERS <==============================

	issues, merges := resolveIssuesAndMerges(issueStore, eventStore, r.stores.commits, r.stores.users)

	r.logger.Debugf("Resolved and sorted GitHub issues (%d) and pull requests (%d)", len(issues), len(merges))
	r.logger.Infof("All GitHub issues (%d) and pull requests (%d) are fetched", len(issues), len(merges))

	return issues, merges, nil
}

// FetchParentCommits retrieves all parent commits of a given commit hash for a GitHub repository.
func (r *repo) FetchParentCommits(ctx context.Context, ref string) (remote.Commits, error) {
	r.logger.Debugf("Fetching all GitHub parent commits for %s ...", ref)

	commits, err := r.getParentCommits(ctx, ref)
	if err != nil {
		return nil, err
	}

	r.logger.Debugf("All GitHub parent commits for %s are fetched", ref)

	return commits, nil
}
