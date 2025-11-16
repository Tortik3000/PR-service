package pr_service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/mocks"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func TestPostTeamAdd(t *testing.T) {
	t.Parallel()

	team := &models.Team{
		Name: "name",
		Members: []models.Member{
			{
				UserID: "id1",
			},
			{
				UserID: "id2",
			},
		},
	}

	tests := []struct {
		name         string
		body         *api.PostTeamAddJSONRequestBody
		mockBehavior func(m *mocks.MockteamUseCase)
		expected     api.PostTeamAddResponseObject
		wantErr      error
	}{
		{
			name: "success 201",
			body: &api.PostTeamAddJSONRequestBody{
				TeamName: team.Name,
				Members:  dto.ToAPIMembers(team.Members),
			},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamAdd(gomock.Any(), models.Team{
						Name:    team.Name,
						Members: team.Members,
					}).
					Return(nil)
			},
			expected: api.PostTeamAdd201JSONResponse{
				Team: dto.ToAPITeam(team),
			},
			wantErr: nil,
		},
		{
			name: "team exists 400",
			body: &api.PostTeamAddJSONRequestBody{},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamAdd(gomock.Any(), gomock.Any()).
					Return(modelsErr.ErrTeamExist)
			},
			expected: api.PostTeamAdd400JSONResponse{
				Error: newErrorResponse(api.TEAMEXISTS, modelsErr.ErrTeamExist.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "unexpected error 500",
			body: &api.PostTeamAddJSONRequestBody{
				TeamName: "crash",
			},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamAdd(gomock.Any(), gomock.Any()).
					Return(errors.New("db fail"))
			},
			expected: nil,
			wantErr:  modelsErr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeam := mocks.NewMockteamUseCase(ctrl)
			tt.mockBehavior(mockTeam)

			svc := NewPRService(
				zap.NewNop(),
				nil,
				mockTeam,
				nil,
			)

			resp, err := svc.PostTeamAdd(t.Context(),
				api.PostTeamAddRequestObject{Body: tt.body})

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, tt.wantErr, err)
			}

			assert.Equal(t, tt.expected, resp)
		})
	}
}

func TestGetTeamGet(t *testing.T) {
	t.Parallel()

	team := &models.Team{
		Name: "core",
		Members: []models.Member{
			{UserID: "a"},
			{UserID: "b"},
		},
	}

	tests := []struct {
		name         string
		params       api.GetTeamGetParams
		mockBehavior func(m *mocks.MockteamUseCase)
		expected     api.GetTeamGetResponseObject
		wantErr      error
	}{
		{
			name: "success 200",
			params: api.GetTeamGetParams{
				TeamName: team.Name,
			},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamGet(gomock.Any(), team.Name).
					Return(team, nil)
			},
			expected: api.GetTeamGet200JSONResponse{
				TeamName: team.Name,
				Members:  dto.ToAPIMembers(team.Members),
			},
			wantErr: nil,
		},
		{
			name:   "not found 404",
			params: api.GetTeamGetParams{},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamGet(gomock.Any(), gomock.Any()).
					Return(nil, modelsErr.ErrTeamNotFound)
			},
			expected: api.GetTeamGet404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, modelsErr.ErrTeamNotFound.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name:   "unexpected 500",
			params: api.GetTeamGetParams{},
			mockBehavior: func(m *mocks.MockteamUseCase) {
				m.EXPECT().
					TeamGet(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db fail"))
			},
			expected: nil,
			wantErr:  modelsErr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTeam := mocks.NewMockteamUseCase(ctrl)
			tt.mockBehavior(mockTeam)

			svc := NewPRService(
				zap.NewNop(),
				nil,
				mockTeam,
				nil,
			)

			resp, err := svc.GetTeamGet(t.Context(),
				api.GetTeamGetRequestObject{
					Params: tt.params,
				})

			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, tt.wantErr, err)
			}

			assert.Equal(t, tt.expected, resp)
		})
	}
}
