-- +goose Up
-- +goose StatementBegin
ALTER TABLE url_rows
    ADD COLUMN user_id INT;

ALTER TABLE url_rows
    ADD CONSTRAINT fk_url_rows_name_users
        FOREIGN KEY (user_id) REFERENCES users(id)
            ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE url_rows
    DROP COLUMN user_id;

ALTER TABLE url_rows
    DROP CONSTRAINT fk_url_rows_name_users;
-- +goose StatementEnd
