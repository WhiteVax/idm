package employee

import (
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type Repository struct {
	db *sqlx.DB
}

func (r *Repository) DB() *sqlx.DB {
	return r.db
}

func NewEmployeeRepository(database *sqlx.DB) *Repository {
	return &Repository{db: database}
}

func (r *Repository) BeginTr() (*sqlx.Tx, error) {
	return r.db.Beginx()
}

func (r *Repository) Add(tx *sqlx.Tx, employee Entity) (id int64, err error) {
	query := `INSERT INTO employee(name, surname, age, created_at, updated_at) 
			  VALUES (:name, :surname, :age, :created_at, :updated_at) 
			  RETURNING id`
	rows, err := tx.NamedQuery(query, &employee)

	if err == nil && rows.Next() && rows.Scan(&id) == nil {
		return id, nil
	}
	return -1, err
}

func (r *Repository) FindByNameAndSurname(tx *sqlx.Tx, name, surname string) (isExists bool, err error) {
	err = tx.Get(
		&isExists,
		"select exists(select from employee where name = $1 and surname = $2)",
		name, surname)
	return isExists, err
}

func (r *Repository) FindById(id int64) (employee Entity, err error) {
	err = r.db.Get(&employee, "SELECT * FROM employee WHERE id = $1", id)
	return employee, err
}

func (r *Repository) FindAll(ctx context.Context) (employees []Entity, err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err = r.db.SelectContext(ctx, &employees, "SELECT * FROM employee")
	return employees, err
}

func (r *Repository) FindBySliceIds(ids []int64) (employees []Entity, err error) {
	query, args, err := sqlx.In("SELECT * FROM employee WHERE id IN (?)", ids)
	if err != nil {
		return employees, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&employees, query, args...)
	return employees, err
}

func (r *Repository) DeleteById(id int64) (bool, error) {
	result, err := r.db.Exec("DELETE FROM employee WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	rowInter, err := result.RowsAffected()
	return rowInter > 0, err
}

func (r *Repository) DeleteBySliceIds(ids []int64) ([]int64, error) {
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
