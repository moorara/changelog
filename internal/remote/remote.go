package remote

import (
	"context"
	"time"
)

// Repo is the abstraction for a remote repository.
type Repo interface {
	// FetchAllTags retrieves all tags.
	FetchAllTags(context.Context) (Tags, error)
	// FetchIssuesAndMerges retrieves closed issues and merged pull/merge requests.
	FetchIssuesAndMerges(context.Context, time.Time) (Changes, Changes, error)
}
