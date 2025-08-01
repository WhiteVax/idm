package employee

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"idm/inner/common"
	"idm/inner/web"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func (svc *MockService) FindAll() (employees []Response, err error) {
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

func TestCreateEmployee(t *testing.T) {
	a := assert.New(t)

	t.Run("Should return created employee id", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
		body := strings.NewReader(`{invalid json}`)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 400 on AlreadyExistsError", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
	t.Run("Should add employee and get id with status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
		body := strings.NewReader("")
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/add", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 500 on internal error", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

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
	t.Run("When delete status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("DeleteById", int64(2)).Return(Response{Id: 2}, nil)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/2", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("DeleteById", int64(0)).Return(Response{Id: 0}, nil)
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/abc", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("DeleteById", int64(0)).Return(Response{Id: 0}, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestDeleteByIdsEmployees(t *testing.T) {
	a := assert.New(t)
	t.Run("When delete with status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

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
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
		// пустое тело передаётся, дальше методы сервиса не вызываются
		req := httptest.NewRequest(fiber.MethodDelete, "/api/v1/employees/ids", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
		svc.AssertExpectations(t)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
	t.Run("When find by ids with status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

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
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/ids", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()
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
	t.Run("When find by id status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("FindById", int64(2)).Return(Response{Id: 2}, nil)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/2", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When fail error 400", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("FindById", int64(0)).Return(Response{Id: 0}, nil)
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/a", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("When fail error 500", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		svc.On("FindById", int64(0)).Return(Response{Id: 0}, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodPost, "/api/v1/employees/0", nil)
		resp, err := server.App.Test(req, -1)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestFindAllEmployees(t *testing.T) {
	a := assert.New(t)

	t.Run("When find all employees status 200", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		svc.On("FindAll").Return(expected, nil)
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees", nil)

		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("When find all employees status 500", func(t *testing.T) {
		server := web.NewServer()
		svc := new(MockService)
		controller := Controller{server: server, employeeService: svc}
		controller.RegisterRoutes()

		expected := []Response{{Id: 1}, {Id: 2}}
		svc.On("FindAll").Return(expected, errors.New("db failure"))
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees", nil)

		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusInternalServerError, resp.StatusCode)
	})
}
