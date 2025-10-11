package web

import (
	_ "idm/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type Server struct {
	App           *fiber.App
	GroupApi      fiber.Router
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

type AuthMiddlewareInterface interface {
	ProtectWithJwt() func(*fiber.Ctx) error
}

func registerMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New())
}

// NewServer - функция-конструктор
func NewServer() *Server {
	// новый веб-вервер
	app := fiber.New(fiber.Config{
		AppName: "Idm app",
	})
	// не поубличный
	groupInternal := app.Group("/internal")
	// группа "/api"
	groupApi := app.Group("/api")
	// подгруппа "api/v1"
	groupApiV1 := groupApi.Group("/v1")
	return &Server{
		App:           app,
		GroupApi:      groupApi,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
