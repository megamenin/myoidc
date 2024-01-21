package oidc

import (
	"myoidc/internal/handler/http"
	"myoidc/internal/usecase"
	"myoidc/internal/usecase/oidc/login"
	"myoidc/pkg/errors"
	"myoidc/pkg/log"

	"github.com/gofiber/fiber/v2"
)

var defaultScopes = []string{"openid", "profile", "email", "phone", "address"}

type LoginHandler struct {
	useCase *login.UseCase
	logger  log.Logger
}

func NewLoginHandler(useCase *login.UseCase, logger log.Logger) *LoginHandler {
	return &LoginHandler{
		useCase: useCase,
		logger:  logger,
	}
}

func (h *LoginHandler) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		providerName := c.Query("providerName")
		if providerName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"message": "fail to parse request: providerName is missing",
			})
		}

		authUrl, err := h.useCase.Execute(c.Context(), providerName, defaultScopes, map[string]interface{}{
			//"backUrl": c.Get("Referer"),
			"backUrl": "/",
		})
		if err != nil {
			switch errors.GetErrCode(err) {
			case usecase.ErrCodeEntityNotFound:
				h.logger.WithError(err).Warn("oidc login error")
				return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
					"message": "providerName is missing or invalid",
				})
			default:
				h.logger.WithError(err).Error(http.UnexpectedErrorMessage(c))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.ErrInternalServerError)
			}
		}

		return c.Redirect(authUrl.String(), 302)
	}
}
