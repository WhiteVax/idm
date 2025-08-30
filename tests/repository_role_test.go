package tests

import (
	"idm/inner/database"
	"idm/inner/role"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryRole(t *testing.T) {
	a := assert.New(t)

	var db = database.ConnectDb()
	var clearDatabase = func() {
		db.MustExec("DELETE FROM role")
	}
	defer func() {
		if r := recover(); r != nil {
			clearDatabase()
		} else {
			clearDatabase()
		}
	}()
	var roleRepository = role.NewRepository(db)
	var fixture = NewFixtureRole(roleRepository)

	t.Run("Find an role by id", func(t *testing.T) {
		var newRoleId = fixture.Role("Role", time.Now(), time.Now())

		got, err := roleRepository.FindById(newRoleId)

		a.Nil(err)
		a.NotEmpty(got)
		a.NotEmpty(got.Id)
		a.NotEmpty(got.Name)
		a.NotEmpty(got.CreatedAt)
		a.NotEmpty(got.UpdatedAt)
		a.Equal("Role", got.Name)
	})
}

func TestRoleRepositoryWhenFindAll(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM role")
	})

	repo := role.NewRepository(db)
	fixture := NewFixtureRole(repo)

	fixture.Role("Guest", time.Now(), time.Now())
	fixture.Role("Admin", time.Now(), time.Now())

	t.Run("Find all", func(t *testing.T) {
		got, err := repo.FindAll()
		a.Nil(err)
		a.Equal(2, len(got))
	})
}

func TestRoleRepositoryWhenFindBySliceIds(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM role")
	})

	repo := role.NewRepository(db)
	fixture := NewFixtureRole(repo)

	id1 := fixture.Role("Guest", time.Now(), time.Now())
	id2 := fixture.Role("Guest1", time.Now(), time.Now())
	fixture.Role("Admin", time.Now(), time.Now())

	t.Run("Find by slice of IDs", func(t *testing.T) {
		ids := []int64{id1, id2}

		got, err := repo.FindBySliceIds(ids)

		a.NoError(err)
		a.Len(got, 2)
		a.ElementsMatch([]int64{got[0].Id, got[1].Id}, ids)
	})
}

func TestRoleRepositoryWhenDeleteById(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM role")
	})

	repo := role.NewRepository(db)
	fixture := NewFixtureRole(repo)

	id := fixture.Role("Guest", time.Now(), time.Now())

	t.Run("Deleting existing role by ID", func(t *testing.T) {
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

func TestRoleRepositoryWhenDeleteByIds(t *testing.T) {
	a := assert.New(t)

	db := database.ConnectDb()
	t.Cleanup(func() {
		db.MustExec("DELETE FROM role")
	})

	repo := role.NewRepository(db)
	fixture := NewFixtureRole(repo)

	id1 := fixture.Role("Guest1", time.Now(), time.Now())
	id2 := fixture.Role("Guest2", time.Now(), time.Now())
	fixture.Role("Guest3", time.Now(), time.Now())
	t.Run("Deleting when correct", func(t *testing.T) {
		ids := []int64{id1, id2}
		got, err := repo.DeleteBySliceIds(ids)
		a.NoError(err)
		a.ElementsMatch(ids, got)
	})

	t.Run("Test deleted ids and finding one role", func(t *testing.T) {
		got, err := repo.FindAll()
		a.NoError(err)
		a.Len(got, 1)
	})
}
