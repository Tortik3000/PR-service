package pr_service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	api "github.com/Tortik3000/PR-service/generated/api/pr-service"
	"github.com/Tortik3000/PR-service/internal/controller/pr-service/dto"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *prService) PostTeamAdd(
	ctx context.Context,
	request api.PostTeamAddRequestObject,
) (api.PostTeamAddResponseObject, error) {
	body := request.Body
	p.logger.Info("PostTeamAdd called",
		zap.String("team_name", body.TeamName),
	)

	team := models.Team{
		Name:    body.TeamName,
		Members: dto.FromAPIMembers(body.Members),
	}
	err := p.teamUseCase.TeamAdd(ctx, team)
	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamExist):
			return api.PostTeamAdd400JSONResponse{
				Error: newErrorResponse(api.TEAMEXISTS, err.Error()).Error,
			}, nil

		default:
			return nil, modelsErr.ErrInternal
		}
	}

	p.logger.Info("PostTeamAdd success",
		zap.String("team_name", body.TeamName),
	)

	return api.PostTeamAdd201JSONResponse{
		Team: dto.ToAPITeam(&team),
	}, nil
}

func (p *prService) GetTeamGet(
	ctx context.Context,
	request api.GetTeamGetRequestObject,
) (api.GetTeamGetResponseObject, error) {
	p.logger.Info("GetTeamGet called",
		zap.String("team_name", request.Params.TeamName),
	)

	team, err := p.teamUseCase.TeamGet(ctx, request.Params.TeamName)

	if err != nil {
		switch {
		case errors.Is(err, modelsErr.ErrTeamNotFound):
			return api.GetTeamGet404JSONResponse{
				Error: newErrorResponse(api.NOTFOUND, err.Error()).Error,
			}, nil

		default:
			return nil, modelsErr.ErrInternal
		}
	}

	p.logger.Info("GetTeamGet success",
		zap.String("team_name", team.Name),
	)
	return api.GetTeamGet200JSONResponse{
		TeamName: team.Name,
		Members:  dto.ToAPIMembers(team.Members),
	}, nil
}
