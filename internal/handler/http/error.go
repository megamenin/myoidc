package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func UnexpectedErrorMessage(c *fiber.Ctx) string {
	return fmt.Sprintf("Unexpected error while performing action %s", c.Path())
}
