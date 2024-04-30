-- +goose Up
-- +goose StatementBegin
create table users(
    id serial primary key,
    token text null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
