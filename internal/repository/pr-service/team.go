package pr_service

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	modelsErr "github.com/Tortik3000/PR-service/internal/models/errors"
	"github.com/Tortik3000/PR-service/internal/models/team"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *postgresRepo) TeamAdd(
	ctx context.Context,
	team *team.DBTeam,
) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	createTeam := sq.Insert("team").
		Columns("name").
		Values(team.TeamName).
		Suffix("RETURNING id")

	createTeamStr, args, err := createTeam.PlaceholderFormat(sq.Dollar).ToSql()
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

	createUsers := sq.Insert("users").
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
	`).PlaceholderFormat(sq.Dollar)

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

func (p *postgresRepo) TeamGet(ctx context.Context, teamName string) (*team.DBTeam, error) {
	query := sq.Select(
		"t.id as team_id",
		"t.name as team_name",
		"u.id as user_id",
		"u.name as username",
		"u.is_active",
	).
		From("team t").
		Join("users u ON t.id = u.team_id").
		Where(sq.Eq{"t.name": teamName}).
		PlaceholderFormat(sq.Dollar)

	queryStr, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.db.Query(ctx, queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbTeam team.DBTeam
	dbTeam.Members = make([]team.DBTeamMember, 0)

	for rows.Next() {
		var member team.DBTeamMember
		var teamID int

		err = rows.Scan(
			&teamID,
			&dbTeam.TeamName,
			&member.UserID,
			&member.Username,
			&member.IsActive,
		)
		if err != nil {
			return nil, err
		}

		dbTeam.Members = append(dbTeam.Members, member)
	}

	return &dbTeam, nil
}
