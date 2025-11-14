package pull_request

type PRStatus int

const (
	OPEN PRStatus = iota
	MERGED
)
