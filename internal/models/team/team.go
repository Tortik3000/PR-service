package team

type Team struct {
	Members  []Member
	TeamName string
}

type Member struct {
	IsActive bool
	UserID   string
	Username string
}
