-- +goose Up
-- +goose StatementBegin
CREATE TABLE client (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    date_of_birth TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_clients_client_uuid_hash ON client USING hash(uuid);

CREATE OR REPLACE FUNCTION on_update_timestamp ()
    RETURNS TRIGGER
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$
    LANGUAGE plpgsql;

CREATE TRIGGER client_updated_at
    BEFORE UPDATE ON client
    FOR EACH ROW
EXECUTE PROCEDURE on_update_timestamp ();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS client;
-- +goose StatementEnd
