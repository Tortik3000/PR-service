package dto

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func FromAPIMembers(members []api.TeamMember) []models.Member {
	ret := make([]models.Member, len(members))
	for i, member := range members {
		ret[i] = models.Member{
			IsActive: member.IsActive,
			UserID:   member.UserId,
			Username: member.Username,
		}
	}
	return ret
}

func ToAPIMembers(members []models.Member) []api.TeamMember {
	ret := make([]api.TeamMember, len(members))
	for i, member := range members {
		ret[i] = api.TeamMember{
			IsActive: member.IsActive,
			UserId:   member.UserID,
			Username: member.Username,
		}
	}
	return ret
}

func FromAPITeam(team *api.Team) *models.Team {
	return &models.Team{
		Name:    team.TeamName,
		Members: FromAPIMembers(team.Members),
	}
}

func ToAPITeam(team *models.Team) *api.Team {
	members := make([]api.TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = api.TeamMember{
			IsActive: m.IsActive,
			UserId:   m.UserID,
			Username: m.Username,
		}
	}
	return &api.Team{
		TeamName: team.Name,
		Members:  members,
	}
}
