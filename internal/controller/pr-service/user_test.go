package pr_service

import (
	"context"
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

func TestGetUsersGetReview(t *testing.T) {
	t.Parallel()

	prs := []models.PRShort{
		{
			ID:       "pr1",
			Name:     "Fix bug",
			AuthorID: "u1",
			Status:   models.PRStatusOPEN,
		},
		{
			ID:       "pr2",
			Name:     "New feature",
			AuthorID: "u3",
			Status:   models.PRStatusOPEN,
		},
	}

	tests := []struct {
		name         string
		params       api.GetUsersGetReviewParams
		mockBehavior func(m *mocks.MockuserUseCase)
		expected     api.GetUsersGetReviewResponseObject
		wantErr      error
	}{
		{
			name: "success 200",
			params: api.GetUsersGetReviewParams{
				UserId: "u2",
			},
			mockBehavior: func(m *mocks.MockuserUseCase) {
				m.EXPECT().
					GetReview(gomock.Any(), "u2").
					Return(prs, nil)
			},
			expected: api.GetUsersGetReview200JSONResponse{
				UserId:       "u2",
				PullRequests: dto.ToAPIShortSlice(prs),
			},
			wantErr: nil,
		},
		{
			name:   "unexpected 500",
			params: api.GetUsersGetReviewParams{},
			mockBehavior: func(m *mocks.MockuserUseCase) {
				m.EXPECT().
					GetReview(gomock.Any(), gomock.Any()).
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

			mockUser := mocks.NewMockuserUseCase(ctrl)
			tt.mockBehavior(mockUser)

			svc := NewPRService(
				zap.NewNop(),
				mockUser,
				nil,
				nil,
			)

			resp, err := svc.GetUsersGetReview(context.Background(),
				api.GetUsersGetReviewRequestObject{
					Params: tt.params,
				})

			if tt.wantErr == nil {
				require.Nil(t, err)
				assert.Equal(t, tt.expected, resp)
				return
			}

			require.Error(t, err)
			require.Equal(t, tt.wantErr.Error(), err.Error())
		})
	}
}

func TestPostUsersSetIsActive(t *testing.T) {
	t.Parallel()

	user := &models.User{
		ID:       "u1",
		Name:     "John",
		IsActive: true,
	}

	tests := []struct {
		name         string
		body         *api.PostUsersSetIsActiveJSONRequestBody
		mockBehavior func(m *mocks.MockuserUseCase)
		expected     api.PostUsersSetIsActiveResponseObject
		wantErr      error
	}{
		{
			name: "success 200",
			body: &api.PostUsersSetIsActiveJSONRequestBody{
				UserId:   "u1",
				IsActive: true,
			},
			mockBehavior: func(m *mocks.MockuserUseCase) {
				m.EXPECT().
					SetIsActive(gomock.Any(), "u1", true).
					Return(user, nil)
			},
			expected: api.PostUsersSetIsActive200JSONResponse{
				User: dto.ToAPIUser(user),
			},
			wantErr: nil,
		},
		{
			name: "user not found 404",
			body: &api.PostUsersSetIsActiveJSONRequestBody{},
			mockBehavior: func(m *mocks.MockuserUseCase) {
				m.EXPECT().
					SetIsActive(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, modelsErr.ErrUserNotFound)
			},
			expected: api.PostUsersSetIsActive404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, modelsErr.ErrUserNotFound.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "unexpected error 500",
			body: &api.PostUsersSetIsActiveJSONRequestBody{},
			mockBehavior: func(m *mocks.MockuserUseCase) {
				m.EXPECT().
					SetIsActive(gomock.Any(), gomock.Any(), gomock.Any()).
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

			mockUser := mocks.NewMockuserUseCase(ctrl)
			tt.mockBehavior(mockUser)

			svc := NewPRService(
				zap.NewNop(),
				mockUser,
				nil,
				nil,
			)

			resp, err := svc.PostUsersSetIsActive(context.Background(),
				api.PostUsersSetIsActiveRequestObject{
					Body: tt.body,
				})

			if tt.wantErr == nil {
				require.Nil(t, err)
			} else {
				require.ErrorIs(t, tt.wantErr, err)
			}

			assert.Equal(t, tt.expected, resp)
		})
	}
}
