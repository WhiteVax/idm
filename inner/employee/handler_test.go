package employee

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"idm/inner/common"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (svc *MockService) Add(employee Entity) (response Response, err error) {
	args := svc.Called(employee)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) FindByIds(ids []int64) ([]Response, error) {
	args := svc.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) DeleteByIds(ids []int64) ([]Response, error) {
	args := svc.Called(ids)
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) DeleteById(id int64) (Response, error) {
	args := svc.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) FindAll(context.Context) (employees []Response, err error) {
	args := svc.Mock.Called()
	return args.Get(0).([]Response), args.Error(1)
}

func (svc *MockService) FindById(id int64) (Response, error) {
	args := svc.Called(id)
	return args.Get(0).(Response), args.Error(1)
}

func (svc *MockService) CreateEmployee(request CreateRequest) (int64, error) {
	args := svc.Called(request)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockService) FindAllWithLimitOffset(ctx context.Context, req PageRequest) (PageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(PageResponse), args.Error(1)
}

func TestCreateEmployee(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("Should return created employee id", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		now := time.Now().UTC().Format(time.RFC3339)

		body := strings.NewReader(fmt.Sprintf(`{
			"name": "John",
			"surname": "Doe",
			"age": 25,
			"created_at": "%s",
			"updated_at": "%s"
		}`, now, now))
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		svc.On("CreateEmployee", mock.AnythingOfType("employee.CreateRequest")).Return(int64(123), nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotNil(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		bytesData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[int64]
		err = json.Unmarshal(bytesData, &responseBody)
		a.Nil(err)
		a.True(responseBody.Success)
		a.Equal(int64(123), responseBody.Data)
		a.Empty(responseBody.Message)
		svc.AssertExpectations(t)
	})

	t.Run("Should return 400 on bad JSON", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		body := strings.NewReader(`{invalid json}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 400 on AlreadyExistsError", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		now := time.Now().UTC().Format(time.RFC3339)
		body := strings.NewReader(fmt.Sprintf(`{
			"name": "John",
			"surname": "Doe",
			"age": 25,
			"created_at": "%s",
			"updated_at": "%s"
		}`, now, now))
		svc.On("CreateEmployee", mock.Anything).Return(int64(0), common.AlreadyExistsError{Message: "employee already exists"})
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 500 on unknown internal error", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		now := time.Now().UTC().Format(time.RFC3339)
		body := strings.NewReader(fmt.Sprintf(`{
			"name": "John",
			"surname": "Doe",
			"age": 25,
			"created_at": "%s",
			"updated_at": "%s"
		}`, now, now))
		svc.On("CreateEmployee", mock.Anything).Return(int64(0), errors.New("db connection error"))
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAddEmployee(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("Should add employee and get id with status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		now := time.Now().UTC()
		entity := Entity{
			Id:        1,
			Name:      "John",
			Surname:   "Doe",
			Age:       25,
			CreatedAt: now,
			UpdatedAt: now}
		body := strings.NewReader(fmt.Sprintf(`{
			"name": "%s",
			"surname": "%s",
			"age": %d,
			"created_at": "%s",
			"updated_at": "%s"
		}`, entity.Name, entity.Surname, entity.Age, now.Format(time.RFC3339), now.Format(time.RFC3339)))
		svc.On("Add", mock.Anything).Return(Response{Id: 1}, nil)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.NotNil(resp)
		a.Equal(http.StatusOK, resp.StatusCode)
		respData, err := io.ReadAll(resp.Body)
		a.Nil(err)
		var responseBody common.Response[Response]
		err = json.Unmarshal(respData, &responseBody)
		a.Nil(err)
		a.Equal(responseBody.Data.Id, entity.Id)
		a.True(responseBody.Success)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		body := strings.NewReader("")
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 500 on internal error", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		now := time.Now().UTC()
		entity := Entity{
			Id:        1,
			Name:      "John",
			Surname:   "Doe",
			Age:       25,
			CreatedAt: now,
			UpdatedAt: now}

		body := strings.NewReader(fmt.Sprintf(`{
			"name": "%s",
			"surname": "%s",
			"age": %d,
			"created_at": "%s",
			"updated_at": "%s"
		}`, entity.Name, entity.Surname, entity.Age, now.Format(time.RFC3339), now.Format(time.RFC3339)))
		svc.On("Add", mock.Anything).Return(Response{}, common.RequestValidationError{Message: "validation failed"})
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestDeleteByIdEmployee(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("When delete status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("DeleteById", int64(2)).Return(Response{Id: 2}, nil)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/2", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("DeleteById", int64(0)).Return(Response{Id: 0}, nil)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/abc", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("DeleteById", int64(0)).Return(Response{Id: 0}, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestDeleteByIdsEmployees(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("When delete with status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		input := []int64{1, 2}
		svc.On("DeleteByIds", input).Return(expected, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/ids", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		resp, err := server.App.Test(req)
		a.Nil(err)

		a.Equal(fiber.StatusOK, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/ids", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		expected := []Response{{Id: 0}}
		input := []int64{1, 2}
		svc.On("DeleteByIds", input).Return(expected, errors.New("service error")).Once()
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/ids", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}

func TestFindByIdsEmployees(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("When find by ids with status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		input := []int64{1, 2}
		svc.On("FindByIds", input).Return(expected, nil).Once()

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/ids", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		resp, err := server.App.Test(req)
		a.Nil(err)

		a.Equal(fiber.StatusOK, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/ids", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()
		expected := []Response{{Id: 0}}
		input := []int64{1, 2}
		svc.On("FindByIds", input).Return(expected, errors.New("service error")).Once()
		body, _ := json.Marshal(input)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/ids", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}

func TestFindByIdEmployee(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("When find by id status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("FindById", int64(2)).Return(Response{Id: 2}, nil)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/2", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("FindById", int64(0)).Return(Response{Id: 0}, nil)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/a", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		svc.On("FindById", int64(0)).Return(Response{Id: 0}, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestFindAllEmployees(t *testing.T) {
	a := assert.New(t)
	logger := &common.Logger{
		Logger: zap.NewNop(),
	}
	t.Run("When find all employees status 200", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		svc.On("FindAll").Return(expected, nil)
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees", nil)

		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When find all employees status 500", func(t *testing.T) {
		t.Parallel()
		server := web.NewServer()
		svc := new(MockService)
		handler := Handler{Server: server, employeeService: svc, logger: logger}
		handler.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		svc.On("FindAll").Return(expected, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees", nil)

		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestFindAllEmployeesWithLimitOffset(t *testing.T) {

	t.Run("Should return 200 OK with valid request", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		handler := NewHandler(server, svc, &common.Logger{Logger: zap.NewNop()})
		handler.RegisterRoutes()
		expectedResponse := PageResponse{
			Result:     []Response{{Id: 1, Name: "John"}},
			TextFilter: "",
			PageSize:   10,
			PageNum:    0,
			Total:      100,
		}
		svc.On("FindAllWithLimitOffset",
			mock.Anything,
			mock.MatchedBy(func(req PageRequest) bool {
				return req.PageNumber == 0 && req.PageSize == 10
			}),
		).Return(expectedResponse, nil)
		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/employees/page?page_number=0&page_size=10",
			nil,
		)
		resp, err := server.App.Test(req, -1) // -1 — без таймаута
		a.Nil(err)
		defer resp.Body.Close()
		a.Equal(fiber.StatusOK, resp.StatusCode)
		var actualResponse PageResponse
		err = json.NewDecoder(resp.Body).Decode(&actualResponse)
		a.Nil(err)
		svc.AssertExpectations(t)
	})

	t.Run("Should return 400 BadRequest on invalid body", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		handler := NewHandler(server, svc, &common.Logger{Logger: zap.NewNop()})
		handler.RegisterRoutes()

		body := `{"page_size":"abc","page_number":1}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/page", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		a.Nil(err)
		defer resp.Body.Close()

		a.Equal(fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 400 BadRequest on validation error", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		server := web.NewServer()
		svc := new(MockService)
		handler := NewHandler(server, svc, &common.Logger{Logger: zap.NewNop()})
		handler.RegisterRoutes()

		body := `{"page_size":3,"page_number":-1}`
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/page", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := server.App.Test(req)
		a.Nil(err)
		defer resp.Body.Close()

		a.Equal(fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 500 InternalServerError on service error", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		server := web.NewServer()
		svc := new(MockService)
		handler := NewHandler(server, svc, &common.Logger{Logger: zap.NewNop()})
		handler.RegisterRoutes()

		svc.On("FindAllWithLimitOffset",
			mock.Anything,
			mock.MatchedBy(func(req PageRequest) bool {
				return req.PageNumber == 1 && req.PageSize == 3
			}),
		).Return(PageResponse{}, errors.New("database connection failed"))

		req := httptest.NewRequest(
			fiber.MethodGet,
			"/api/v1/employees/page?page_number=1&page_size=3",
			nil,
		)

		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		defer resp.Body.Close()

		a.Equal(fiber.StatusInternalServerError, resp.StatusCode)

		svc.AssertExpectations(t)
	})

	t.Run("Should return 408 RequestTimeout on context deadline exceeded", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)

		server := web.NewServer()
		svc := new(MockService)
		handler := NewHandler(server, svc, &common.Logger{Logger: zap.NewNop()})
		handler.RegisterRoutes()

		svc.On("FindAllWithLimitOffset",
			mock.Anything,
			mock.MatchedBy(func(req PageRequest) bool {
				return req.PageNumber == 1 && req.PageSize == 3
			}),
		).Return(PageResponse{}, context.DeadlineExceeded)

		req := httptest.NewRequest(
			fiber.MethodGet,
			"/api/v1/employees/page?page_number=1&page_size=3",
			nil,
		)

		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		defer resp.Body.Close()

		a.Equal(fiber.StatusRequestTimeout, resp.StatusCode)
		svc.AssertExpectations(t)
	})
}
