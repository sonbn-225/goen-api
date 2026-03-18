-- +goose Up
-- +goose StatementBegin
ALTER TYPE security_event_type ADD VALUE 'stock_dividend';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- PostgreSQL does not support removing values from an ENUM type.
-- +goose StatementEnd
