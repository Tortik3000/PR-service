-- +goose Up

CREATE TYPE enum_status as ENUM ('MERGED', 'OPEN');

CREATE TABLE status_table
(
    id     INT PRIMARY KEY,
    status enum_status
);

INSERT INTO status_table (status, id) VALUES ('MERGED', 1);
INSERT INTO status_table (status, id) VALUES ('OPEN', 0);



-- +goose Down
DROP TABLE status_table;