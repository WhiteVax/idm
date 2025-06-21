package role

import "time"

type RoleEntity struct {
	Id       int64     `db:"id"`
	Name     string    `db:"name"`
	CreateAt time.Time `db:"create_at"`
	UpdateAt time.Time `db:"update_at"`
}
