package pr_service

import (
	"context"
	"database/sql"
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

	insertIntoTeam := sq.Insert("team").
		Columns("name").
		Values(team.TeamName).
		Suffix("RETURNING id")

	insertIntoTeamStr, args, err := insertIntoTeam.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	var teamID int
	err = tx.QueryRow(ctx, insertIntoTeamStr, args...).Scan(&teamID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueKeyViolationCode {
			return modelsErr.ErrTeamExist
		}
		return err
	}

	insertIntoUser := sq.Insert("users").
		Columns("id", "name", "is_active", "team_id")
	for _, member := range team.Members {
		insertIntoUser = insertIntoUser.
			Values(
				member.UserID,
				member.Username,
				member.IsActive,
				teamID,
			)
	}

	insertIntoUser = insertIntoUser.Suffix(`
	ON CONFLICT (id) DO UPDATE
	SET name = EXCLUDED.name,
	    is_active = EXCLUDED.is_active,
	    team_id = EXCLUDED.team_id
	`).PlaceholderFormat(sq.Dollar)

	insertIntoUserStr, args, err := insertIntoUser.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertIntoUserStr, args...)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (p *postgresRepo) TeamGet(ctx context.Context, teamName string) (*team.DBTeam, error) {
	getTeamByID := sq.Select("id").
		From("team").
		Where(sq.Eq{"name": teamName}).PlaceholderFormat(sq.Dollar)

	getTeamByIDStr, args, err := getTeamByID.ToSql()
	if err != nil {
		return nil, err
	}

	var teamID int
	err = p.db.QueryRow(ctx, getTeamByIDStr, args...).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, modelsErr.ErrTeamNotFound
		}
		return nil, err
	}

	getUsers := sq.Select("id", "name", "is_active").
		From("users").
		Where(sq.Eq{"team_id": teamID}).PlaceholderFormat(sq.Dollar)

	getUsersStr, args, err := getUsers.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := p.db.Query(ctx, getUsersStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbTeam team.DBTeam
	dbTeam.TeamName = teamName
	for rows.Next() {
		var member team.DBTeamMember
		err = rows.Scan(
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
