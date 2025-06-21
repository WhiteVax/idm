package employee

import "time"

type EmployeeEntity struct {
	Id       int64     `db:"id"`
	Name     string    `db:"name"`
	Surname  string    `db:"surname"`
	Age      int8      `db:"age"`
	CreateAt time.Time `db:"create_at"`
	UpdateAt time.Time `db:"update_at"`
}
