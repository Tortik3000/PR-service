package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func TestToAPIUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *models.User
		expected *api.User
	}{
		{
			name: "active user",
			input: &models.User{
				ID:       "123",
				Name:     "Alice",
				TeamName: "TeamA",
				IsActive: true,
			},
			expected: &api.User{
				UserId:   "123",
				Username: "Alice",
				TeamName: "TeamA",
				IsActive: true,
			},
		},
		{
			name: "inactive user",
			input: &models.User{
				ID:       "456",
				IsActive: false,
			},
			expected: &api.User{
				UserId:   "456",
				IsActive: false,
			},
		},
		{
			name:     "nil user",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.input == nil {
				require.Nil(t, ToAPIUser(tt.input))
				return
			}

			result := ToAPIUser(tt.input)
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result)
		})
	}
}
