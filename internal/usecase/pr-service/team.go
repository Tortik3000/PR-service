package pr_service

import (
	"context"

	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	models "github.com/Tortik3000/PR-service/internal/models/team"
)

func (u *useCase) TeamAdd(
	ctx context.Context,
	team *models.Team,
) error {
	err := u.teamRepository.TeamAdd(ctx, models.ToDB(team))
	if err != nil {
		return err
	}

	return nil
}

func (u *useCase) TeamGet(
	ctx context.Context,
	teamName string,
) (*models.Team, error) {
	dbTeam, err := u.teamRepository.TeamGet(ctx, teamName)
	if err != nil {
		return nil, err
	}

	if len(dbTeam.Members) == 0 {
		return nil, modelsErr.ErrTeamNotFound
	}

	return models.FromDB(dbTeam), nil
}
