package role

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockRoleRepo struct {
	mock.Mock
}

func (m *MockRoleRepo) Add(entity Entity) (int64, error) {
	args := m.Called(entity)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockRoleRepo) FindAll() ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockRoleRepo) DeleteById(id int64) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) DeleteBySliceIds(ids []int64) ([]int64, error) {
	args := m.Called(ids)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRoleRepo) FindBySliceIds(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func TestFindByIdRole(t *testing.T) {
	t.Run("Should return found role", func(t *testing.T) {
		a := assert.New(t)
		repo := new(MockRoleRepo)
		svc := NewService(repo)
		entity := Entity{
			Id:        1,
			Name:      "Admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("FindById", int64(1)).Return(entity, nil)
		got, err := svc.FindById(1)
		want := entity.ToResponse()
		a.NoError(err)
		a.Equal(want, got)
		repo.AssertExpectations(t)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		a := assert.New(t)

		repo := new(MockRoleRepo)
		svc := NewService(repo)

		got, err := svc.FindById(0)

		a.Error(err)
		a.Equal(Response{}, got)
	})
}

func TestAddRole(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRoleRepo)
	svc := NewService(repo)

	t.Run("Should add role", func(t *testing.T) {
		entity := Entity{
			Id:        1,
			Name:      "Admin",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		expectedId := int64(1)
		entityExpected := Response{
			Id:        1,
			Name:      entity.Name,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		}

		repo.On("Add", entity).Return(expectedId, nil)
		got, err := svc.Add(entity)

		a.Nil(err)
		a.Equal(entityExpected, got)
	})

	t.Run("Should return error if any role field is empty", func(t *testing.T) {
		entity := Entity{
			Id:        0,
			Name:      "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		got, err := svc.Add(entity)
		a.Equal(Response{}, got)
		a.Error(err)
	})
}

func TestFindAllRoles(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRoleRepo)
	svc := NewService(repo)
	t.Run("Should find empty slice roles", func(t *testing.T) {
		repo.On("FindAll").Return([]Entity(nil), nil)
		got, err := svc.FindAll()
		a.Nil(err)
		a.Len(got, 0)
	})
}

func TestDeleteByIdRole(t *testing.T) {
	a := assert.New(t)
	repo := new(MockRoleRepo)
	svc := NewService(repo)
	t.Run("Should delete role", func(t *testing.T) {
		repo.On("DeleteById", int64(1)).Return(true, nil)
		got, err := svc.DeleteById(1)
		a.Nil(err)
		a.Equal(Response{Id: 1}, got)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		repo.On("DeleteById", int64(0)).Return(false, errors.New("Wrong id: 1"))
		got, err := svc.DeleteById(0)
		a.Equal(Response{}, got)
		a.Error(err)
	})

	t.Run("Should return error if any role field is empty", func(t *testing.T) {
		repo.On("DeleteById", int64(5)).Return(false, errors.New("Error deleting role with id"))
		got, err := svc.DeleteById(5)
		a.Equal(Response{}, got)
		a.Error(err)
	})
}

func TestFindByIdsRoles(t *testing.T) {
	t.Run("Should return finding roles", func(t *testing.T) {
		a := assert.New(t)
		mockRepo := new(MockRoleRepo)
		svc := NewService(mockRepo)
		now := time.Now()
		roles := []Entity{
			{Id: 1, Name: "Admin", CreatedAt: now, UpdatedAt: now},
			{Id: 2, Name: "Guest", CreatedAt: now, UpdatedAt: now},
		}
		ids := []int64{1, 2}
		mockRepo.On("FindBySliceIds", ids).Return(roles, nil)
		got, err := svc.FindByIds(ids)
		a.NoError(err)
		a.Len(got, 2)
		a.Equal(int64(1), got[0].Id)
		a.Equal("Guest", got[1].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Should return error if empty ids", func(t *testing.T) {
		a := assert.New(t)
		mockRepo := new(MockRoleRepo)
		svc := NewService(mockRepo)
		got, err := svc.FindByIds([]int64{})
		a.Error(err)
		a.Empty(got)
	})
}

func TestDeleteByIdsRoles(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockRoleRepo)
	svc := NewService(mockRepo)
	t.Run("Should delete role", func(t *testing.T) {
		ids := []int64{1, 2}
		mockRepo.On("DeleteBySliceIds", ids).Return(ids, nil)
		got, err := svc.DeleteByIds(ids)
		expected := []Response{{Id: 1}, {Id: 2}}
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
