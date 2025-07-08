package tests

import (
	"fmt"
	_ "idm/inner/database"
	"idm/inner/employee"
	"time"
)

type FixtureEmployee struct {
	employee *employee.Repository
}

func NewFixtureEmployee(employee *employee.Repository) *FixtureEmployee {
	if err := InitSchemaEmployee(employee); err != nil {
		panic(err)
	}
	return &FixtureEmployee{employee}
}

func InitSchemaEmployee(r *employee.Repository) error {
	schema := `
	CREATE TABLE IF NOT EXISTS employee (
		id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		name        TEXT NOT NULL,
		surname     TEXT NOT NULL,
		age         SMALLINT CHECK (age > 16 AND age < 91),
		"create_at"  TIMESTAMPTZ NOT NULL DEFAULT now(),
		"update_at"  TIMESTAMPTZ NOT NULL DEFAULT now()
	);`
	// Подключение к бд с созданием таблицы
	_, err := r.DB().Exec(schema)
	if err != nil {
		return fmt.Errorf("InitSchema error: %w", err)
	}
	return nil
}

func (f *FixtureEmployee) Employee(name string, surname string, age int8,
	createdAt time.Time, updatedAt time.Time) int64 {
	var entity = employee.Entity{
		Name:      name,
		Surname:   surname,
		Age:       age,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	tx, err := f.employee.BeginTr()
	if err != nil {
		panic(fmt.Errorf("Failed to begin transaction: %w", err))
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	newId, err := f.employee.Add(tx, entity)
	if err != nil {
		panic(err)
	}
	return newId
}
