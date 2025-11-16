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

func TestPostPullRequestCreate(t *testing.T) {
	t.Parallel()

	pr := &models.PR{
		AuthorID:          "user1",
		ID:                "pr123",
		Name:              "My PR",
		Status:            models.PRStatusOPEN,
		AssignedReviewers: []string{"u2", "u3"},
	}
	tests := []struct {
		name         string
		body         *api.PostPullRequestCreateJSONRequestBody
		mockBehavior func(m *mocks.MockpullRequestUseCase)
		expected     api.PostPullRequestCreateResponseObject
		wantErr      error
	}{
		{
			name: "success 201",
			body: &api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        "user1",
				PullRequestId:   "pr123",
				PullRequestName: "My PR",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "user1", "pr123", "My PR").
					Return(pr, nil)
			},
			expected: api.PostPullRequestCreate201JSONResponse{
				Pr: dto.ToAPIPullRequest(pr),
			},
			wantErr: nil,
		},
		{
			name: "team not found → 404",
			body: &api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        "userX",
				PullRequestId:   "pr404",
				PullRequestName: "PR 404",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "userX", "pr404", "PR 404").
					Return(nil, modelsErr.ErrTeamNotFound)
			},
			expected: api.PostPullRequestCreate404JSONResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: modelsErr.ErrTeamNotFound.Error(),
				},
			},
			wantErr: nil,
		},
		{
			name: "PR exists → 409",
			body: &api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        "user1",
				PullRequestId:   "duplicated",
				PullRequestName: "DUP",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "user1", "duplicated", "DUP").
					Return(nil, modelsErr.ErrPullRequestExist)
			},
			expected: api.PostPullRequestCreate409JSONResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.PREXISTS,
					Message: modelsErr.ErrPullRequestExist.Error(),
				},
			},
			wantErr: nil,
		},
		{
			name: "unexpected error → 500 (err returned)",
			body: &api.PostPullRequestCreateJSONRequestBody{
				AuthorId:        "user1",
				PullRequestId:   "some",
				PullRequestName: "Some",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestCreate(gomock.Any(), "user1", "some", "Some").
					Return(nil, errors.New("db crash"))
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

			mockPR := mocks.NewMockpullRequestUseCase(ctrl)
			tt.mockBehavior(mockPR)

			svc := NewPRService(
				zap.NewNop(),
				nil,
				nil,
				mockPR,
			)

			resp, err := svc.PostPullRequestCreate(t.Context(), api.PostPullRequestCreateRequestObject{
				Body: tt.body,
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

func TestPostPullRequestMerge(t *testing.T) {
	t.Parallel()

	mergedPR := &models.PR{
		ID:                "pr1",
		Name:              "Test PR",
		AuthorID:          "u1",
		Status:            models.PRStatusMERGED,
		AssignedReviewers: []string{"u2", "u3"},
	}

	tests := []struct {
		name         string
		body         *api.PostPullRequestMergeJSONRequestBody
		mockBehavior func(m *mocks.MockpullRequestUseCase)
		expected     api.PostPullRequestMergeResponseObject
		wantErr      error
	}{
		{
			name: "success 200",
			body: &api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "pr1",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "pr1").
					Return(mergedPR, nil)
			},
			expected: api.PostPullRequestMerge200JSONResponse{
				Pr: dto.ToAPIPullRequest(mergedPR),
			},
			wantErr: nil,
		},
		{
			name: "PR not found → 404",
			body: &api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "not_found",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "not_found").
					Return(nil, modelsErr.ErrPRNotFound)
			},
			expected: api.PostPullRequestMerge404JSONResponse{
				Error: struct {
					Code    api.ErrorResponseErrorCode `json:"code"`
					Message string                     `json:"message"`
				}{
					Code:    api.NOTFOUND,
					Message: "pull request not found",
				},
			},
			wantErr: nil,
		},
		{
			name: "unexpected error → 500",
			body: &api.PostPullRequestMergeJSONRequestBody{
				PullRequestId: "prX",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestMerge(gomock.Any(), "prX").
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

			mockPR := mocks.NewMockpullRequestUseCase(ctrl)
			tt.mockBehavior(mockPR)

			svc := NewPRService(
				zap.NewNop(),
				nil,
				nil,
				mockPR,
			)

			resp, err := svc.PostPullRequestMerge(t.Context(),
				api.PostPullRequestMergeRequestObject{
					Body: tt.body,
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

func TestPostPullRequestReassign(t *testing.T) {
	t.Parallel()

	pr := &models.PR{
		ID:                "pr1",
		Name:              "PR name",
		AuthorID:          "u1",
		Status:            models.PRStatusOPEN,
		AssignedReviewers: []string{"u2", "u3"},
	}

	tests := []struct {
		name         string
		body         *api.PostPullRequestReassignJSONRequestBody
		mockBehavior func(m *mocks.MockpullRequestUseCase)
		expected     api.PostPullRequestReassignResponseObject
		wantErr      error
	}{
		{
			name: "success 200",
			body: &api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: pr.ID,
				OldUserId:     "u2",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), pr.ID, "u2").
					Return(pr, "NEWUSER", nil)
			},
			expected: api.PostPullRequestReassign200JSONResponse{
				Pr:         *dto.ToAPIPullRequest(pr),
				ReplacedBy: "NEWUSER",
			},
			wantErr: nil,
		},
		{
			name: "PR not found 404",
			body: &api.PostPullRequestReassignJSONRequestBody{},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, "", modelsErr.ErrPRNotFound)
			},
			expected: api.PostPullRequestReassign404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, modelsErr.ErrPRNotFound.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "user not found 404",
			body: &api.PostPullRequestReassignJSONRequestBody{},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, "", modelsErr.ErrUserNotFound)
			},
			expected: api.PostPullRequestReassign404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, modelsErr.ErrUserNotFound.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "PR merged 409",
			body: &api.PostPullRequestReassignJSONRequestBody{},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, "", modelsErr.ErrPRMerged)
			},
			expected: api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.PRMERGED, modelsErr.ErrPRMerged.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "not assigned 409",
			body: &api.PostPullRequestReassignJSONRequestBody{},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, "", modelsErr.ErrNotAssigned)
			},
			expected: api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.NOTASSIGNED, modelsErr.ErrNotAssigned.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "not active candidate 409",
			body: &api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: pr.ID,
				OldUserId:     "uX",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, "", modelsErr.ErrNotActiveCandidate)
			},
			expected: api.PostPullRequestReassign409JSONResponse{
				Error: newErrorResponse(api.NOCANDIDATE, modelsErr.ErrNotActiveCandidate.Error()).Error,
			},
			wantErr: nil,
		},
		{
			name: "unexpected error 500",
			body: &api.PostPullRequestReassignJSONRequestBody{
				PullRequestId: "err",
				OldUserId:     "err",
			},
			mockBehavior: func(m *mocks.MockpullRequestUseCase) {
				m.EXPECT().
					PullRequestReassign(gomock.Any(), "err", "err").
					Return(nil, "", errors.New("db fail"))
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

			mockPR := mocks.NewMockpullRequestUseCase(ctrl)
			tt.mockBehavior(mockPR)

			svc := NewPRService(
				zap.NewNop(),
				nil,
				nil,
				mockPR,
			)

			resp, err := svc.PostPullRequestReassign(t.Context(),
				api.PostPullRequestReassignRequestObject{
					Body: tt.body,
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
