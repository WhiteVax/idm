package role

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"
)

type Controller struct {
	server  *web.Server
	service Svc
}

type Svc interface {
	Add(role Entity) (Response, error)
	FindById(id int64) (Response, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) (Response, error)
	FindAll() (roles []Entity, err error)
}

func NewController(server *web.Server, roleService Svc) *Controller {
	return &Controller{
		server:  server,
		service: roleService,
	}
}

func (c *Controller) RegisterRouters() {
	c.server.GroupApiV1.Post("/roles/add", c.AddRoles)
	c.server.GroupApiV1.Post("/roles/ids", c.FindByIds)
	c.server.GroupApiV1.Post("/roles/:id", c.FindById)
	c.server.GroupApiV1.Delete("/roles/ids", c.DeleteByIds)
	c.server.GroupApiV1.Delete("/roles/:id", c.DeleteById)
	c.server.GroupApiV1.Get("/roles", c.FindAll)
}

func (c *Controller) AddRoles(ctx *fiber.Ctx) error {
	var entity Entity
	if err := ctx.BodyParser(&entity); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	var newRoleId, err = c.service.Add(entity)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newRoleId)
}

func (c *Controller) FindById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		_ = common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	employee, err := c.service.FindById(id)
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
	roles, err := c.service.FindByIds(ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error finding roles")
	}
	return common.OkResponse(ctx, roles)
}

func (c *Controller) DeleteById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	rsl, err := c.service.DeleteById(id)
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
	rsl, err := c.service.DeleteByIds(ids)
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

func (c *Controller) FindAll(ctx *fiber.Ctx) error {
	roles, err := c.service.FindAll()
	if err != nil {
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, roles)
}
