-- +goose Up
-- +goose StatementBegin
CREATE TABLE url_rows (
    uuid UUID PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE url_rows;
-- +goose StatementEnd
