-- +goose Up
-- +goose StatementBegin
ALTER TABLE url_rows
    ADD COLUMN user_id UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url_rows
    DROP COLUMN user_id;
-- +goose StatementEnd
