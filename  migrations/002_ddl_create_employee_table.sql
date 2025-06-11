CREATE TABLE IF NOT EXISTS employee (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL,
    surname     TEXT NOT NULL,
    age         SMALLINT CHECK (age > 16 AND age < 91),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMENT ON TABLE employee IS 'Сотрудники'