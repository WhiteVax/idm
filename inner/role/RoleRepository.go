package role

import (
	"errors"
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(database *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: database}
}

func (r *RoleRepository) Add(role RoleEntity) (id int64, err error) {
	query := `INSERT INTO role(name, create_at, updated_at) VALUES ($1, $2, $3) RETURNING id`
	err = r.db.QueryRow(query, role.Name, role.Create_at, role.Update_at).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (r *RoleRepository) FindById(id int64) (role RoleEntity, err error) {
	err = r.db.Get(&role, "SELECT * FROM role WHERE id = $1", id)
	return role, err
}

func (r *RoleRepository) FindAll() (roles []RoleEntity, err error) {
	err = r.db.Select(&roles, "SELECT * FROM role")
	return roles, err
}

func (r *RoleRepository) FindBySliceIds(ids []int64) (roles []RoleEntity, err error) {
	query, args, err := sqlx.In("SELECT * FROM role WHERE id IN (?)", ids)
	if err != nil {
		return roles, err
	}
	query = r.db.Rebind(query)
	err = r.db.Select(&roles, query, args...)
	return roles, err
}

func (r *RoleRepository) DeleteById(id int64) (bool, error) {
	result, err := r.db.Exec("DELETE FROM role WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	rowInter, err := result.RowsAffected()
	return rowInter > 0, err
}

func (r *RoleRepository) DeleteBySliceIds(ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, errors.New("Roles ids is empty")
	}
	query, args, err := sqlx.In("DELETE FROM role WHERE id IN (?)", ids)
	query = r.db.Rebind(query)
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return false, err
	}
	rowIds, err := result.RowsAffected()
	return rowIds > 0, err
}
