package employee

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEmployeeRepo struct {
	mock.Mock
}

type StubRepoEmployee struct {
	Employees []Entity
	Err       error
}

func (m *MockEmployeeRepo) BeginTr() (*sqlx.Tx, error) {
	args := m.Called()
	tx, _ := args.Get(0).(*sqlx.Tx)
	return tx, args.Error(1)
}

func (m *MockEmployeeRepo) FindByNameAndSurname(tx *sqlx.Tx, name, surname string) (bool, error) {
	args := m.Called(tx, name, surname)
	return args.Bool(0), args.Error(1)
}

func (m *MockEmployeeRepo) Add(tx *sqlx.Tx, emp Entity) (int64, error) {
	args := m.Called(tx, emp)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockEmployeeRepo) FindById(id int64) (Entity, error) {
	args := m.Called(id)
	return args.Get(0).(Entity), args.Error(1)
}

func (m *MockEmployeeRepo) FindAll(context.Context) ([]Entity, error) {
	args := m.Called()
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockEmployeeRepo) DeleteById(id int64) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockEmployeeRepo) DeleteBySliceIds(ids []int64) ([]int64, error) {
	args := m.Called(ids)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockEmployeeRepo) FindBySliceIds(ids []int64) ([]Entity, error) {
	args := m.Called(ids)
	return args.Get(0).([]Entity), args.Error(1)
}

func (m *MockEmployeeRepo) FindWithLimitOffsetAndFilter(ctx context.Context, limit int64, offset int64, filter string) (employees []Entity, total int64, err error) {
	args := m.Called(ctx, limit, offset, filter)
	return args.Get(0).([]Entity), args.Get(1).(int64), args.Error(2)
}

func TestFindById(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	t.Run("Should return found employee", func(t *testing.T) {
		t.Parallel()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)

		entity := Entity{
			Id:        int64(1),
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		want := entity.ToResponse()
		repo.On("FindById", int64(1)).Return(entity, nil)
		got, err := svc.FindById(ctx, 1)

		a.Nil(err)
		a.Equal(want, got)
		repo.AssertExpectations(t)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		t.Parallel()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)

		got, err := svc.FindById(ctx, 0)

		a.Error(err)
		a.Equal(Response{}, got)
	})
}

func TestServiceAdd(t *testing.T) {
	a := assert.New(t)
	ctx := context.Background()
	t.Run("Should add employee", func(t *testing.T) {
		t.Parallel()
		db, mockTr, err := sqlmock.New()
		a.Nil(err)
		defer db.Close()

		repo := new(MockEmployeeRepo)
		svc := NewService(repo)
		sqlxDB := sqlx.NewDb(db, "sqlmock_db")
		mockTr.ExpectBegin()
		mockTr.ExpectCommit()
		tx, err := sqlxDB.Beginx()
		a.Nil(err)

		employee := Entity{
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTr").Return(tx, nil)
		repo.On("FindByNameAndSurname", tx, employee.Name, employee.Surname).Return(false, nil)
		repo.On("Add", tx, mock.MatchedBy(func(e Entity) bool {
			return e.Name == "John" && e.Surname == "Doe" && e.Age == 30
		})).Return(int64(1), nil)
		rsl, err := svc.Add(ctx, employee)
		a.Nil(err)
		a.Equal(employee.Name, rsl.Name)
		a.Equal(employee.Surname, rsl.Surname)
		a.Equal(employee.CreatedAt, rsl.CreatedAt)
		a.NoError(mockTr.ExpectationsWereMet())
		repo.AssertExpectations(t)
	})

	t.Run("Should fail when add duplicated", func(t *testing.T) {
		t.Parallel()
		db, mockTr, err := sqlmock.New()
		a.Nil(err)
		defer db.Close()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)
		sqlxDB := sqlx.NewDb(db, "sqlmock_db")
		mockTr.ExpectBegin()
		mockTr.ExpectRollback()
		tx, err := sqlxDB.Beginx()
		a.Nil(err)

		duplicated := Entity{
			Name:      "John",
			Surname:   "Sina",
			Age:       40,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTr").Return(tx, nil)
		repo.On("FindByNameAndSurname", tx, duplicated.Name, duplicated.Surname).Return(true, nil)
		rsl, err := svc.Add(ctx, duplicated)
		a.Error(err)
		expectedError := fmt.Sprintf("Employee with name '%s' and surname '%s' already exists", duplicated.Name, duplicated.Surname)
		a.Contains(err.Error(), expectedError)
		a.Equal(Response{}, rsl)
		a.NoError(mockTr.ExpectationsWereMet())
		repo.AssertExpectations(t)
	})

	t.Run("Should fail on empty entity", func(t *testing.T) {
		t.Parallel()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)
		rsl, err := svc.Add(ctx, Entity{})
		a.Error(err)
		a.Contains(err.Error(), "Entity is empty")
		a.Equal(Response{}, rsl)
	})

	t.Run("Should fail on invalid fields", func(t *testing.T) {
		t.Parallel()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)
		badEmployee := Entity{
			Name:    "",
			Surname: "Doe",
			Age:     15,
		}
		rsl, err := svc.Add(ctx, badEmployee)
		a.Error(err)
		a.Contains(err.Error(), "Invalid field")
		a.Equal(Response{}, rsl)
	})

	t.Run("Should fail on transaction begin error", func(t *testing.T) {
		t.Parallel()
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)
		employee := Entity{
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTr").Return(nil, fmt.Errorf("tx error"))

		rsl, err := svc.Add(ctx, employee)
		a.Error(err)
		a.Contains(err.Error(), "Failed to begin transaction")
		a.Equal(Response{}, rsl)
	})
}

func TestFindAll(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := NewService(repo)
	t.Run("Should find empty slice employees", func(t *testing.T) {
		repo.On("FindAll").Return([]Entity(nil), nil)
		got, err := svc.FindAll(context.Background())
		a.Nil(err)
		a.Len(got, 0)
	})
}

func TestDeleteById(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := NewService(repo)
	ctx := context.Background()
	t.Run("Should delete employee", func(t *testing.T) {
		t.Parallel()
		repo.On("DeleteById", int64(1)).Return(true, nil)
		got, err := svc.DeleteById(ctx, 1)
		a.Nil(err)
		a.Equal(Response{Id: 1}, got)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		t.Parallel()
		repo.On("DeleteById", int64(0)).Return(false, errors.New("Wrong id: 1"))
		got, err := svc.DeleteById(ctx, 0)
		a.Equal(Response{}, got)
		a.Error(err)
	})

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		t.Parallel()
		repo.On("DeleteById", int64(5)).Return(false, errors.New("Error deleting employee with id"))
		got, err := svc.DeleteById(ctx, 5)
		a.Equal(Response{}, got)
		a.Error(err)
	})
}

func TestFindByIds(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockEmployeeRepo)
	svc := NewService(mockRepo)
	ctx := context.Background()
	t.Run("Should return finding employees", func(t *testing.T) {
		t.Parallel()
		stub := &StubRepoEmployee{
			Employees: []Entity{
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

		got, err := svc.FindByIds(ctx, ids)
		a.Nil(err)
		a.Len(got, 2)
		a.Equal(got[0].Id, int64(1))
		a.Equal(got[1].Name, "Jane")
		a.True(mockRepo.AssertCalled(t, "FindBySliceIds", ids))
	})

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		t.Parallel()
		got, err := svc.FindByIds(ctx, []int64{})
		a.Empty(got)
		a.Error(err)
	})
}

func TestDeleteByIds(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockEmployeeRepo)
	svc := NewService(mockRepo)
	ctx := context.Background()
	t.Run("Should delete employee", func(t *testing.T) {
		t.Parallel()
		ids := []int64{1, 2}
		mockRepo.On("DeleteBySliceIds", ids).Return(ids, nil)
		got, err := svc.DeleteByIds(ctx, ids)
		expected := []Response{{Id: 1}, {Id: 2}}
		a.Nil(err)
		a.Equal(expected, got)
	})

	t.Run("Should return error if ids is empty", func(t *testing.T) {
		t.Parallel()
		var ids []int64
		mockRepo.On("DeleteBySliceIds", ids).Return(nil, errors.New("Wrong ids"))
		got, err := svc.DeleteByIds(ctx, ids)
		a.Empty(got)
		a.Error(err)
	})
}
