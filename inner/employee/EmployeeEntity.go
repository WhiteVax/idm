package employee

import "time"

type EmployeeEntity struct {
	id        int64     `db:"id"`
	name      string    `db:"name"`
	surname   string    `db:"surname"`
	age       int8      `db:"age"`
	create_at time.Time `db:"create_at"`
	update_at time.Time `db:"update_at"`
}
