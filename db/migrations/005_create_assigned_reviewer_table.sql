-- +goose Up

CREATE TABLE assigned_reviewer
(
    user_id TEXT REFERENCES users(id),
    pr_id   TEXT REFERENCES pull_request (id),
    PRIMARY KEY (user_id, pr_id)
);


-- +goose Down
DROP TABLE assigned_reviewer;