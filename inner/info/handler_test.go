package info

import (
	"idm/inner/common"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
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
		ctrl := NewHandler(server, cfg, nil, logger)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/info", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(200, resp.StatusCode)
	})
}

func TestGetHealth(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("Get health with status 200", func(t *testing.T) {
		t.Parallel()
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
		sqlxDB := sqlx.NewDb(mockDb, "sqlmock")
		ctrl := NewHandler(server, cfg, sqlxDB, logger)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/health", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Get health with status 500", func(t *testing.T) {
		t.Parallel()
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
		sqlxDB := sqlx.NewDb(mockDb, "sqlmock")
		ctrl := NewHandler(server, cfg, sqlxDB, logger)
		ctrl.RegisterRoutes()
		req := httptest.NewRequest(http.MethodGet, "/internal/health", nil)
		resp, err := app.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusServiceUnavailable, resp.StatusCode)
	})
}
