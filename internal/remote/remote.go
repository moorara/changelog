package remote

import (
	"context"
	"time"
)

// Repo is the abstraction for a remote repository.
type Repo interface {
	FetchClosedIssuesAndMerges(context.Context, time.Time) ([]Change, []Change, error)
}
