package dto

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func ToAPIUser(user *models.User) *api.User {
	return &api.User{
		UserId:   user.ID,
		Username: user.Name,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}
