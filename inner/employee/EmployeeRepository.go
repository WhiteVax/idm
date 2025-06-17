package employee

import (
	"errors"
	"github.com/jmoiron/sqlx"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: database}
}

func (r *EmployeeRepository) Add(employee EmployeeEntity) (id int64, err error) {
	query := `INSERT INTO employee(name, surname, age, create_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = r.db.QueryRow(query, employee.Name, employee.Surname, employee.Age, employee.Create_at, employee.Update_at).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *EmployeeRepository) FindById(id int64) (employee EmployeeEntity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employee WHERE id = $1", id)
	return employee, err
}

func (r *EmployeeRepository) FindAll() (employees []EmployeeEntity, err error) {
	err = r.db.Select(&employees, "SELECT * FROM employee")
	return employees, err
}

func (r *EmployeeRepository) FindBySliceIds(ids []int64) (employees []EmployeeEntity, err error) {
	query, args, err := sqlx.In("SELECT * FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return employees, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&employees, query, args...)
	return employees, err
}

func (r *EmployeeRepository) DeleteById(id int64) (bool, error) {
	result, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	rowInter, err := result.RowsAffected()
	return rowInter > 0, err
}

func (r *EmployeeRepository) DeleteBySliceIds(ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, errors.New("Employee ids is empty")
	}
	query, args, err := sqlx.In("DELETE FROM employee WHERE id IN (?)", ids)
	query = r.db.Rebind(query)
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return false, err
	}
	rowIds, err := result.RowsAffected()
	return rowIds > 0, err
}
