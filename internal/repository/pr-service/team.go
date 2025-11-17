package pr_service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
)

func (p *postgresRepo) TeamAdd(
	ctx context.Context,
	team models.Team,
) (txErr error) {
	logger := p.logger.With(
		zap.String("team_name", team.Name),
		zap.Any("members", team.Members),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		logger.Error("beginTx", zap.Error(err))
		return err
	}
	defer func() {
		rollback(txErr)
	}()

	createTeam := p.queryBuilder.Insert("team").
		Columns("name").
		Values(team.Name).
		Suffix("RETURNING id")

	createTeamStr, args, err := createTeam.ToSql()
	if err != nil {
		logger.Error("build SQL (create team)", zap.Error(err))
		return err
	}

	logger.Debug("Executing create team SQL",
		zap.String("query", createTeamStr),
		zap.Any("args", args),
	)

	var teamID int
	err = tx.QueryRow(ctx, createTeamStr, args...).Scan(&teamID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			logger.Warn("team already exists", zap.Error(err))
			return modelsErr.ErrTeamExist
		}
		logger.Error("create team query", zap.Error(err))
		return err
	}

	createUsers := p.queryBuilder.Insert("users").
		Columns("id", "name", "is_active", "team_id")

	for _, m := range team.Members {
		createUsers = createUsers.Values(
			m.UserID,
			m.Username,
			m.IsActive,
			teamID,
		)
	}

	createUsers = createUsers.Suffix(`
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
			is_active = EXCLUDED.is_active,
			team_id = EXCLUDED.team_id
	`)

	createUsersStr, args, err := createUsers.ToSql()
	if err != nil {
		logger.Error("build SQL (insert users)", zap.Error(err))
		return err
	}

	logger.Debug("Executing insert users SQL",
		zap.String("query", createUsersStr),
		zap.Any("args", args),
	)

	_, err = tx.Exec(ctx, createUsersStr, args...)
	if err != nil {
		logger.Error("insert users", zap.Error(err))
		return err
	}

	return nil
}

func (p *postgresRepo) TeamGet(
	ctx context.Context,
	teamName string,
) (*models.Team, error) {
	logger := p.logger.With(zap.String("team_name", teamName))

	query := p.queryBuilder.Select(
		"t.id as team_id",
		"t.name as team_name",
		"u.id as user_id",
		"u.name as username",
		"u.is_active",
	).
		From("team t").
		Join("users u ON t.id = u.team_id").
		Where(sq.Eq{"t.name": teamName})

	queryStr, args, err := query.ToSql()
	if err != nil {
		logger.Error("build SQL (get team)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing get team SQL",
		zap.String("query", queryStr),
		zap.Any("args", args),
	)

	rows, err := p.db.Query(ctx, queryStr, args...)
	if err != nil {
		logger.Error("get team query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var team models.Team
	team.Members = make([]models.Member, 0)

	for rows.Next() {
		var member models.Member
		var teamID int

		err = rows.Scan(
			&teamID,
			&team.Name,
			&member.UserID,
			&member.Username,
			&member.IsActive,
		)
		if err != nil {
			logger.Error("scan team row", zap.Error(err))
			return nil, err
		}

		team.Members = append(team.Members, member)
	}

	return &team, nil
}

func (p *postgresRepo) GetActiveTeammates(
	ctx context.Context,
	teamID string,
	excludedUsers []string,
	count uint64,
) (teammateIDs []string, txErr error) {
	logger := p.logger.With(
		zap.String("team_id", teamID),
		zap.Any("excluded_users", excludedUsers),
		zap.Uint64("count", count),
	)

	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		logger.Error("beginTx", zap.Error(err))
		return nil, err
	}
	defer func() {
		rollback(txErr)
	}()

	getTeammate := p.queryBuilder.Select("id").
		From("users").
		Where(
			sq.And{
				sq.Eq{"team_id": teamID},
				sq.Eq{"is_active": true},
				sq.NotEq{"id": excludedUsers},
			},
		).
		Limit(count).
		Suffix("FOR UPDATE")

	getTeammateStr, args, err := getTeammate.ToSql()
	if err != nil {
		logger.Error("build SQL (get active teammates)", zap.Error(err))
		return nil, err
	}

	logger.Debug("Executing get teammates SQL",
		zap.String("query", getTeammateStr),
		zap.Any("args", args),
	)

	rows, err := tx.Query(ctx, getTeammateStr, args...)
	if err != nil {
		logger.Error("get teammates query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
			logger.Error("scan teammate ID", zap.Error(err))
			return nil, err
		}
		teammateIDs = append(teammateIDs, userID)
	}

	return teammateIDs, nil
}

func (p *postgresRepo) GetTeamIDByUserID(
	ctx context.Context,
	userID string,
) (teamID string, err error) {
	logger := p.logger.With(zap.String("user_id", userID))

	getTeamID := p.queryBuilder.Select("team_id").
		From("users").
		Where(sq.Eq{"id": userID})

	getTeamIDStr, args, err := getTeamID.ToSql()
	if err != nil {
		logger.Error("build SQL (get team ID)", zap.Error(err))
		return "", err
	}

	logger.Debug("Executing get team ID SQL",
		zap.String("query", getTeamIDStr),
		zap.Any("args", args),
	)

	err = p.db.QueryRow(ctx, getTeamIDStr, args...).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found")
			return "", modelsErr.ErrUserNotFound
		}
		logger.Error("get team ID query", zap.Error(err))
		return "", err
	}

	if teamID == "" {
		logger.Error("team not found")
		return "", modelsErr.ErrTeamNotFound
	}

	return teamID, nil
}
