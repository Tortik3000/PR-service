package team

type DBTeam struct {
	Members  []DBTeamMember
	TeamName string
}

type DBTeamMember struct {
	IsActive bool
	UserID   string
	Username string
}
