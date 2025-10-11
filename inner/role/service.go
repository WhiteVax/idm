package role

import "fmt"

type Service struct {
	repo Repo
}

type Repo interface {
	Add(role Entity) (id int64, err error)
	FindById(id int64) (role Entity, err error)
	FindAll() (roles []Entity, err error)
	FindBySliceIds(ids []int64) (roles []Entity, err error)
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
		return Response{}, fmt.Errorf("Wrong id role: %d", id)
	}
	var entity, err = svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("Error finding role with id %d: %w", id, err)
	}
	return entity.ToResponse(), nil
}

func (svc *Service) Add(role Entity) (Response, error) {
	if role == (Entity{}) || role.Name == "" {
		return Response{Name: role.Name}, fmt.Errorf("Invalid field, please check the role")
	}
	var rsl, err = svc.repo.Add(role)
	if err != nil {
		return Response{}, fmt.Errorf("Error adding role %+v: %w", role, err)
	}
	return Response{
		Id:        rsl,
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt}, nil
}

func (svc *Service) FindByIds(ids []int64) ([]Response, error) {
	if len(ids) == 0 {
		return []Response{}, fmt.Errorf("No roles ids provided")
	}
	var rsl, err = svc.repo.FindBySliceIds(ids)
	if err != nil {
		return []Response{}, fmt.Errorf("Error finding roles by ids %+v: %w", ids, err)
	}

	responses := make([]Response, 0, len(rsl))
	for _, e := range rsl {
		responses = append(responses, e.ToResponse())
	}
	return responses, nil
}

func (svc *Service) DeleteByIds(ids []int64) ([]Response, error) {
	if len(ids) == 0 {
		return []Response{}, fmt.Errorf("No roles ids provided")
	}
	rsl, err := svc.repo.DeleteBySliceIds(ids)
	if err != nil {
		return []Response{}, fmt.Errorf("Error deleting roles by ids %+v: %w", ids, err)
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
		return Response{}, fmt.Errorf("Error deleting role with id %d: %w", id, err)
	}
	return Response{Id: id}, nil
}

func (svc *Service) FindAll() (roles []Entity, err error) {
	return svc.repo.FindAll()
}
