-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS ref_token (
    ref_token VARCHAR(255) UNIQUE,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ref_token;
-- +goose StatementEnd