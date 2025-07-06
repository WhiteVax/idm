package tests

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/database"
	"idm/inner/employee"
	"testing"
	"time"
)

func TestRepositoryEmployee(t *testing.T) {
	a := assert.New(t)

	var db = database.ConnectDb()
	var clearDatabase = func() {
		db.MustExec("DELETE FROM employee")
	}
	defer func() {
		if r := recover(); r != nil {
			clearDatabase()
		} else {
			clearDatabase()
		}
	}()
	var employeeRepository = employee.NewEmployeeRepository(db)
	var fixture = NewFixtureEmployee(employeeRepository)

	t.Run("Find an employee by id", func(t *testing.T) {
		var newEmployeeId = fixture.Employee("Name", "Surname", 18, time.Now(), time.Now())

		got, err := employeeRepository.FindById(newEmployeeId)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.Name)
		a.NotEmpty(got.Surname)
		a.NotEmpty(got.Age)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal("Name", got.Name)
	})
}

func TestEmployeeRepositoryWhenAdd(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()

	var clearDatabase = func() {
		db.MustExec("DELETE FROM employee")
	}
	defer func() {
		if r := recover(); r != nil {
			clearDatabase()
		} else {
			clearDatabase()
		}
	}()

	repo := employee.NewEmployeeRepository(db)
	t.Run("Add employee in transaction", func(t *testing.T) {
		tx, err := repo.BeginTr()
		a.Nil(err)
		entity := employee.Entity{
			Name:      "Alice",
			Surname:   "Wonder",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		id, err := repo.Add(tx, entity)
		a.Nil(err)
		a.True(id > 0)
		err = tx.Commit()
		a.Nil(err)

		employees, err := repo.FindAll()
		a.Nil(err)
		a.Equal(1, len(employees))
		a.Equal(entity.Name, employees[0].Name)
		a.Equal(entity.Surname, employees[0].Surname)
	})
}

func TestEmployeeRepositoryWhenFindByNameAndSurnameThenTrue(t *testing.T) {
	a := assert.New(t)
	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM employee")
	})
	repo := employee.NewEmployeeRepository(db)
	fixture := NewFixtureEmployee(repo)
	expected := employee.Entity{Name: "John", Surname: "Smith", Age: 60, CreatedAt: time.Now()}
	fixture.Employee(expected.Name, expected.Surname, expected.Age, expected.CreatedAt, expected.UpdatedAt)
	t.Run("Find employee by name and surname", func(t *testing.T) {
		tr, err := repo.BeginTr()
		a.Nil(err)
		rsl, err := repo.FindByNameAndSurname(tr, expected.Name, expected.Surname)
		a.Nil(err)
		a.True(rsl)
	})
}

func TestEmployeeRepositoryWhenFindAll(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM employee")
	})

	repo := employee.NewEmployeeRepository(db)
	fixture := NewFixtureEmployee(repo)

	fixture.Employee("John", "Smith", 18, time.Now(), time.Now())
	fixture.Employee("John", "Vi", 60, time.Now(), time.Now())

	t.Run("Find all", func(t *testing.T) {
		got, err := repo.FindAll()
		a.Nil(err)
		a.Equal(2, len(got))
	})
}

func TestEmployeeRepositoryWhenFindBySliceIds(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM employee")
	})

	repo := employee.NewEmployeeRepository(db)
	fixture := NewFixtureEmployee(repo)

	id1 := fixture.Employee("Alice", "Walker", 30, time.Now(), time.Now())
	id2 := fixture.Employee("Bob", "Johnson", 35, time.Now(), time.Now())
	fixture.Employee("Carol", "Smith", 40, time.Now(), time.Now())

	t.Run("Find by slice of IDs", func(t *testing.T) {
		ids := []int64{id1, id2}

		got, err := repo.FindBySliceIds(ids)

		a.NoError(err)
		a.Len(got, 2)
		a.ElementsMatch([]int64{got[0].Id, got[1].Id}, ids)
	})
}

func TestEmployeeRepositoryWhenDeleteById(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM employee")
	})

	repo := employee.NewEmployeeRepository(db)
	fixture := NewFixtureEmployee(repo)

	id := fixture.Employee("Deleted", "Sara", 30, time.Now(), time.Now())

	t.Run("Deleting existing employee by ID", func(t *testing.T) {
		got, err := repo.DeleteById(id)
		a.NoError(err)
		a.True(got)
	})

	t.Run("Deleting when false", func(t *testing.T) {
		deleted, err := repo.DeleteById(912384)
		a.NoError(err)
		a.False(deleted)
	})
}

func TestEmployeeRepositoryWhenDeleteByIds(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM employee")
	})

	repo := employee.NewEmployeeRepository(db)
	fixture := NewFixtureEmployee(repo)

	id1 := fixture.Employee("Deleted1", "Sara", 30, time.Now(), time.Now())
	id2 := fixture.Employee("Deleted2", "Sara", 30, time.Now(), time.Now())
	fixture.Employee("Conor", "Sara", 30, time.Now(), time.Now())
	t.Run("Deleting when correct", func(t *testing.T) {
		ids := []int64{id1, id2}
		got, err := repo.DeleteBySliceIds(ids)
		a.NoError(err)
		a.ElementsMatch(ids, got)
	})

	t.Run("Test deleted ids and finding one employee", func(t *testing.T) {
		got, err := repo.FindAll()
		a.NoError(err)
		a.Len(got, 1)
	})
}
