package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"idm/inner/common"
	"idm/inner/database"
	"idm/inner/employee"
	"idm/inner/web"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func SetupTestServerAdmin(t *testing.T) (*web.Server, *sqlx.DB) {
	t.Helper()

	cfg := common.GetConfig(".env")
	logger := common.NewLogger(cfg)

	db := sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	_, err := db.Exec("TRUNCATE employee RESTART IDENTITY CASCADE")
	if err != nil {
		t.Fatalf("failed to truncate table: %v", err)
	}

	server := web.NewServer()

	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmAdmin}},
	}
	auth := func(c *fiber.Ctx) error {
		c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
		return c.Next()
	}
	server.GroupApi.Use(auth)

	employeeRepo := employee.NewEmployeeRepository(db)
	employeeService := employee.NewService(employeeRepo)
	employeeHandler := employee.NewHandler(server, employeeService, logger)
	employeeHandler.RegisterRoutes()

	return server, db
}

func SetupTestServerUser(t *testing.T) *web.Server {
	t.Helper()

	cfg := common.GetConfig(".env")
	logger := common.NewLogger(cfg)
	db := sqlx.MustConnect(cfg.DbDriverName, cfg.Dsn)
	server := web.NewServer()
	claims := &web.IdmClaims{
		RealmAccess: web.RealmAccessClaims{Roles: []string{web.IdmUser}},
	}
	auth := func(c *fiber.Ctx) error {
		c.Locals(web.JwtKey, &jwt.Token{Claims: claims})
		return c.Next()
	}
	server.GroupApi.Use(auth)
	employeeRepo := employee.NewEmployeeRepository(db)
	employeeService := employee.NewService(employeeRepo)
	employeeHandler := employee.NewHandler(server, employeeService, logger)
	employeeHandler.RegisterRoutes()
	return server
}

func CreateEmployee(t *testing.T, app *web.Server, name, surname string, age int8) {
	req := employee.CreateRequest{
		Name:      name,
		Surname:   surname,
		Age:       age,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	a := assert.New(t)
	body, _ := json.Marshal(req)
	reqHTTP := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(body))
	reqHTTP.Header.Set("Content-Type", "application/json")
	resp, _ := app.App.Test(reqHTTP)
	a.Equal(http.StatusOK, resp.StatusCode)
}

func TestServiceNilCheck(t *testing.T) {
	a := assert.New(t)

	cfg := common.GetConfig(".env")
	db := database.ConnectDbWithCfg(cfg)
	defer db.Close()

	repo := employee.NewEmployeeRepository(db)
	a.NotNil(repo)

	svc := employee.NewService(repo)
	a.NotNil(svc)

	ctx := context.Background()
	req := employee.PageRequest{PageSize: 10, PageNumber: 0}

	result, err := svc.FindAllWithLimitOffset(ctx, req)

	if err != nil {
		t.Errorf("Service error: %+v", err)
	} else {
		t.Logf("Service result: %+v", result)
	}
}

func TestIntegrationAddEmployee(t *testing.T) {
	a := assert.New(t)
	server, db := SetupTestServerAdmin(t)
	defer db.Close()

	newEmployee := employee.Entity{
		Name:      "John",
		Surname:   "Doe",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	body, _ := json.Marshal(newEmployee)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/employees/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.App.Test(req, -1)
	a.NoError(err)
	a.Equal(http.StatusOK, resp.StatusCode)

	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM employee WHERE name=$1 AND surname=$2", newEmployee.Name, newEmployee.Surname)
	a.NoError(err)
	var respBody map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	data := respBody["data"].(map[string]interface{})
	a.Equal(1, int(data["id"].(float64)))
	a.Equal("John", data["name"])
}

func TestEmployeePagination(t *testing.T) {
	a := assert.New(t)
	appAdmin, _ := SetupTestServerAdmin(t)

	for i := 1; i <= 5; i++ {
		CreateEmployee(t, appAdmin, fmt.Sprintf("Name_%d", i), fmt.Sprintf("Surname_%d", i), 20+int8(i))
	}

	addUser := SetupTestServerUser(t)

	t.Run("First page with 3 entries - 3", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet,
			"/api/v1/employees/page?page_number=0&page_size=3&text_filter=name_", nil)
		resp, _ := addUser.App.Test(req)

		a.Equal(http.StatusOK, resp.StatusCode)

		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)

		fmt.Println(pageResp.Data.Result)

		a.Len(pageResp.Data.Result, 3)
		a.Equal(int64(5), pageResp.Data.Total)
		a.True(pageResp.Success)
	})

	t.Run("Second page with 2 entries - 2", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?page_number=1&page_size=3", nil)
		resp, _ := addUser.App.Test(req)

		a.Equal(http.StatusOK, resp.StatusCode)

		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)
		a.Len(pageResp.Data.Result, 2)
	})

	t.Run("Third page with 3 entries - 0", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?page_number=2&page_size=3", nil)
		resp, _ := addUser.App.Test(req)

		a.Equal(http.StatusOK, resp.StatusCode)

		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)
		a.Len(pageResp.Data.Result, 0)
	})
	t.Run("Invalid web request", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?page_number=abc&page_size=xyz", nil)
		resp, _ := addUser.App.Test(req)
		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)

		a.Equal(http.StatusBadRequest, resp.StatusCode)
		a.False(pageResp.Success)
		a.Nil(pageResp.Data.Result)
	})

	t.Run("Without instructions Page_number", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?page_size=3", nil)
		resp, _ := addUser.App.Test(req)

		a.Equal(http.StatusOK, resp.StatusCode)

		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)
		a.Len(pageResp.Data.Result, 3)
	})

	t.Run("Without instructions PageSize", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/page?page_number=0", nil)
		resp, _ := addUser.App.Test(req)
		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)

		a.Equal(http.StatusBadRequest, resp.StatusCode)
		a.False(pageResp.Success)
	})

	t.Run("With text filter witch not found in database", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet,
			"/api/v1/employees/page?page_number=0&page_size=3&text_filter=super_", nil)
		resp, _ := addUser.App.Test(req)
		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)

		a.Equal(http.StatusOK, resp.StatusCode)
		a.True(pageResp.Success)
		a.Equal("super_", pageResp.Data.TextFilter)
		a.Equal(int64(0), pageResp.Data.Total)
		a.Equal([]employee.Response{}, pageResp.Data.Result)
	})

	t.Run("With text filter witch less string len 3", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet,
			"/api/v1/employees/page?page_number=0&page_size=5&text_filter=na", nil)
		resp, _ := addUser.App.Test(req)
		var pageResp employee.EntityPageResponse
		_ = json.NewDecoder(resp.Body).Decode(&pageResp)

		a.Equal(http.StatusOK, resp.StatusCode)
		a.True(pageResp.Success)
		a.Equal("na", pageResp.Data.TextFilter)
		a.Equal(int64(5), pageResp.Data.Total)
		a.Equal(5, len(pageResp.Data.Result))
	})
}
