package pr_service

import (
	"context"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (u *useCase) TeamAdd(
	ctx context.Context,
	team models.Team,
) error {
	err := u.teamRepository.TeamAdd(ctx, team)
	if err != nil {
		return err
	}

	return nil
}

func (u *useCase) TeamGet(
	ctx context.Context,
	teamName string,
) (*models.Team, error) {
	team, err := u.teamRepository.TeamGet(ctx, teamName)
	if err != nil {
		return nil, err
	}

	if len(team.Members) == 0 {
		return nil, modelsErr.ErrTeamNotFound
	}

	return team, nil
}
