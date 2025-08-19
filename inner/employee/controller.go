package employee

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
	"time"
)

type Controller struct {
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
}

func NewController(server *web.Server, employeeService Svc, logger *common.Logger) *Controller {
	return &Controller{
		Server:          server,
		employeeService: employeeService,
		logger:          logger,
	}
}

func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/employees"
	c.Server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.Server.GroupApiV1.Post("/employees/add", c.AddEmployee)
	c.Server.GroupApiV1.Post("/employees/ids", c.FindByIds)
	c.Server.GroupApiV1.Post("/employees/:id", c.FindById)
	c.Server.GroupApiV1.Delete("/employees/ids", c.DeleteByIds)
	c.Server.GroupApiV1.Delete("/employees/:id", c.DeleteById)
	c.Server.GroupApiV1.Get("/employees", c.FindAll)
}

func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {
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

func (c *Controller) AddEmployee(ctx *fiber.Ctx) error {
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

func (c *Controller) FindById(ctx *fiber.Ctx) error {
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

func (c *Controller) FindByIds(ctx *fiber.Ctx) error {
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

func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
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

func (c *Controller) DeleteByIds(ctx *fiber.Ctx) error {
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

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
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
