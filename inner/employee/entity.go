package employee

import "time"

type Entity struct {
	Id        int64     `db:"id"`
	Name      string    `db:"name"`
	Surname   string    `db:"surname"`
	Age       int8      `db:"age"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CreateRequest struct {
	Name      string    `json:"name" validate:"required,min=2,max=155"`
	Surname   string    `json:"surname" validate:"required,min=2,max=155"`
	Age       int8      `json:"age" validate:"required,min=16,max=90"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

func (req *CreateRequest) ToEntity() Entity {
	return Entity{Name: req.Name,
		Surname:   req.Surname,
		Age:       req.Age,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt}
}

func (e *Entity) ToResponse() Response {
	return Response{
		Id:        e.Id,
		Name:      e.Name,
		Surname:   e.Surname,
		Age:       e.Age,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

type Response struct {
	Id        int64     `json:"id" query:"id"`
	Name      string    `json:"name" query:"name"`
	Surname   string    `json:"surname" query:"surname"`
	Age       int8      `json:"age" query:"age"`
	CreatedAt time.Time `json:"created_at" query:"created_at"`
	UpdatedAt time.Time `json:"updated_at" query:"updated_at"`
}

type PageRequest struct {
	PageNumber int `json:"page_number" query:"page_number" validate:"min=0"`
	PageSize   int `json:"page_size" query:"page_size" validate:"min=1,max=100"`
}

type PageResponse struct {
	Result   []Response `json:"result" query:"result"`
	PageSize int        `json:"page_size" query:"page_size"`
	PageNum  int        `json:"page_num" query:"page_num"`
	Total    int64      `json:"total" query:"total"`
}

type EntityPageResponse struct {
	Success bool         `json:"success"`
	Error   string       `json:"error"`
	Data    PageResponse `json:"data"`
}
