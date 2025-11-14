-- +goose Up

CREATE TABLE team
(
    id   BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);


-- +goose Down
DROP TABLE team;