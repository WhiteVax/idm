package employee

import (
	"fmt"
)

type Service struct {
	repo Repo
}

type Repo interface {
	Add(entity Entity) (int64, error)
	FindById(id int64) (Entity, error)
	FindAll() ([]Entity, error)
	FindBySliceIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) (bool, error)
	DeleteBySliceIds(ids []int64) ([]int64, error)
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
	return entity.toResponse(), nil
}

func (svc *Service) Add(employee Entity) (Response, error) {
	if employee == (Entity{}) {
		return Response{}, fmt.Errorf("Entity is empty, please check the employee")
	}
	if employee.Name == "" || employee.Surname == "" || employee.Age <= 16 {
		return Response{}, fmt.Errorf("Invalid field, please check the employee %+v", employee)
	}
	var rsl, err = svc.repo.Add(employee)
	if err != nil {
		return Response{}, fmt.Errorf("Error adding employee %+v: %w", employee, err)
	}
	return Response{
		Id:        rsl,
		Name:      employee.Name,
		Surname:   employee.Surname,
		Age:       employee.Age,
		CreatedAt: employee.CreatedAt,
		UpdatedAt: employee.UpdatedAt}, nil
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
		responses = append(responses, e.toResponse())
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

func (svc *Service) FindAll() (employees []Entity, err error) {
	return svc.repo.FindAll()
}
