package role

import "time"

type RoleEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	Create_at time.Time `db:"create_at"`
	Update_at time.Time `db:"update_at"`
}
