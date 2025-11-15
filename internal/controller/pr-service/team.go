package pr_service

import (
	"context"
	"errors"

	generated "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"go.uber.org/zap"
)

func (p *prService) PostTeamAdd(
	ctx context.Context,
	request generated.PostTeamAddRequestObject,
) (generated.PostTeamAddResponseObject, error) {
	body := request.Body

	team := models.Team{
		Name:    body.TeamName,
		Members: dto.FromAPIMembers(body.Members),
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
		Team: dto.ToAPITeam(&team),
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
		TeamName: team.Name,
		Members:  dto.ToAPIMembers(team.Members),
	}, nil
}
