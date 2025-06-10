CREATE TABLE IF NOT EXISTS employees (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT,
    surname     TEXT NOT NULL,
    age         SMALLINT CHECK (age > 16 AND age < 91),
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);

COMMENT ON TABLE employees IS 'Сотрудники'