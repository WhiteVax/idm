package employee

import (
	"context"
	"encoding/json"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"slices"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type Handler struct {
	Server          *web.Server
	employeeService Svc
	logger          *common.Logger
}

type Svc interface {
	Add(ctx context.Context, employee Entity) (response Response, err error)
	FindById(ctx context.Context, id int64) (Response, error)
	CreateEmployee(ctx context.Context, request CreateRequest) (int64, error)
	FindByIds(ctx context.Context, ids []int64) ([]Response, error)
	DeleteByIds(ctx context.Context, ids []int64) ([]Response, error)
	DeleteById(ctx context.Context, id int64) (Response, error)
	FindAll(ctx context.Context) (employees []Response, err error)
	FindAllWithLimitOffset(ctx context.Context, req PageRequest) (result PageResponse, err error)
}

func NewHandler(server *web.Server, employeeService Svc, logger *common.Logger) *Handler {
	return &Handler{
		Server:          server,
		employeeService: employeeService,
		logger:          logger,
	}
}

// RegisterRoutes - регистрация маршрута "/api/v1/employees"
func (c *Handler) RegisterRoutes() {
	c.Server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.Server.GroupApiV1.Post("/employees/add", c.AddEmployee)
	c.Server.GroupApiV1.Post("/employees/ids", c.FindByIds)
	c.Server.GroupApiV1.Post("/employees/:id", c.FindById)
	c.Server.GroupApiV1.Delete("/employees/ids", c.DeleteByIds)
	c.Server.GroupApiV1.Delete("/employees/:id", c.DeleteById)
	c.Server.GroupApiV1.Get("/employees", c.FindAll)
	c.Server.GroupApiV1.Get("/employees/page", c.FindByPagesWithFilter)
}

// Функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees"
// @Description Create a new employee.
// @Summary create a new employee
// @Tags employee
// @Accept json
// @Produce json
// @Param request body CreateRequest true "create employee request"
// @Success 200 {object} common.Response[employee.Entity]
// @Failure 400 {object} common.Response[employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[employee.Entity] "error db"
// @Router /employees [post]
// @Security BearerAuth
func (c *Handler) CreateEmployee(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.ErrorCtx(ctx.Context(), "CreateEmployee: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.DebugCtx(ctx.Context(), "Create employee: received request", zap.Any("request", request))
	var newEmployeeId, err = c.employeeService.CreateEmployee(ctx.Context(), request)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "CreateEmployee: error creating", zap.Error(err))
		switch {
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	return common.OkResponse(ctx, newEmployeeId)
}

// Функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees/add"
// @Description Create a new employee with secure transaction.
// @Summary create a new employee with transaction
// @Tags employee
// @Accept json
// @Produce json
// @Param request body CreateRequest true "create employee request"
// @Success 200 {object} common.Response[employee.Entity]
// @Failure 400 {object} common.Response[employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[employee.Entity] "error db"
// @Router /employees/add [post]
// @Security BearerAuth
func (c *Handler) AddEmployee(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	var entity Entity
	if err := ctx.BodyParser(&entity); err != nil {
		c.logger.ErrorCtx(ctx.Context(), "AddEmployee: : error body parse", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	c.logger.DebugCtx(ctx.Context(), "AddEmployee: receive entity", zap.Any("entity", entity))
	var newEmployeeId, err = c.employeeService.Add(ctx.Context(), entity)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "AddEmployee: error adding", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newEmployeeId)
}

// Функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees/:id"
// @Description Find employee by id.
// @Summary find employee
// @Tags employee
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} common.Response[employee.Entity]
// @Failure 400 {object} common.Response[employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[employee.Entity] "error db"
// @Router /employees/{id} [post]
// @Security BearerAuth
func (c *Handler) FindById(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "FindById: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.DebugCtx(ctx.Context(), "FindById: receive idParam", zap.Any("idParam", idParam))
	employee, err := c.employeeService.FindById(ctx.Context(), id)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "FindById: error finding", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employee)
}

// Функция-хендлер, которая будет вызываться при POST запросе по маршруту "/api/v1/employees/ids"
// @Description Find employees by ids.
// @Summary find employees
// @Tags employee
// @Accept json
// @Produce json
// @Param request body []int64 true "Employee IDs"
// @Success 200 {object} common.Response[[]employee.Entity]
// @Failure 400 {object} common.Response[[]employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[[]employee.Entity] "error db"
// @Router /employees/ids [post]
// @Security BearerAuth
func (c *Handler) FindByIds(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		c.logger.ErrorCtx(ctx.Context(), "FindByIds: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.DebugCtx(ctx.Context(), "FindByIds: receive ids", zap.Any("ids", ids))
	employees, err := c.employeeService.FindByIds(ctx.Context(), ids)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "FindByIds: error finding", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error finding employees")
	}
	return common.OkResponse(ctx, employees)
}

// Функция-хендлер, которая будет вызываться при DELETE запросе по маршруту "/api/v1/employees/:id"
// @Description Delete employee by id.
// @Summary delete employee
// @Tags employee
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} common.Response[employee.Entity]
// @Failure 400 {object} common.Response[employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[employee.Entity] "error db"
// @Router /employees/{id} [delete]
// @Security BearerAuth
func (c *Handler) DeleteById(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "DeleteById: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.DebugCtx(ctx.Context(), "DeleteById: receive idParam", zap.Any("idParam", idParam))
	rsl, err := c.employeeService.DeleteById(ctx.Context(), id)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "DeleteById: error deleting", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

// Функция-хендлер, которая будет вызываться при DELETE запросе по маршруту "/api/v1/employees/ids"
// @Description Delete employees by ids.
// @Summary delete employees
// @Tags employee
// @Accept json
// @Produce json
// @Param request body []int64 true "Employee IDs"
// @Success 200 {object} common.Response[[]employee.Entity]
// @Failure 400 {object} common.Response[[]employee.Entity] "invalid request"
// @Failure 500 {object} common.Response[[]employee.Entity] "error db"
// @Router /employees/ids [delete]
// @Security BearerAuth
func (c *Handler) DeleteByIds(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmAdmin) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	bodyBytes := ctx.Body()
	var ids []int64
	if err := json.Unmarshal(bodyBytes, &ids); err != nil {
		c.logger.ErrorCtx(ctx.Context(), "DeleteByIds: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.DebugCtx(ctx.Context(), "DeleteByIds: receive ids", zap.Any("ids", ids))
	rsl, err := c.employeeService.DeleteByIds(ctx.Context(), ids)
	if err != nil {
		c.logger.ErrorCtx(ctx.Context(), "DeleteByIds: error deleting", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

// Функция-хендлер, которая будет вызываться при GET запросе по маршруту "/api/v1/employees"
// @Description Get all employees.
// @Summary get employees
// @Tags employee
// @Accept json
// @Produce json
// @Success 200 {object} common.Response[employee.Entity]
// @Failure 400 {object} map[string]string "invalid request"
// @Failure 500 {object} map[string]string "error db"
// @Router /employees [get]
// @Security BearerAuth
func (c *Handler) FindAll(ctx *fiber.Ctx) error {
	var token = ctx.Locals(web.JwtKey).(*jwt.Token)
	var claims = token.Claims.(*web.IdmClaims)
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	con, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	employees, err := c.employeeService.FindAll(con)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			c.logger.ErrorCtx(ctx.Context(), "FindAll: request timeout", zap.Error(err))
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"Error": "request timeout"})
		}
		c.logger.ErrorCtx(ctx.Context(), "FindAll: error finding", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employees)
}

// Функция-хендлер, которая будет вызываться при GET запросе по маршруту "/api/v1/employees/page?page_number=0&page_size=3&text_filter=name_"
// @Description Find employees by name within limit and offsett.
// @Summary find employees by conditions
// @Tags employee
// @Accept json
// @Produce json
// @Param page_number query int false "Page number"
// @Param page_size query int false "Page size"
// @Param text_filter query string false "Text filter"
// @Success 200 {object} PageResponse[]
// @Failure 400 {object} PageResponse[]
// @Failure 408 {object} PageResponse[] "time out request"
// @Failure 500 {object} PageResponse[] "db error"
// @Failure 401 {object} PageResponse[] "Unauthorized or token expired"
// @Failure 403 {object} PageResponse[] "Permission denied"
// @Router /employees/page [get]
// @Security BearerAuth
func (c *Handler) FindByPagesWithFilter(ctx *fiber.Ctx) error {
	jwtToken, ok := ctx.Locals(web.JwtKey).(*jwt.Token)
	if !ok || jwtToken == nil {
		return common.ErrResponse(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}
	claims, ok := jwtToken.Claims.(*web.IdmClaims)
	if !ok || claims == nil {
		return common.ErrResponse(ctx, fiber.StatusUnauthorized, "Unauthorized")
	}
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return common.ErrResponse(ctx, fiber.StatusUnauthorized, "Token expired")
	}
	if !slices.Contains(claims.RealmAccess.Roles, web.IdmUser) {
		return common.ErrResponse(ctx, fiber.StatusForbidden, "Permission denied")
	}
	var request PageRequest
	if err := ctx.QueryParser(&request); err != nil {
		c.logger.ErrorCtx(ctx.Context(), "FindByPagesWithFilter: query parse error", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid query parameters")
	}
	c.logger.DebugCtx(ctx.Context(), "FindByPagesWithFilter: received page request", zap.Any("request", request))

	con, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	employees, err := c.employeeService.FindAllWithLimitOffset(con, request)
	if err != nil {
		var validationErr common.RequestValidationError
		if ok := errors.As(err, &validationErr); ok {
			c.logger.ErrorCtx(ctx.Context(), "FindByPagesWithFilter: validation error", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		if errors.Is(err, context.DeadlineExceeded) {
			c.logger.ErrorCtx(ctx.Context(), "FindByPagesWithFilter: request timeout", zap.Error(err))
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"error": "request timeout"})
		}
		c.logger.ErrorCtx(ctx.Context(), "FindByPagesWithFilter: internal error", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employees)
}
