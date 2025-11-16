package pr_service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"github.com/Tortik3000/PR-service/internal/usecase/pr-service/mocks"
)

func TestUseCase_PullRequestCreate(t *testing.T) {
	t.Parallel()

	inputPR := models.PR{
		ID:                "pr1",
		Name:              "name1",
		AuthorID:          "author1",
		AssignedReviewers: []string{"u1", "u2"},
	}
	tests := []struct {
		name string

		pr       models.PR
		expectPR *models.PR
		teamID   string

		getTeamErr      error
		getTeammatesErr error
		createPrErr     error
		wantErr         error
	}{
		{
			name:     "success",
			pr:       inputPR,
			expectPR: &inputPR,
			teamID:   "team",

			getTeamErr:      nil,
			getTeammatesErr: nil,
			createPrErr:     nil,
			wantErr:         nil,
		},
		{
			name:            "error in GetTeamIDByUserID",
			pr:              inputPR,
			expectPR:        nil,
			teamID:          "team",
			getTeamErr:      modelsErr.ErrInternal,
			getTeammatesErr: nil,
			createPrErr:     nil,
			wantErr:         modelsErr.ErrInternal,
		},
		{
			name:            "error in GetTeammates",
			pr:              inputPR,
			expectPR:        nil,
			teamID:          "team",
			getTeamErr:      nil,
			getTeammatesErr: modelsErr.ErrInternal,
			createPrErr:     nil,
			wantErr:         modelsErr.ErrInternal,
		},
		{
			name:            "error in createPR",
			pr:              inputPR,
			expectPR:        nil,
			teamID:          "team",
			getTeamErr:      nil,
			getTeammatesErr: nil,
			createPrErr:     modelsErr.ErrInternal,
			wantErr:         modelsErr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTransactor := mocks.NewMocktransactor(ctrl)
			mockTeamRepo := mocks.NewMockteamRepository(ctrl)
			mockPRRepo := mocks.NewMockpullRequestsRepository(ctrl)

			ctx := t.Context()

			u := &useCase{
				transactor:             mockTransactor,
				teamRepository:         mockTeamRepo,
				pullRequestsRepository: mockPRRepo,
				logger:                 zap.NewNop(),
			}

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(_ context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				},
			)

			mockTeamRepo.EXPECT().GetTeamIDByUserID(ctx, tt.pr.AuthorID).Return(tt.teamID, tt.getTeamErr)
			if tt.getTeamErr == nil {
				mockTeamRepo.EXPECT().GetActiveTeammates(ctx, tt.teamID, []string{tt.pr.AuthorID}, uint64(2)).
					Return(tt.pr.AssignedReviewers, tt.getTeammatesErr)
			}
			if tt.getTeammatesErr == nil && tt.getTeamErr == nil {
				mockPRRepo.EXPECT().PullRequestCreate(ctx, tt.pr.AuthorID, tt.pr.ID, tt.pr.Name, tt.pr.AssignedReviewers).
					Return(tt.expectPR, tt.createPrErr)
			}

			pr, err := u.PullRequestCreate(ctx, tt.pr.AuthorID, tt.pr.ID, tt.pr.Name)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, pr, tt.expectPR)
		})
	}
}

func TestUseCase_PullRequestReassign(t *testing.T) {
	t.Parallel()

	inputPR := models.PR{
		ID:                "pr1",
		Name:              "name1",
		AuthorID:          "author1",
		AssignedReviewers: []string{"u1", "u2"},
	}

	tests := []struct {
		name string

		teamID        string
		pr            models.PR
		oldReviewerID string
		expectPR      *models.PR
		expectNewID   []string

		getTeamErr      error
		getPRerr        error
		getTeammatesErr error
		reassignPRErr   error
		wantErr         error
	}{
		{
			name:          "success",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR: &models.PR{
				ID:                "pr1",
				Name:              "name1",
				AuthorID:          "author1",
				AssignedReviewers: []string{"u3", "u2"},
			},
			expectNewID: []string{"u3"},

			getTeamErr:      nil,
			getPRerr:        nil,
			getTeammatesErr: nil,
			reassignPRErr:   nil,
			wantErr:         nil,
		},
		{
			name:          "error in GetTeamIDByUserID",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR:      nil,
			expectNewID:   []string{""},

			getTeamErr:      modelsErr.ErrInternal,
			getPRerr:        nil,
			getTeammatesErr: nil,
			reassignPRErr:   nil,
			wantErr:         modelsErr.ErrInternal,
		},
		{
			name:          "error in GetPullRequest",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR:      nil,
			expectNewID:   []string{""},

			getTeamErr:      nil,
			getPRerr:        modelsErr.ErrInternal,
			getTeammatesErr: nil,
			reassignPRErr:   nil,
			wantErr:         modelsErr.ErrInternal,
		},
		{
			name:          "old reviewer not assigned",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "uX",
			expectPR:      nil,
			expectNewID:   []string{""},

			getTeamErr:      nil,
			getPRerr:        nil,
			getTeammatesErr: nil,
			reassignPRErr:   nil,
			wantErr:         modelsErr.ErrNotAssigned,
		},
		{
			name:          "err in get active teammates",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR:      nil,
			expectNewID:   []string{""},

			getTeamErr:      nil,
			getPRerr:        nil,
			getTeammatesErr: modelsErr.ErrInternal,
			reassignPRErr:   nil,
			wantErr:         modelsErr.ErrInternal,
		},
		{
			name:          "no active teammates",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR:      nil,
			expectNewID:   []string{},

			getTeamErr:      nil,
			getPRerr:        nil,
			getTeammatesErr: nil,
			reassignPRErr:   nil,
			wantErr:         modelsErr.ErrNotActiveCandidate,
		},
		{
			name:          "error in PullRequestReassign",
			teamID:        "team",
			pr:            inputPR,
			oldReviewerID: "u1",
			expectPR:      nil,
			expectNewID:   []string{""},

			getTeamErr:      nil,
			getPRerr:        nil,
			getTeammatesErr: nil,
			reassignPRErr:   modelsErr.ErrInternal,
			wantErr:         modelsErr.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTransactor := mocks.NewMocktransactor(ctrl)
			mockTeamRepo := mocks.NewMockteamRepository(ctrl)
			mockPRRepo := mocks.NewMockpullRequestsRepository(ctrl)

			ctx := t.Context()

			u := &useCase{
				transactor:             mockTransactor,
				teamRepository:         mockTeamRepo,
				pullRequestsRepository: mockPRRepo,
				logger:                 zap.NewNop(),
			}

			mockTransactor.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(
				func(_ context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				},
			)

			mockTeamRepo.EXPECT().GetTeamIDByUserID(ctx, tt.oldReviewerID).Return(tt.teamID, tt.getTeamErr)
			if tt.getTeamErr == nil {
				mockPRRepo.EXPECT().GetPullRequest(ctx, tt.pr.ID).Return(&tt.pr, tt.getPRerr)
			}
			wasReviewer := false
			if tt.getTeamErr == nil && tt.getPRerr == nil {
				for _, r := range inputPR.AssignedReviewers {
					if r == tt.oldReviewerID {
						wasReviewer = true
						break
					}
				}
				if wasReviewer {
					mockTeamRepo.EXPECT().GetActiveTeammates(ctx, "team", gomock.Any(), uint64(1)).
						Return(tt.expectNewID, tt.getTeammatesErr)
				}
			}

			if tt.getTeamErr == nil &&
				tt.getPRerr == nil &&
				tt.getTeammatesErr == nil &&
				len(tt.expectNewID) != 0 &&
				wasReviewer {
				mockPRRepo.EXPECT().PullRequestReassign(ctx, tt.pr.ID, tt.oldReviewerID, tt.expectNewID[0]).Return(tt.reassignPRErr)
			}

			pr, newReviewerID, err := u.PullRequestReassign(ctx, tt.pr.ID, tt.oldReviewerID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectPR.AssignedReviewers, pr.AssignedReviewers)
				assert.Equal(t, tt.expectNewID[0], newReviewerID)
			}
		})
	}
}
