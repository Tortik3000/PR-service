package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/models"
)

func TestToAPIPullRequest(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name     string
		input    *models.PR
		expected *api.PullRequest
	}{
		{
			name: "merged PR",
			input: &models.PR{
				ID:                "pr1",
				Name:              "Fix bug",
				AuthorID:          "user1",
				Status:            models.PRStatusMERGED,
				AssignedReviewers: []string{"rev1", "rev2"},
				CreatedAt:         &now,
				MergedAt:          &now,
			},
			expected: &api.PullRequest{
				PullRequestId:     "pr1",
				PullRequestName:   "Fix bug",
				AuthorId:          "user1",
				Status:            api.PullRequestStatusMERGED,
				AssignedReviewers: []string{"rev1", "rev2"},
				CreatedAt:         &now,
				MergedAt:          &now,
			},
		},
		{
			name: "open PR",
			input: &models.PR{
				Status:    models.PRStatusOPEN,
				CreatedAt: &now,
				MergedAt:  nil,
			},
			expected: &api.PullRequest{
				Status:    api.PullRequestStatusOPEN,
				CreatedAt: &now,
				MergedAt:  nil,
			},
		},
		{
			name:     "nil pr",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToAPIPullRequest(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToAPIPullRequestShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *models.PRShort
		expected *api.PullRequestShort
	}{
		{
			name: "merged PR short",
			input: &models.PRShort{
				ID:       "pr1",
				Name:     "Fix bug",
				AuthorID: "user1",
				Status:   models.PRStatusMERGED,
			},
			expected: &api.PullRequestShort{
				PullRequestId:   "pr1",
				PullRequestName: "Fix bug",
				AuthorId:        "user1",
				Status:          api.PullRequestShortStatus(api.PullRequestStatusMERGED),
			},
		},
		{
			name: "open PR short",
			input: &models.PRShort{
				Status: models.PRStatusOPEN,
			},
			expected: &api.PullRequestShort{
				Status: api.PullRequestShortStatus(api.PullRequestStatusOPEN),
			},
		},
		{
			name:     "nil pr",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ToAPIPullRequestShort(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToAPIShortSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []models.PRShort
		expected []api.PullRequestShort
	}{
		{
			name: "multiple PRs",
			input: []models.PRShort{
				{ID: "pr1", Name: "name1", AuthorID: "user1", Status: models.PRStatusMERGED},
				{Status: models.PRStatusOPEN},
			},
			expected: []api.PullRequestShort{
				{PullRequestId: "pr1", PullRequestName: "name1", AuthorId: "user1", Status: api.PullRequestShortStatus(api.PullRequestStatusMERGED)},
				{Status: api.PullRequestShortStatus(api.PullRequestStatusOPEN)},
			},
		},
		{
			name:     "empty slice",
			input:    []models.PRShort{},
			expected: []api.PullRequestShort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ToAPIShortSlice(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
