package remote

// Repo is the abstraction for a remote repository.
type Repo interface {
	FetchClosedIssuesAndMerges() ([]Issue, []Merge, error)
}

// Issue represents a single issue.
type Issue struct {
}

// Merge represents a single pull/merge/change request.
type Merge struct {
}
