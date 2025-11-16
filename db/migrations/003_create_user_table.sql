-- +goose Up

CREATE TABLE users
(
    id   TEXT PRIMARY KEY,
    name      TEXT        NOT NULL,
    is_active BOOLEAN     NOT NULL DEFAULT TRUE,
    team_id   BIGINT REFERENCES team (id)
);



-- +goose Down
DROP TABLE users;