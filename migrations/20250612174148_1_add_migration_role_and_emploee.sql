-- +goose Up
CREATE TABLE IF NOT EXISTS role
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
COMMENT ON TABLE role is 'Роли';
CREATE TABLE IF NOT EXISTS employee (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL,
    surname     TEXT NOT NULL,
    age         SMALLINT CHECK (age > 16 AND age < 91),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
    );
COMMENT ON TABLE employee IS 'Сотрудники';
-- +goose Down
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS role;