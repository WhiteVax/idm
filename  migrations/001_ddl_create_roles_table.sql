CREATE TABLE IF NOT EXISTS roles (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);

COMMENT ON TABLE roles is 'Роли'