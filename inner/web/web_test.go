package web

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewServerMiddleware(t *testing.T) {
	t.Run("Server panic", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		server := NewServer()
		server.App.Get("/panic", func(c *fiber.Ctx) error {
			panic("panic")
		})
		req, err := http.NewRequest("GET", "/panic", nil)
		a.Nil(err)
		res, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(res.StatusCode, http.StatusInternalServerError)
	})

	t.Run("RequestId", func(t *testing.T) {
		t.Parallel()
		a := assert.New(t)
		server := NewServer()
		server.App.Get("/id", func(c *fiber.Ctx) error {
			return c.SendString("X-Request-ID")
		})

		req, err := http.NewRequest("GET", "/id", nil)
		a.Nil(err)
		res, err := server.App.Test(req)
		a.Nil(err)
		a.Equal(res.StatusCode, http.StatusOK)
		a.NotEmpty(res.Header.Get("X-Request-ID"))
	})
}
