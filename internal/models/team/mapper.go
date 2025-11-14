package team

import (
	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
)

func FromDB(dbTeam *DBTeam) *Team {
	members := make([]Member, len(dbTeam.Members))
	for i, m := range dbTeam.Members {
		members[i] = Member{
			IsActive: m.IsActive,
			UserID:   m.UserID,
			Username: m.Username,
		}
	}
	return &Team{
		TeamName: dbTeam.TeamName,
		Members:  members,
	}
}

func ToDB(team *Team) *DBTeam {
	members := make([]DBTeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = DBTeamMember{
			IsActive: m.IsActive,
			UserID:   m.UserID,
			Username: m.Username,
		}
	}
	return &DBTeam{
		TeamName: team.TeamName,
		Members:  members,
	}
}

func FromAPIMembers(members []api.TeamMember) []Member {
	ret := make([]Member, len(members))
	for i, member := range members {
		ret[i] = Member{
			IsActive: member.IsActive,
			UserID:   member.UserId,
			Username: member.Username,
		}
	}
	return ret
}

func ToAPIMembers(members []Member) []api.TeamMember {
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

func FromAPITeam(team *api.Team) *Team {
	return &Team{
		TeamName: team.TeamName,
		Members:  FromAPIMembers(team.Members),
	}
}

func ToAPITeam(team *Team) *api.Team {
	members := make([]api.TeamMember, len(team.Members))
	for i, m := range team.Members {
		members[i] = api.TeamMember{
			IsActive: m.IsActive,
			UserId:   m.UserID,
			Username: m.Username,
		}
	}
	return &api.Team{
		TeamName: team.TeamName,
		Members:  members,
	}
}
