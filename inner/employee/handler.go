package employee

import (
	"context"
	"encoding/json"
	"errors"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	Server          *web.Server
	employeeService Svc
	logger          *common.Logger
}

type Svc interface {
	Add(employee Entity) (response Response, err error)
	FindById(id int64) (Response, error)
	CreateEmployee(request CreateRequest) (int64, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) (Response, error)
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
func (c *Handler) CreateEmployee(ctx *fiber.Ctx) error {
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		c.logger.Error("CreateEmployee: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("CreateEmployee: receive request", zap.Any("request", request))
	var newEmployeeId, err = c.employeeService.CreateEmployee(request)
	if err != nil {
		c.logger.Error("CreateEmployee: error creating", zap.Error(err))
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
func (c *Handler) AddEmployee(ctx *fiber.Ctx) error {
	var entity Entity
	if err := ctx.BodyParser(&entity); err != nil {
		c.logger.Error("AddEmployee: : error body parse", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	c.logger.Debug("AddEmployee: receive entity", zap.Any("entity", entity))
	var newEmployeeId, err = c.employeeService.Add(entity)
	if err != nil {
		c.logger.Error("AddEmployee: error adding", zap.Error(err))
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
func (c *Handler) FindById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("FindById: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("FindById: receive idParam", zap.Any("idParam", idParam))
	employee, err := c.employeeService.FindById(id)
	if err != nil {
		c.logger.Error("FindById: error finding", zap.Error(err))
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
func (c *Handler) FindByIds(ctx *fiber.Ctx) error {
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		c.logger.Error("FindByIds: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.Debug("FindByIds: receive ids", zap.Any("ids", ids))
	employees, err := c.employeeService.FindByIds(ids)
	if err != nil {
		c.logger.Error("FindByIds: error finding", zap.Error(err))
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
func (c *Handler) DeleteById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("DeleteById: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("DeleteById: receive idParam", zap.Any("idParam", idParam))
	rsl, err := c.employeeService.DeleteById(id)
	if err != nil {
		c.logger.Error("DeleteById: error deleting", zap.Error(err))
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
func (c *Handler) DeleteByIds(ctx *fiber.Ctx) error {
	bodyBytes := ctx.Body()
	var ids []int64
	if err := json.Unmarshal(bodyBytes, &ids); err != nil {
		c.logger.Error("DeleteByIds: : error body parse", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.Debug("DeleteByIds: receive ids", zap.Any("ids", ids))
	rsl, err := c.employeeService.DeleteByIds(ids)
	if err != nil {
		c.logger.Error("DeleteByIds: error deleting", zap.Error(err))
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
func (c *Handler) FindAll(ctx *fiber.Ctx) error {
	con, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	employees, err := c.employeeService.FindAll(con)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			c.logger.Error("FindAll: request timeout", zap.Error(err))
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"Error": "request timeout"})
		}
		c.logger.Error("FindAll: error finding", zap.Error(err))
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
// @Router /employees/page [get]
func (c *Handler) FindByPagesWithFilter(ctx *fiber.Ctx) error {
	var request PageRequest
	if err := ctx.QueryParser(&request); err != nil {
		c.logger.Error("FindByPagesWithFilter: query parse error", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid query parameters")
	}
	c.logger.Debug("FindByPagesWithFilter: received page request", zap.Any("request", request))

	con, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	employees, err := c.employeeService.FindAllWithLimitOffset(con, request)
	if err != nil {
		var validationErr common.RequestValidationError
		if ok := errors.As(err, &validationErr); ok {
			c.logger.Error("FindByPagesWithFilter: validation error", zap.Error(err))
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		}
		if errors.Is(err, context.DeadlineExceeded) {
			c.logger.Error("FindByPagesWithFilter: request timeout", zap.Error(err))
			return ctx.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{"error": "request timeout"})
		}
		c.logger.Error("FindByPagesWithFilter: internal error", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employees)
}
