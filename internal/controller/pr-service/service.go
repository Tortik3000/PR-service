package pr_service

import (
	"go.uber.org/zap"
)

type prService struct {
	logger             *zap.Logger
	userUseCase        userUseCase
	teamUseCase        teamUseCase
	pullRequestUseCase pullRequestUseCase
}

func NewPRService(
	logger *zap.Logger,
	userUseCase userUseCase,
	teamUseCase teamUseCase,
	pullRequestUseCase pullRequestUseCase) *prService {
	return &prService{
		logger:             logger,
		userUseCase:        userUseCase,
		teamUseCase:        teamUseCase,
		pullRequestUseCase: pullRequestUseCase,
	}
}
