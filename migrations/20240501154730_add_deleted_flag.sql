-- +goose Up
-- +goose StatementBegin
ALTER TABLE url_rows
    ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url_rows
    DROP COLUMN is_deleted;
-- +goose StatementEnd
