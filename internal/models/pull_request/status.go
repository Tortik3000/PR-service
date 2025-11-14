package pull_request

type PRStatus int

const (
	MERGED PRStatus = iota
	OPEN
)
