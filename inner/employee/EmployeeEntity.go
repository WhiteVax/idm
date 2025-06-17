package employee

import "time"

type EmployeeEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	Surname   string    `db:"surname"`
	Age       int8      `db:"age"`
	Create_at time.Time `db:"create_at"`
	Update_at time.Time `db:"update_at"`
}
