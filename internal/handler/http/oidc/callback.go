package oidc

import (
	"myoidc/internal/handler/http"
	"myoidc/internal/usecase"
	"myoidc/internal/usecase/oidc/callback"
	"myoidc/pkg/errors"
	"myoidc/pkg/log"

	"github.com/gofiber/fiber/v2"
)

type CallbackHandler struct {
	uc *callback.UseCase
	l  log.Logger
}

func NewCallbackHandler(uc *callback.UseCase, l log.Logger) *CallbackHandler {
	return &CallbackHandler{
		uc: uc,
		l:  l,
	}
}

func (h *CallbackHandler) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		providerName := c.Query("providerName")
		if providerName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"message": "providerName is missing",
			})
		}
		code := c.Query("code")
		if code == "" {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"message": "code is missing",
			})
		}
		sessId := c.Query("token")
		if sessId == "" {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"message": "token is missing",
			})
		}
		state := c.Query("state")
		if state == "" {
			return c.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"message": "state is missing",
			})
		}

		res, err := h.uc.Execute(c.Context(), providerName, code, state, sessId)
		if err != nil {
			switch errors.GetErrCode(err) {
			case usecase.ErrCodeEntityNotFound:
				h.l.WithError(err).Warnf("oidc login failed with not found error")
				return c.Status(fiber.StatusNotFound).JSON(fiber.ErrUnauthorized)
			case usecase.ErrCodeInvalidCredentials:
				h.l.WithError(err).Warnf("oidc login failed with invalid credentials error")
				return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.ErrUnprocessableEntity)
			case usecase.ErrCodeUserUnauthorized:
				h.l.WithError(err).Warnf("oidc login error")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.ErrUnauthorized)
			default:
				h.l.WithError(err).Errorf(http.UnexpectedErrorMessage(c))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.ErrInternalServerError)
			}
		}

		http.SetCookie(c, http.CookieSessionId, res.Session.Id)
		return c.Redirect(res.RedirectURL, 302)
	}
}
