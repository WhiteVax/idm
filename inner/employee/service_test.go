package employee

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
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

func (m *MockEmployeeRepo) FindAll() ([]Entity, error) {
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

func TestFindById(t *testing.T) {
	a := assert.New(t)

	t.Run("Should return found employee", func(t *testing.T) {
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
		got, err := svc.FindById(1)

		a.Nil(err)
		a.Equal(want, got)
		repo.AssertExpectations(t)
	})

	t.Run("Should return error if id <= 0", func(t *testing.T) {
		repo := new(MockEmployeeRepo)
		svc := NewService(repo)

		got, err := svc.FindById(0)

		a.Error(err)
		a.Equal(Response{}, got)
	})
}

func TestService_Add(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := NewService(repo)

	db, mockTr, err := sqlmock.New()
	a.Nil(err)
	defer db.Close()

	t.Run("Should add employee", func(t *testing.T) {
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
		repo.On("Add", tx, employee).Return(int64(1), nil)
		rsl, err := svc.Add(employee)
		a.Nil(err)
		a.Equal(employee.Name, rsl.Name)
		a.Equal(employee.Surname, rsl.Surname)
		a.Equal(employee.CreatedAt, rsl.CreatedAt)
		a.NoError(mockTr.ExpectationsWereMet())
		repo.AssertExpectations(t)
	})

	t.Run("Should fail when add duplicated", func(t *testing.T) {
		sqlxDB := sqlx.NewDb(db, "sqlmock_db")
		mockTr.ExpectBegin()
		mockTr.ExpectRollback()
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
		repo.On("FindByNameAndSurname", tx, employee.Name, employee.Surname).Return(true, nil)

		rsl, err := svc.Add(employee)
		a.Error(err)
		expectedError := fmt.Errorf("Employee with name '%s' and surname '%s' already exists", employee.Name, employee.Surname)
		a.Contains(err.Error(), expectedError.Error())
		a.Equal(Response{}, rsl)
		a.NoError(mockTr.ExpectationsWereMet())
		repo.AssertExpectations(t)
	})

	t.Run("Should fail on empty entity", func(t *testing.T) {
		rsl, err := svc.Add(Entity{})
		a.Error(err)
		a.Contains(err.Error(), "Entity is empty")
		a.Equal(Response{}, rsl)
	})

	t.Run("Should fail on invalid fields", func(t *testing.T) {
		badEmployee := Entity{
			Name:    "",
			Surname: "Doe",
			Age:     15,
		}
		rsl, err := svc.Add(badEmployee)
		a.Error(err)
		a.Contains(err.Error(), "Invalid field")
		a.Equal(Response{}, rsl)
	})

	t.Run("Should fail on transaction begin error", func(t *testing.T) {
		employee := Entity{
			Name:      "John",
			Surname:   "Doe",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo.On("BeginTr").Return(nil, fmt.Errorf("tx error"))

		rsl, err := svc.Add(employee)
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
		got, err := svc.FindAll()
		a.Nil(err)
		a.Len(got, 0)
	})
}

func TestDeleteById(t *testing.T) {
	a := assert.New(t)
	repo := new(MockEmployeeRepo)
	svc := NewService(repo)
	t.Run("Should delete employee", func(t *testing.T) {
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

	t.Run("Should return error if any employee field is empty", func(t *testing.T) {
		repo.On("DeleteById", int64(5)).Return(false, errors.New("Error deleting employee with id"))
		got, err := svc.DeleteById(5)
		a.Equal(Response{}, got)
		a.Error(err)
	})
}

func TestFindByIds(t *testing.T) {
	a := assert.New(t)
	mockRepo := new(MockEmployeeRepo)
	svc := NewService(mockRepo)

	t.Run("Should return finding employees", func(t *testing.T) {
		// Stub
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
	svc := NewService(mockRepo)
	t.Run("Should delete employee", func(t *testing.T) {
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
