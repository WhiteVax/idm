package role

import (
	"encoding/json"
	"idm/inner/common"
	"idm/inner/web"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	server  *web.Server
	service Svc
	logger  *common.Logger
}

type Svc interface {
	Add(role Entity) (Response, error)
	FindById(id int64) (Response, error)
	FindByIds(ids []int64) ([]Response, error)
	DeleteByIds(ids []int64) ([]Response, error)
	DeleteById(id int64) (Response, error)
	FindAll() (roles []Entity, err error)
}

func NewHandler(server *web.Server, roleService Svc, logger *common.Logger) *Handler {
	return &Handler{
		server:  server,
		service: roleService,
		logger:  logger,
	}
}

func (c *Handler) RegisterRouters() {
	c.server.GroupApiV1.Post("/roles/add", c.AddRoles)
	c.server.GroupApiV1.Post("/roles/ids", c.FindByIds)
	c.server.GroupApiV1.Post("/roles/:id", c.FindById)
	c.server.GroupApiV1.Delete("/roles/ids", c.DeleteByIds)
	c.server.GroupApiV1.Delete("/roles/:id", c.DeleteById)
	c.server.GroupApiV1.Get("/roles", c.FindAll)
}

func (c *Handler) AddRoles(ctx *fiber.Ctx) error {
	var entity Entity
	if err := ctx.BodyParser(&entity); err != nil {
		c.logger.Error("AddRoles: invalid request body", zap.Error(err))
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	c.logger.Debug("AddRoles: receive entity", zap.Any("entity", ctx.Body()))
	var newRoleId, err = c.service.Add(entity)
	if err != nil {
		c.logger.Error("AddRoles: error adding role", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, newRoleId)
}

func (c *Handler) FindById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("FindById: invalid request", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("FindById: receive id", zap.Any("id", idParam))
	employee, err := c.service.FindById(id)
	if err != nil {
		c.logger.Error("FindById: error finding employee", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, employee)
}

func (c *Handler) FindByIds(ctx *fiber.Ctx) error {
	var ids []int64
	if err := ctx.BodyParser(&ids); err != nil {
		c.logger.Error("FindByIds: invalid request body", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.Debug("FindByIds: receive ids", zap.Any("ids", ids))
	roles, err := c.service.FindByIds(ids)
	if err != nil {
		c.logger.Error("FindByIds: error finding roles", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error finding roles")
	}
	return common.OkResponse(ctx, roles)
}

func (c *Handler) DeleteById(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.logger.Error("DeleteById: invalid request", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, err.Error())
	}
	c.logger.Debug("DeleteById: receive id", zap.Any("id", idParam))
	rsl, err := c.service.DeleteById(id)
	if err != nil {
		c.logger.Error("DeleteById: error deleting role", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

func (c *Handler) DeleteByIds(ctx *fiber.Ctx) error {
	bodyBytes := ctx.Body()
	var ids []int64
	if err := json.Unmarshal(bodyBytes, &ids); err != nil {
		c.logger.Error("DeleteByIds: invalid request body", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusBadRequest, "Invalid request body")
	}
	c.logger.Debug("DeleteByIds: receive ids", zap.Any("ids", ids))
	rsl, err := c.service.DeleteByIds(ids)
	if err != nil {
		c.logger.Error("DeleteByIds: error deleting roles", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, rsl)
}

func (c *Handler) FindAll(ctx *fiber.Ctx) error {
	roles, err := c.service.FindAll()
	if err != nil {
		c.logger.Error("FindAll: error finding roles", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, err.Error())
	}
	return common.OkResponse(ctx, roles)
}
