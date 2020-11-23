package github

import (
	"fmt"
	"time"

	"github.com/moorara/go-github"

	"github.com/moorara/changelog/internal/remote"
)

func toUser(u github.User) remote.User {
	return remote.User{
		Name:     u.Name,
		Email:    u.Email,
		Username: u.Login,
		WebURL:   u.HTMLURL,
	}
}

func toCommit(c github.Commit) remote.Commit {
	return remote.Commit{
		Hash: c.SHA,
		Time: c.Commit.Committer.Time,
	}
}

func toBranch(b github.Branch) remote.Branch {
	return remote.Branch{
		Name:   b.Name,
		Commit: toCommit(b.Commit),
	}
}

func toTag(t github.Tag, c github.Commit, owner, repo string) remote.Tag {
	return remote.Tag{
		Name:   t.Name,
		Time:   c.Commit.Committer.Time,
		Commit: toCommit(c),
		WebURL: fmt.Sprintf("https://github.com/%s/%s/tree/%s", owner, repo, t.Name),
	}
}

func toIssue(i github.Issue, e github.Event, author, closer github.User) remote.Issue {
	labels := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if i.Milestone != nil {
		milestone = i.Milestone.Title
	}

	// *i.ClosedAt and e.CreatedAt are the same times
	var time time.Time
	if i.ClosedAt != nil {
		time = *i.ClosedAt
	}

	return remote.Issue{
		Change: remote.Change{
			Number:    i.Number,
			Title:     i.Title,
			Labels:    labels,
			Milestone: milestone,
			Time:      time,
			Author:    toUser(author),
			WebURL:    i.HTMLURL,
		},
		Closer: toUser(closer),
	}
}

func toMerge(i github.Issue, e github.Event, c github.Commit, author, merger github.User) remote.Merge {
	labels := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if i.Milestone != nil {
		milestone = i.Milestone.Title
	}

	// p.MergedAt and e.CreatedAt are the same times
	// c.Commit.Committer.Time is the actual time of merge
	time := c.Commit.Committer.Time

	return remote.Merge{
		Change: remote.Change{
			Number:    i.Number,
			Title:     i.Title,
			Labels:    labels,
			Milestone: milestone,
			Time:      time,
			Author:    toUser(author),
			WebURL:    i.HTMLURL,
		},
		Merger: toUser(merger),
		Commit: toCommit(c),
	}
}

func resolveTags(gitHubTags, gitHubCommits *store, owner, repo string) remote.Tags {
	tags := remote.Tags{}

	_ = gitHubTags.ForEach(func(k, v interface{}) error {
		t := v.(github.Tag)

		if v, ok := gitHubCommits.Load(t.Commit.SHA); ok {
			c := v.(github.Commit)
			tags = append(tags, toTag(t, c, owner, repo))
		}

		return nil
	})

	return tags
}

func resolveIssuesAndMerges(gitHubIssues, gitHubEvents, gitHubCommits, gitHubUsers *store) (remote.Issues, remote.Merges) {
	issues := remote.Issues{}
	merges := remote.Merges{}

	_ = gitHubIssues.ForEach(func(k, v interface{}) error {
		num := k.(int)
		i := v.(github.Issue)

		if i.PullURLs == nil { // Issue
			v, _ := gitHubEvents.Load(num)
			e := v.(github.Event)

			v, _ = gitHubUsers.Load(i.User.Login)
			author := v.(github.User)

			v, _ = gitHubUsers.Load(e.Actor.Login)
			closer := v.(github.User)

			issues = append(issues, toIssue(i, e, author, closer))
		} else { // Pull request
			// If no event found, the pull request is closed without being merged
			if v, ok := gitHubEvents.Load(num); ok {
				e := v.(github.Event)

				v, _ = gitHubCommits.Load(e.CommitID)
				c := v.(github.Commit)

				v, _ = gitHubUsers.Load(i.User.Login)
				author := v.(github.User)

				v, _ = gitHubUsers.Load(e.Actor.Login)
				merger := v.(github.User)

				merges = append(merges, toMerge(i, e, c, author, merger))
			}
		}

		return nil
	})

	issues = issues.Sort()
	merges = merges.Sort()

	return issues, merges
}
