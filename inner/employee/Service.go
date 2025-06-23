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

func (svc *Service) Add(entity Entity) (Response, error) {
	if entity == (Entity{}) {
		return Response{}, fmt.Errorf("Entity is empty, please check the employee")
	}
	if entity.Name == "" || entity.Surname == "" || entity.Age <= 16 {
		return Response{}, fmt.Errorf("Invalid field, please check the employee %+v", entity)
	}
	var rsl, err = svc.repo.Add(entity)
	if err != nil {
		return Response{}, fmt.Errorf("Error adding employee %+v: %w", entity, err)
	}
	return Response{
		Id:        rsl,
		Name:      entity.Name,
		Surname:   entity.Surname,
		Age:       entity.Age,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt}, nil
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

func (svc *Service) DeleteById(id int64) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("Wrong id: %d", id)
	}
	var rsl, err = svc.repo.DeleteById(id)
	if err != nil || !rsl {
		return "", fmt.Errorf("Error deleting employee with id %d: %w", id, err)
	}
	return fmt.Sprintf("Employee with id %d deleted", id), nil
}
