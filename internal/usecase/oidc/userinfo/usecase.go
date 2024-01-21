package userinfo

import (
	"context"
	"myoidc/internal/domain"
	oidccli "myoidc/internal/service/oidc/client"
	"myoidc/internal/service/session"
	"myoidc/internal/usecase"
	"myoidc/pkg/errors"
)

type UseCase struct {
	reg oidccli.ClientRegistry
	sm  session.Manager
}

func NewUseCase(
	reg oidccli.ClientRegistry,
	sm session.Manager,
) *UseCase {
	return &UseCase{
		reg: reg,
		sm:  sm,
	}
}

func (uc UseCase) Execute(ctx context.Context, sessId string) (*domain.User, error) {
	sess, err := uc.sm.Get(ctx, sessId)
	if err != nil {
		err = errors.Wrap(err, "session not found")
		err = errors.WithField(err, "sessId", sessId)
		return nil, errors.WithCode(err, usecase.ErrCodeUserUnauthorized)
	}

	auth := session.AuthData(sess.Data)
	if !auth.IsValid() {
		defer uc.sm.Destroy(ctx, sessId) // destroy invalid session
		err = errors.Errorf("session authentication data not found or invalid: %+v", sess.Data)
		err = errors.WithField(err, "sessId", sessId)
		return nil, errors.WithCode(err, usecase.ErrCodeSessionInterrupt)
	}

	client, err := uc.reg.GetClient(ctx, auth.GetProviderName())
	if err != nil {
		defer uc.sm.Destroy(ctx, sessId) // destroy invalid session
		err = errors.Wrapf(err, "client not found for provider %s", auth.GetProviderName())
		err = errors.WithField(err, "sessId", sessId)
		return nil, errors.WithCode(err, usecase.ErrCodeEntityNotFound)
	}

	user, err := client.FetchUserByToken(ctx, &oidccli.Token{
		Access:  auth.GetAccessToken(),
		Refresh: auth.GetRefreshToken(),
	})
	if err != nil {
		err = errors.Wrap(err, "failed to fetch user by oidc token")
		err = errors.WithField(err, "sessId", sessId)
		return nil, err
	}

	return user, nil
}
