-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_unique_original_url ON url_rows (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_unique_original_url;
-- +goose StatementEnd
