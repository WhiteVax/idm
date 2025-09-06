package info

import (
	"idm/inner/common"
	"idm/inner/web"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Handler struct {
	server *web.Server
	cfg    common.Config
	db     *sqlx.DB
	logger *common.Logger
}

func NewHandler(server *web.Server, cfg common.Config, db *sqlx.DB, logger *common.Logger) *Handler {
	return &Handler{
		server: server,
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

type InfoResponse struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// RegisterRoutes() - регистрация  "/internal/info" / "/internal/health"
func (c *Handler) RegisterRoutes() {
	c.server.GroupInternal.Get("/info", c.GetInfo)
	c.server.GroupInternal.Get("/health", c.GetHealth)
}

// GetInfo - получение информации о приложении
func (c *Handler) GetInfo(ctx *fiber.Ctx) error {
	var err = ctx.Status(fiber.StatusOK).JSON(&InfoResponse{
		Name:    c.cfg.AppName,
		Version: c.cfg.AppVersion,
	})
	if err != nil {
		c.logger.Error("GetInfo", zap.Error(err))
		return common.ErrResponse(ctx, fiber.StatusInternalServerError, "Error returning info")
	}
	return nil
}

// GetHealth - проверка работоспособности приложения
func (c *Handler) GetHealth(ctx *fiber.Ctx) error {
	if err := c.db.Ping(); err != nil {
		c.logger.Error("GetHealth", zap.Error(err))
		return ctx.Status(fiber.StatusServiceUnavailable).SendString("Error pinging database")

	}
	return ctx.Status(fiber.StatusOK).SendString("OK")
}
