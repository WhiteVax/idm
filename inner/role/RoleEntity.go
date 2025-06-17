package role

import "time"

type RoleEntity struct {
	id        int64     `db:"id"`
	name      string    `db:"name"`
	create_at time.Time `db:"create_at"`
	update_at time.Time `db:"update_at"`
}
