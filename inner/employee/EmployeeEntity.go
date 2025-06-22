package employee

import "time"

type EmployeeEntity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	Surname   string    `db:"surname"`
	Age       int8      `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
