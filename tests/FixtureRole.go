package tests

import (
	"fmt"
	"idm/inner/role"
	"time"
)

type FixtureRole struct {
	role *role.Repository
}

func NewFixtureRole(role *role.Repository) *FixtureRole {
	if err := InitSchemaRole(role); err != nil {
		panic(err)
	}
	return &FixtureRole{role}
}

func InitSchemaRole(r *role.Repository) error {
	schema := `CREATE TABLE IF NOT EXISTS role (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW());`
	_, err := r.DB().Exec(schema)
	if err != nil {
		return fmt.Errorf("InitSchema error: %w", err)
	}
	return nil
}

func (f *FixtureRole) Role(name string, createdAt time.Time, updatedAt time.Time) int64 {
	var entity = role.Entity{
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	var newId, err = f.role.Add(entity)
	if err != nil {
		panic(err)
	}
	return newId
}
