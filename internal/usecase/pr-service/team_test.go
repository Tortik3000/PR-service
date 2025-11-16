package pr_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"github.com/Tortik3000/PR-service/internal/usecase/pr-service/mocks"
)

func TestUseCase_TeamGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		teamName   string
		mockReturn *models.Team
		mockErr    error
		wantTeam   *models.Team
		wantErr    error
	}{
		{
			name:     "success with members",
			teamName: "team",
			mockReturn: &models.Team{
				Name: "team",
				Members: []models.Member{
					{UserID: "u1", Username: "Alice", IsActive: true},
					{},
				},
			},
			mockErr: nil,
			wantTeam: &models.Team{
				Name: "team",
				Members: []models.Member{
					{UserID: "u1", Username: "Alice", IsActive: true},
					{},
				},
			},
			wantErr: nil,
		},
		{
			name:       "team found but no members",
			teamName:   "team",
			mockReturn: &models.Team{Name: "team", Members: []models.Member{}},
			mockErr:    nil,
			wantTeam:   nil,
			wantErr:    modelsErr.ErrTeamNotFound,
		},
		{
			name:       "error from repository",
			teamName:   "team",
			mockReturn: nil,
			mockErr:    modelsErr.ErrInternal,
			wantTeam:   nil,
			wantErr:    modelsErr.ErrInternal,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeamRepo := mocks.NewMockteamRepository(ctrl)
			u := &useCase{
				teamRepository: mockTeamRepo,
			}

			mockTeamRepo.EXPECT().
				TeamGet(gomock.Any(), tt.teamName).
				Return(tt.mockReturn, tt.mockErr)

			team, err := u.TeamGet(nil, tt.teamName)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTeam, team)
			}
		})
	}
}
