package web

import (
	_ "idm/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
)

type Server struct {
	App           *fiber.App
	GroupApiV1    fiber.Router
	GroupInternal fiber.Router
}

func registerMiddleware(app *fiber.App) {
	app.Use(recover.New())
	app.Use(requestid.New())
}

// функция-конструктор
func NewServer() *Server {
	// создаём новый веб-вервер
	app := fiber.New(fiber.Config{
		AppName: "Idm app",
	})
	// не поубличный
	registerMiddleware(app)
	app.Get("/swagger/*", swagger.HandlerDefault)
	groupInternal := app.Group("/internal")
	// создаём группу "/api"
	groupApi := app.Group("/api")
	// создаём подгруппу "api/v1"
	groupApiV1 := groupApi.Group("/v1")
	return &Server{
		App:           app,
		GroupApiV1:    groupApiV1,
		GroupInternal: groupInternal,
	}
}
