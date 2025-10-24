package employee

import (
	"context"
	"fmt"
	"idm/inner/common"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo      Repo
	validator *validator.Validate
	logger    *common.Logger
}

type Repo interface {
	Add(tx *sqlx.Tx, employee Entity) (id int64, err error)
	FindById(id int64) (Entity, error)
	FindAll(ctx context.Context) ([]Entity, error)
	FindBySliceIds(ids []int64) ([]Entity, error)
	DeleteById(id int64) (bool, error)
	DeleteBySliceIds(ids []int64) ([]int64, error)
	BeginTr() (*sqlx.Tx, error)
	FindByNameAndSurname(tx *sqlx.Tx, name, surname string) (isExists bool, err error)
	FindWithLimitOffsetAndFilter(ctx context.Context, limit int64, offset int64, filter string) (employees []Entity, total int64, err error)
}

type Validator interface {
	Validate(request any) error
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo, validator: validator.New()}
}

func (svc *Service) FindById(ctx context.Context, id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("Wrong id: %d", id)
	}
	var entity, err = svc.repo.FindById(id)
	if err != nil {
		return Response{}, fmt.Errorf("Error finding employee with id %d: %w", id, err)
	}
	return entity.ToResponse(), nil
}

func (svc *Service) Add(ctx context.Context, employee Entity) (response Response, err error) {
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

func (svc *Service) CreateEmployee(ctx context.Context, request CreateRequest) (int64, error) {

	var err = svc.validator.Struct(request)
	if err != nil {
		return 0, common.RequestValidationError{Message: err.Error()}
	}

	tx, err := svc.repo.BeginTr()
	if err != nil || tx == nil {
		return 0, fmt.Errorf("Failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Сreating employee panic: %v", r)
			err := tx.Rollback()
			if err != nil {
				_ = fmt.Errorf("Creating employee: rolling back transaction errors: %w", err)
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
	if err != nil {
		return 0, fmt.Errorf("Error create employee: error creating transaction: %w", err)
	}

	isExist, err := svc.repo.FindByNameAndSurname(tx, request.Name, request.Surname)
	if err != nil {
		return 0, fmt.Errorf("Error finding employee by name and suename : %s, %s, %w", request.Name, request.Surname, err)
	}
	if isExist {
		return 0, common.AlreadyExistsError{
			Message: fmt.Sprintf("Employee with name %s and surname %s already exists", request.Name, request.Surname),
		}
	}

	newEmployeeId, err := svc.repo.Add(tx, request.ToEntity())
	if err != nil {
		err = fmt.Errorf("Error creating employee with name and sruanem: %s  %s %v", request.Name, request.Surname, err)
	}
	return newEmployeeId, err
}

func (svc *Service) FindByIds(ctx context.Context, ids []int64) ([]Response, error) {
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

func (svc *Service) DeleteByIds(ctx context.Context, ids []int64) ([]Response, error) {
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

func (svc *Service) DeleteById(ctx context.Context, id int64) (Response, error) {
	if id <= 0 {
		return Response{}, fmt.Errorf("Wrong id: %d", id)
	}
	var rsl, err = svc.repo.DeleteById(id)
	if err != nil || !rsl {
		return Response{}, fmt.Errorf("Error deleting employee with id %d: %w", id, err)
	}
	return Response{Id: id}, nil
}

func (svc *Service) FindAll(ctx context.Context) (employees []Response, err error) {
	rsl, err := svc.repo.FindAll(ctx)
	if err != nil {
		return []Response{}, fmt.Errorf("Error finding employees: %w", err)
	}
	for _, e := range rsl {
		employees = append(employees, e.ToResponse())
	}
	return employees, nil
}

func (svc *Service) FindAllWithLimitOffset(ctx context.Context, req PageRequest) (result PageResponse, err error) {
	if err := svc.validator.Struct(req); err != nil {
		return PageResponse{}, common.RequestValidationError{Message: err.Error()}
	}
	limit := req.PageSize
	offset := req.PageNumber * req.PageSize
	entities, total, err := svc.repo.FindWithLimitOffsetAndFilter(ctx, int64(limit), int64(offset), req.TextFilter)
	if err != nil {
		return PageResponse{}, fmt.Errorf("Error finding employees with limit/offset: %w", err)
	}

	resp := make([]Response, 0, len(entities))
	for _, e := range entities {
		resp = append(resp, e.ToResponse())
	}
	return PageResponse{
		Result:     resp,
		PageSize:   req.PageSize,
		PageNum:    req.PageNumber,
		Total:      total,
		TextFilter: req.TextFilter,
	}, nil
}
