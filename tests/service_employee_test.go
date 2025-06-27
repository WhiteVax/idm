package tests

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/employee"
	"testing"
	"time"
)

type MockEmployeeRepo struct {
	mock.Mock
}

type StubRepo struct {
	Employees []employee.Entity
	Err       error
}

func (m *MockEmployeeRepo) Add(entity employee.Entity) (int64, error) {
	args := m.Called(entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEmployeeRepo) FindById(id int64) (employee.Entity, error) {
	args := m.Called(id)
	return args.Get(0).(employee.Entity), args.Error(1)
}

func (m *MockEmployeeRepo) FindAll() ([]employee.Entity, error) {
	args := m.Called()
	return args.Get(0).([]employee.Entity), args.Error(1)
}

func (m *MockEmployeeRepo) DeleteById(id int64) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockEmployeeRepo) DeleteBySliceIds(ids []int64) ([]int64, error) {
	args := m.Called(ids)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEmployeeRepo) FindBySliceIds(ids []int64) ([]employee.Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]employee.Entity), args.Error(1)
}

func TestFindById(t *testing.T) {
	a := assert.New(t)

	t.Run("Should return found employee", func(t *testing.T) {
		repo := new(MockEmployeeRepo)
		svc := employee.NewService(repo)

		entity := employee.Entity{
			Id:        int64(1),
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		want := entity.ToResponse()
		repo.On("FindById", int64(1)).Return(entity, nil)
		got, err := svc.FindById(1)

		a.Nil(err)
		a.Equal(want, got)
		repo.AssertExpectations(t)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		repo := new(MockEmployeeRepo)
		svc := employee.NewService(repo)

		got, err := svc.FindById(0)

		a.Error(err)
		a.Equal(employee.Response{}, got)
	})
}

func TestAdd(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := employee.NewService(repo)

	t.Run("Should add employee", func(t *testing.T) {
		entity := employee.Entity{
			Id:        1,
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		expectedId := int64(1)
		entityExpected := employee.Response{
			Id:        1,
			Name:      entity.Name,
			Surname:   entity.Surname,
			Age:       entity.Age,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		}

		repo.On("Add", entity).Return(expectedId, nil)
		got, err := svc.Add(entity)

		a.Nil(err)
		a.Equal(entityExpected, got)
	})

	t.Run("Should return error if employee empty", func(t *testing.T) {
		entity := employee.Entity{}
		got, err := svc.Add(entity)
		a.Equal(employee.Response{}, got)
		a.Error(err)
		a.Contains(err.Error(), "Entity is empty, please check the employee")
	})

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		entity := employee.Entity{
			Id:        0,
			Name:      "John",
			Surname:   "",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		got, err := svc.Add(entity)
		a.Equal(employee.Response{}, got)
		a.Error(err)
	})
}

func TestFindAll(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := employee.NewService(repo)
	t.Run("Should find empty slice employees", func(t *testing.T) {
		repo.On("FindAll").Return([]employee.Entity(nil), nil)
		got, err := svc.FindAll()
		a.Nil(err)
		a.Len(got, 0)
	})
}

func TestDeleteById(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := employee.NewService(repo)
	t.Run("Should delete employee", func(t *testing.T) {
		repo.On("DeleteById", int64(1)).Return(true, nil)
		got, err := svc.DeleteById(1)
		a.Nil(err)
		a.Equal(employee.Response{Id: 1}, got)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		repo.On("DeleteById", int64(0)).Return(false, errors.New("Wrong id: 1"))
		got, err := svc.DeleteById(0)
		a.Equal(employee.Response{}, got)
		a.Error(err)
	})

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		repo.On("DeleteById", int64(5)).Return(false, errors.New("Error deleting employee with id"))
		got, err := svc.DeleteById(5)
		a.Equal(employee.Response{}, got)
		a.Error(err)
	})
}

func TestFindByIds(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockEmployeeRepo)
	svc := employee.NewService(mockRepo)

	t.Run("Should return finding employees", func(t *testing.T) {
		// Stub
		stub := &StubRepo{
			Employees: []employee.Entity{
				{
					Id:        1,
					Name:      "John",
					Surname:   "Doe",
					Age:       30,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				{
					Id:        2,
					Name:      "Jane",
					Surname:   "Smith",
					Age:       28,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		ids := []int64{1, 2}
		mockRepo.On("FindBySliceIds", ids).Return(stub.Employees, nil)

		got, err := svc.FindByIds(ids)
		a.Nil(err)
		a.Len(got, 2)
		a.Equal(got[0].Id, int64(1))
		a.Equal(got[1].Name, "Jane")
		a.True(mockRepo.AssertCalled(t, "FindBySliceIds", ids))
	})

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		got, err := svc.FindByIds([]int64{})
		a.Empty(got)
		a.Error(err)
	})
}

func TestDeleteByIds(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockEmployeeRepo)
	svc := employee.NewService(mockRepo)
	t.Run("Should delete employee", func(t *testing.T) {
		ids := []int64{1, 2}
		mockRepo.On("DeleteBySliceIds", ids).Return(ids, nil)
		got, err := svc.DeleteByIds(ids)
		expected := []employee.Response{{Id: 1}, {Id: 2}}
		a.Nil(err)
		a.Equal(expected, got)
	})

	t.Run("Should return error if ids is empty", func(t *testing.T) {
		var ids []int64
		mockRepo.On("DeleteBySliceIds", ids).Return(nil, errors.New("Wrong ids"))
		got, err := svc.DeleteByIds(ids)
		a.Empty(got)
		a.Error(err)
	})
}
