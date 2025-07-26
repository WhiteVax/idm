package info

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockDb struct {
	mock.Mock
}

func (m *MockDb) Ping() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDb) Status(code int) *MockDb {
	m.Called(code)
	return m
}

func (m *MockDb) JSON(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestGetInfo(t *testing.T) {
	a := assert.New(t)
	t.Run("Get Info with status 200", func(t *testing.T) {
		app := fiber.New()
		server := &web.Server{
			App:           app,
			GroupInternal: app.Group("/internal"),
		}
		cfg := common.Config{
			AppName:    "test",
			AppVersion: "1.0.0",
		}
		ctrl := NewController(server, cfg, nil)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/info", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(200, resp.StatusCode)
	})
}

func TestGetHealth(t *testing.T) {
	a := assert.New(t)
	t.Run("Get health with status 200", func(t *testing.T) {
		app := fiber.New()
		server := &web.Server{
			App:           app,
			GroupInternal: app.Group("/internal"),
		}
		cfg := common.Config{
			AppName:    "test",
			AppVersion: "1.0.0",
		}
		mockDb, _, err := sqlmock.New()
		a.Nil(err)
		ctrl := NewController(server, cfg, mockDb)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/health", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Get health with status 500", func(t *testing.T) {
		app := fiber.New()
		server := &web.Server{
			App:           app,
			GroupInternal: app.Group("/internal"),
		}
		cfg := common.Config{
			AppName:    "test",
			AppVersion: "1.0.0",
		}
		mockDb, _, err := sqlmock.New()
		a.Nil(err)
		mockDb.Close()
		ctrl := NewController(server, cfg, mockDb)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/health", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}
