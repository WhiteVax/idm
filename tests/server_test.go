package tests

import (
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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func EnsureEmployeeTable(db *employee.Repository) error {
	return InitSchemaEmployee(db)
}

func TestEmployeePaginationIntegration(t *testing.T) {
	a := assert.New(t)

	cfg := common.GetConfig(".env")
	db := database.ConnectDbWithCfg(cfg)
	repo := employee.NewEmployeeRepository(db)

	if err := EnsureEmployeeTable(repo); err != nil {
		t.Fatal(err)
	}

	_, err := db.Exec(`TRUNCATE employee RESTART IDENTITY CASCADE`)
	a.Nil(err)

	for i := 1; i <= 5; i++ {
		emp := employee.Entity{
			Name:      fmt.Sprintf("Name%d", i),
			Surname:   fmt.Sprintf("Surname%d", i),
			Age:       int8(20 + i),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		tx, err := repo.BeginTr()
		a.Nil(err)
		_, err = repo.Add(tx, emp)
		a.Nil(err)
		a.Nil(tx.Commit())
	}

	// Проверка записей через репозиторий
	con, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	em, err := repo.FindAll(con)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(len(em))
	for _, e := range em {
		fmt.Println()
		fmt.Println(e)
	}

	// Создаём сервер и контроллер
	svc := employee.NewService(repo)
	server := web.NewServer()
	logger := &common.Logger{Logger: zap.NewNop()}
	ctrl := employee.NewController(server, svc, logger)
	ctrl.RegisterRoutes()

	getPage := func(pageNumber, pageSize int) (*http.Response, employee.PageResponse) {
		url := fmt.Sprintf("/api/v1/employees/page?page_number=%d&page_size=%d", pageNumber, pageSize)
		req := httptest.NewRequest("GET", url, nil)

		resp, err := server.App.Test(req)
		a.Nil(err)

		var pageResp employee.PageResponse
		if resp.StatusCode == http.StatusOK {
			err := json.NewDecoder(resp.Body).Decode(&pageResp)
			a.Nil(err)
		}
		return resp, pageResp
	}

	t.Run("Should use default values", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/page", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)

		var pageResp employee.PageResponse
		if resp.StatusCode == http.StatusOK {
			err := json.NewDecoder(resp.Body).Decode(&pageResp)
			a.Nil(err)
			a.Equal(0, pageResp.PageNum)
			a.Equal(10, pageResp.PageSize)
			a.Equal(int64(5), pageResp.Total)
			a.Len(pageResp.Result, 5)
		}
	})

	t.Run("First page with size 3", func(t *testing.T) {
		resp, pageResp := getPage(0, 3)
		a.Equal(http.StatusOK, resp.StatusCode)
		a.Len(pageResp.Result, 3)
		a.Equal(0, pageResp.PageNum)
		a.Equal(3, pageResp.PageSize)
		a.Equal(int64(5), pageResp.Total)

		req := httptest.NewRequest("GET", "/api/v1/employees/", nil)
		resp, err := server.App.Test(req, int(200*time.Second))
		a.Nil(err)
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Body)

	})

	t.Run("Second page with size 3", func(t *testing.T) {
		resp, pageResp := getPage(1, 3)
		a.Equal(http.StatusOK, resp.StatusCode)
		a.Len(pageResp.Result, 2) // Осталось 2 записи
		a.Equal(1, pageResp.PageNum)
		a.Equal(3, pageResp.PageSize)
	})

	t.Run("Third page with size 3", func(t *testing.T) {
		resp, pageResp := getPage(2, 3)
		a.Equal(http.StatusOK, resp.StatusCode)
		a.Len(pageResp.Result, 0) // Нет записей
		a.Equal(2, pageResp.PageNum)
	})

	t.Run("Should return 400 for invalid parameters", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/page?page_size=-1&page_number=0", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Should return 400 for pageSize too large", func(t *testing.T) {
		req := httptest.NewRequest(fiber.MethodGet, "/api/v1/employees/page?page_size=150&page_number=0", nil)
		resp, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

func TestServiceNilCheck(t *testing.T) {
	a := assert.New(t)

	cfg := common.GetConfig(".env")
	db := database.ConnectDbWithCfg(cfg)
	defer db.Close()

	repo := employee.NewEmployeeRepository(db)
	a.NotNil(repo, "Repository should not be nil")

	svc := employee.NewService(repo)
	a.NotNil(svc, "Service should not be nil")

	ctx := context.Background()
	req := employee.PageRequest{PageSize: 10, PageNumber: 0}

	result, err := svc.FindAllWithLimitOffset(ctx, req)

	if err != nil {
		t.Errorf("Service error: %+v", err)
	} else {
		t.Logf("Service result: %+v", result)
	}
}
