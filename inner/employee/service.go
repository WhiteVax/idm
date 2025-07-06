package employee

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo Repo
}

type Repo interface {
	Add(tx *sqlx.Tx, employee Entity) (id int64, err error)
	FindById(id int64) (Entity, error)
	FindAll() ([]Entity, error)
	FindBySliceIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) (bool, error)
	DeleteBySliceIds(ids []int64) ([]int64, error)
	BeginTr() (*sqlx.Tx, error)
	FindByNameAndSurname(tx *sqlx.Tx, name, surname string) (isExists bool, err error)
}

func NewService(
	repo Repo,
) *Service {
	return &Service{
		repo: repo,
	}
}

func (svc *Service) FindById(id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("Wrong id: %d", id)
	}
	var entity, err = svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("Error finding employee with id %d: %w", id, err)
	}
	return entity.ToResponse(), nil
}

func (svc *Service) Add(employee Entity) (response Response, err error) {
	if employee == (Entity{}) {
		return Response{}, fmt.Errorf("Entity is empty, please check the employee")
	}
	if employee.Name == "" || employee.Surname == "" || employee.Age <= 16 {
		return Response{}, fmt.Errorf("Invalid field, please check the employee %+v", employee)
	}

	tx, err := svc.repo.BeginTr()
	if err != nil || tx == nil {
		return Response{}, fmt.Errorf("Failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Сreating employee panic: %v", r)
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("Сreating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		} else if err != nil {
			errTx := tx.Rollback()
			if errTx != nil {
				err = fmt.Errorf("Сreating employee: rolling back transaction errors: %w, %w", err, errTx)
			}
		} else {
			errTx := tx.Commit()
			if errTx != nil {
				err = fmt.Errorf("Сreating employee: commiting transaction error: %w", errTx)
			}
		}
	}()

	exists, err := svc.repo.FindByNameAndSurname(tx, employee.Name, employee.Surname)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to check existence: %w", err)
	}
	if exists {
		return Response{}, fmt.Errorf("Employee with name '%s' and surname '%s' already exists", employee.Name, employee.Surname)
	}

	id, err := svc.repo.Add(tx, employee)
	if err != nil {
		return Response{}, fmt.Errorf("Failed to add employee: %w", err)
	}

	return Response{
		Id:        id,
		Name:      employee.Name,
		Surname:   employee.Surname,
		Age:       employee.Age,
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt,
	}, nil
}

func (svc *Service) FindByIds(ids []int64) ([]Response, error) {
	if len(ids) == 0 {
		return []Response{}, fmt.Errorf("No employees ids provided")
	}
	var rsl, err = svc.repo.FindBySliceIds(ids)
	if err != nil {
		return []Response{}, fmt.Errorf("Error finding employees by ids %+v: %w", ids, err)
	}

	responses := make([]Response, 0, len(rsl))
	for _, e := range rsl {
		responses = append(responses, e.ToResponse())
	}
	return responses, nil
}

func (svc *Service) DeleteByIds(ids []int64) ([]Response, error) {
	if len(ids) == 0 {
		return []Response{}, fmt.Errorf("No employees ids provided")
	}
	rsl, err := svc.repo.DeleteBySliceIds(ids)
	if err != nil {
		return []Response{}, fmt.Errorf("Error deleting employees by ids %+v: %w", ids, err)
	}
	responses := make([]Response, 0, len(rsl))
	for _, id := range rsl {
		responses = append(responses, Response{Id: id})
	}
	return responses, nil
}

func (svc *Service) DeleteById(id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("Wrong id: %d", id)
	}
	var rsl, err = svc.repo.DeleteById(id)
	if err != nil || !rsl {
		return Response{}, fmt.Errorf("Error deleting employee with id %d: %w", id, err)
	}
	return Response{Id: id}, nil
}

func (svc *Service) FindAll() (employees []Response, err error) {
	rsl, err := svc.repo.FindAll()
	if err != nil {
		return []Response{}, fmt.Errorf("Error finding employees: %w", err)
	}
	for _, e := range rsl {
		employees = append(employees, e.ToResponse())
	}
	return employees, nil
}
