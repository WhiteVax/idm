package tests

import (
	"fmt"
	_ "idm/inner/database"
	"idm/inner/employee"
	"time"
)

type Fixture struct {
	employees *employee.EmployeeRepository
}

func NewFixture(employees *employee.EmployeeRepository) *Fixture {
	if err := InitSchema(employees); err != nil {
		panic(err)
	}
	return &Fixture{employees}
}

func InitSchema(r *employee.EmployeeRepository) error {
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

func (f *Fixture) Employee(name string, surname string, age int8,
	createAt time.Time, updateAt time.Time) int64 {
	var entity = employee.EmployeeEntity{
		Name:      name,
		Surname:   surname,
		Age:       age,
		CreatesAt: createAt,
		UpdatedAt: updateAt,
	}
	var newId, err = f.employees.Add(entity)
	if err != nil {
		panic(err)
	}
	return newId
}
