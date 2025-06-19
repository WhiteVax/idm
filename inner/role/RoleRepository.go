package role

import (
	"github.com/jmoiron/sqlx"
)

type RoleRepository struct {
	db *sqlx.DB
}

func NewRoleRepository(database *sqlx.DB) *RoleRepository {
	return &RoleRepository{db: database}
}

func (r *RoleRepository) Add(role RoleEntity) (id int64, err error) {
	query := `INSERT INTO role(name, create_at, update_at)
      	 	  VALUES (:name, :create_at, :update_at)
      	 	  RETURNING id`
	rows, err := r.db.NamedQuery(query, &role)

	if err == nil && rows.Next() && rows.Scan(&id) == nil {
		return id, nil
	}
	return -1, err
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

func (r *RoleRepository) DeleteBySliceIds(ids []int64) ([]int64, error) {

	query, args, err := sqlx.In("DELETE FROM role WHERE id IN (?) RETURNING id", ids)
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
