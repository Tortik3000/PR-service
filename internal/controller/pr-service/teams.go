package pr_service

import (
	"context"
	"errors"

	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	models "github.com/Tortik3000/PR-service/internal/models/team"
	"go.uber.org/zap"
)

func (p *prService) PostTeamAdd(
	ctx context.Context,
	request generated.PostTeamAddRequestObject,
) (generated.PostTeamAddResponseObject, error) {
	body := request.Body

	team := &models.Team{
		TeamName: body.TeamName,
		Members:  models.FromAPIMembers(body.Members),
	}
	err := p.teamUseCase.TeamAdd(ctx, team)
	p.logger.Error("err", zap.Error(err))
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamExist):
			return generated.PostTeamAdd400JSONResponse{
				Error: newErrorResponse(generated.TEAMEXISTS, err.Error()).Error,
			}, nil

		default:
			return nil, err
		}
	}

	return generated.PostTeamAdd201JSONResponse{
		Team: models.ToAPITeam(team),
	}, nil
}

func (p *prService) GetTeamGet(
	ctx context.Context,
	request generated.GetTeamGetRequestObject,
) (generated.GetTeamGetResponseObject, error) {
	team, err := p.teamUseCase.TeamGet(ctx, request.Params.TeamName)

	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamNotFound):
			return generated.GetTeamGet404JSONResponse{
				Error: newErrorResponse(generated.NOTFOUND, err.Error()).Error,
			}, nil

		default:
			return nil, err
		}
	}

	return generated.GetTeamGet200JSONResponse{
		TeamName: team.TeamName,
		Members:  models.ToAPIMembers(team.Members),
	}, nil
}
