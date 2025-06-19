package employee

import (
	"github.com/jmoiron/sqlx"
)

type EmployeeRepository struct {
	db *sqlx.DB
}

func NewEmployeeRepository(database *sqlx.DB) *EmployeeRepository {
	return &EmployeeRepository{db: database}
}

func (r *EmployeeRepository) Add(employee EmployeeEntity) (id int64, err error) {
	query := `INSERT INTO employee(name, surname, age, create_at, updated_at) 
			  VALUES (:name, :surname, :age, :create_at, :updated_at) 
			  RETURNING id`
	rows, err := r.db.NamedQuery(query, &employee)

	if err == nil && rows.Next() && rows.Scan(&id) == nil {
		return id, nil
	}
	return -1, err
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

func (r *EmployeeRepository) DeleteBySliceIds(ids []int64) ([]int64, error) {
	query, args, err := sqlx.In("DELETE FROM employee WHERE id IN (?) RETURNING id", ids)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}

	var deletedIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		deletedIDs = append(deletedIDs, id)
	}
	return deletedIDs, nil
}
