package oidc

import (
	"myoidc/internal/handler/http"
	"myoidc/internal/usecase"
	"myoidc/internal/usecase/oidc/userinfo"
	"myoidc/pkg/errors"
	"myoidc/pkg/log"

	"github.com/gofiber/fiber/v2"
)

type UserInfoHandler struct {
	uc *userinfo.UseCase
	l  log.Logger
}

func NewUserInfoHandler(uc *userinfo.UseCase, l log.Logger) *UserInfoHandler {
	return &UserInfoHandler{
		uc: uc,
		l:  l,
	}
}

func (h *UserInfoHandler) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessId := http.GetCookie(c, http.CookieSessionId)

		user, err := h.uc.Execute(c.Context(), sessId)
		if err != nil {
			// clear session cookie on any error
			http.DelCookie(c, http.CookieSessionId)

			switch errors.GetErrCode(err) {
			case usecase.ErrCodeEntityNotFound:
				h.l.WithError(err).Warnf("oidc userinfo failed with not found error")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.ErrUnauthorized)
			case usecase.ErrCodeSessionInterrupt:
				h.l.WithError(err).Warnf("oidc userinfo failed with session interrupt error")
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.ErrUnprocessableEntity)
			case usecase.ErrCodeUserUnauthorized:
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.ErrUnauthorized)
			default:
				h.l.WithError(err).Errorf(http.UnexpectedErrorMessage(c))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.ErrInternalServerError)
			}
		}

		return c.JSON(user)
	}
}
