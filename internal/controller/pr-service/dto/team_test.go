package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func TestFromAPIMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []api.TeamMember
		expected []models.Member
	}{
		{
			name: "multiple members",
			input: []api.TeamMember{
				{UserId: "u1", Username: "Alice", IsActive: true},
				{},
			},
			expected: []models.Member{
				{UserID: "u1", Username: "Alice", IsActive: true},
				{},
			},
		},
		{
			name:     "empty slice",
			input:    []api.TeamMember{},
			expected: []models.Member{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := FromAPIMembers(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToAPIMembers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []models.Member
		expected []api.TeamMember
	}{
		{
			name: "multiple members",
			input: []models.Member{
				{UserID: "u1", Username: "Alice", IsActive: true},
				{UserID: "u2", Username: "Bob", IsActive: false},
			},
			expected: []api.TeamMember{
				{UserId: "u1", Username: "Alice", IsActive: true},
				{UserId: "u2", Username: "Bob", IsActive: false},
			},
		},
		{
			name:     "empty slice",
			input:    []models.Member{},
			expected: []api.TeamMember{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ToAPIMembers(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToAPITeam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *models.Team
		expected *api.Team
	}{
		{
			name: "team with members",
			input: &models.Team{
				Name: "TeamA",
				Members: []models.Member{
					{UserID: "u1", Username: "Alice", IsActive: true},
					{},
				},
			},
			expected: &api.Team{
				TeamName: "TeamA",
				Members: []api.TeamMember{
					{UserId: "u1", Username: "Alice", IsActive: true},
					{},
				},
			},
		},
		{
			name: "team without members",
			input: &models.Team{
				Name:    "EmptyTeam",
				Members: []models.Member{},
			},
			expected: &api.Team{
				TeamName: "EmptyTeam",
				Members:  []api.TeamMember{},
			},
		},
		{
			name:     "nil team",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToAPITeam(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
