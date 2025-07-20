package employee

import (
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
)

type Controller struct {
	server          *web.Server
	employeeService Svc
}

type Svc interface {
	Add(employee Entity) (response Response, err error)
	FindById(id int64) (Response, error)
	CreateEmployee(request CreateRequest) (int64, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) (Response, error)
	FindAll() (employees []Response, err error)
}

func (c *Controller) RegisterRoutes() {
	// полный маршрут получится "/api/v1/employees"
	c.server.GroupApiV1.Post("/employees", c.CreateEmployee)
	c.server.GroupApiV1.Post("/employees/add", c.AddEmployee)
	c.server.GroupApiV1.Post("/employees/ids", c.FindByIds)
	c.server.GroupApiV1.Post("/employees/:id", c.FindById)
	c.server.GroupApiV1.Delete("/employees/ids", c.DeleteByIds)
	c.server.GroupApiV1.Delete("/employees/:id", c.DeleteById)
	c.server.GroupApiV1.Get("/employees", c.FindAll)
}

func (c *Controller) CreateEmployee(ctx *fiber.Ctx) error {
	var request CreateRequest
	if err := ctx.BodyParser(&request); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	var newEmployeeId, err = c.employeeService.CreateEmployee(request)
	if err != nil {
		switch {
		case errors.As(err, &common.RequestValidationError{}) || errors.As(err, &common.AlreadyExistsError{}):
			return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
		default:
			return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
		}
	}
	if err = common.OkResponse(ctx, newEmployeeId); err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error returning created employee id")
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newEmployeeId)
}

func (c *Controller) AddEmployee(ctx *fiber.Ctx) error {
	var entity Entity
	if err := ctx.BodyParser(&entity); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	var newEmployeeId, err = c.employeeService.Add(entity)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newEmployeeId)
}

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	employee, err := c.employeeService.FindById(id)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employee)
}

func (c *Controller) FindByIds(ctx *fiber.Ctx) error {
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	employees, err := c.employeeService.FindByIds(ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error finding employees")
	}
	return common.OkResponse(ctx, employees)
}

func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	rsl, err := c.employeeService.DeleteById(id)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

func (c *Controller) DeleteByIds(ctx *fiber.Ctx) error {
	bodyBytes := ctx.Body()
	var ids []int64
	if err := json.Unmarshal(bodyBytes, &ids); err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	rsl, err := c.employeeService.DeleteByIds(ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	employees, err := c.employeeService.FindAll()
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employees)
}
