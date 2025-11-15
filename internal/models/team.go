package models

type Team struct {
	Members []Member
	Name    string
}

type Member struct {
	IsActive bool
	UserID   string
	Username string
}
