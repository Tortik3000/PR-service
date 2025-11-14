package user

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
)

func ToDB(user *User) *DBUser {
	return &DBUser{
		ID:       user.ID,
		Name:     user.Name,
		IsActive: user.IsActive,
		TeamName: user.TeamName,
	}
}

func FromDB(dbUser *DBUser) *User {
	return &User{
		ID:       dbUser.ID,
		Name:     dbUser.Name,
		IsActive: dbUser.IsActive,
		TeamName: dbUser.TeamName,
	}
}

func ToAPIUser(user *User) *api.User {
	return &api.User{
		UserId:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
