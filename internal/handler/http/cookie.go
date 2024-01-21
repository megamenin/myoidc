package http

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

const CookieSessionId = "sessId"

func GetCookie(c *fiber.Ctx, key string) string {
	return string(c.Request().Header.Cookie(key))
}

func SetCookie(c *fiber.Ctx, key string, value string) {
	c.Cookie(&fiber.Cookie{
		Name:     key,
		Value:    value,
		Path:     "/",
		HTTPOnly: true,
	})
}

func DelCookie(c *fiber.Ctx, key string) {
	c.Cookie(&fiber.Cookie{
		Name:     key,
		Expires:  time.Now(),
		Path:     "/",
		HTTPOnly: true,
	})
}
