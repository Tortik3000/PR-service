-- +goose Up

CREATE TABLE pull_request
(
    id                  TEXT PRIMARY KEY,
    name                TEXT    NOT NULL,
    author_id             TEXT REFERENCES users (id),
    created_at          TIMESTAMP        DEFAULT now() NOT NULL,
    merged_at           TIMESTAMP,
    status              int              DEFAULT 0
);


-- +goose Down
DROP TABLE pull_request;