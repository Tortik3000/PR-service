package pr_service

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/Tortik3000/PR-service/internal/models"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *postgresRepo) TeamAdd(
	ctx context.Context,
	team models.Team,
) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	createTeam := p.queryBuilder.Insert("team").
		Columns("name").
		Values(team.Name).
		Suffix("RETURNING id")

	createTeamStr, args, err := createTeam.ToSql()
	if err != nil {
		return err
	}

	var teamID int
	err = tx.QueryRow(ctx, createTeamStr, args...).Scan(&teamID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			return modelsErr.ErrTeamExist
		}
		return err
	}

	createUsers := p.queryBuilder.Insert("users").
		Columns("id", "name", "is_active", "team_id")
	for _, member := range team.Members {
		createUsers = createUsers.
			Values(
				member.UserID,
				member.Username,
				member.IsActive,
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
		return err
	}

	_, err = tx.Exec(ctx, createUsersStr, args...)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *postgresRepo) TeamGet(
	ctx context.Context,
	teamName string,
) (*models.Team, error) {
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
		return nil, err
	}

	rows, err := p.db.Query(ctx, queryStr, args...)
	if err != nil {
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
	tx, rollback, err := p.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback(txErr)

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
		return nil, err
	}

	rows, err := tx.Query(ctx, getTeammateStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if err = rows.Scan(&userID); err != nil {
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
	getTeamID := p.queryBuilder.Select("team_id").
		From("users").
		Where(sq.Eq{"id": userID})

	getTeamIDStr, args, err := getTeamID.ToSql()
	if err != nil {
		return "", err
	}

	err = p.db.QueryRow(ctx, getTeamIDStr, args...).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", modelsErr.ErrUserNotFound
		}
		return "", err
	}
	if teamID == "" {
		return "", modelsErr.ErrTeamNotFound
	}

	return teamID, nil
}
