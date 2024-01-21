package callback

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

type Result struct {
	Session     *session.Session
	ActiveUser  *domain.User
	RedirectURL string
}

func NewUseCase(reg oidccli.ClientRegistry, sm session.Manager) *UseCase {
	return &UseCase{
		reg: reg,
		sm:  sm,
	}
}

func (uc UseCase) Execute(ctx context.Context, providerName string, code string, state string, tempSessId string) (*Result, error) {
	client, err := uc.reg.GetClient(ctx, providerName)
	if err != nil {
		err = errors.Wrapf(err, "client not found for provider %s", providerName)
		return nil, errors.WithCode(err, usecase.ErrCodeEntityNotFound)
	}

	sess, err := uc.sm.Get(ctx, tempSessId)
	if err != nil {
		err = errors.Wrap(err, "temporary session not found")
		err = errors.WithField(err, "tempSessId", tempSessId)
		return nil, errors.WithCode(err, usecase.ErrCodeUserUnauthorized)
	}
	defer uc.sm.Destroy(ctx, tempSessId)

	var params = make([]oidccli.UrlParam, 0)
	if security, ok := client.(oidccli.SecurityClient); ok {
		if security.SupportsState(ctx) {
			sessState := session.GetString(sess.Data, "oidcState", "")
			if state != sessState {
				err = errors.Error("state doesn't match")
				return nil, errors.WithCode(err, usecase.ErrCodeUserUnauthorized)
			}
		}
		if security.SupportsPKCE(ctx) {
			codeVerifier := session.GetString(sess.Data, "oidcCodeVerifier", "")
			params = append(params, oidccli.NewUrlParam("code_verifier", codeVerifier))
		}
	}
	token, err := client.FetchTokenByCode(ctx, code, tempSessId, params...)
	if err != nil {
		err = errors.Wrap(err, "code is invalid")
		err = errors.WithField(err, "oidcCode", code)
		return nil, errors.WithCode(err, usecase.ErrCodeUserUnauthorized)
	}

	// fetch user from external system
	user, err := client.FetchUserByToken(ctx, token)
	if err != nil {
		err = errors.Wrap(err, "token is invalid")
		err = errors.WithField(err, "oidcAuthToken", token)
		return nil, errors.WithCode(err, usecase.ErrCodeUserUnauthorized)
	}

	// create persistent session instead of temp
	newSess, err := uc.sm.Create(ctx, user.Id, session.NewAuthData(providerName, token.Access, token.Refresh))
	if err != nil {
		err = errors.Wrap(err, "failed to create temp session")
		return nil, err
	}

	var redirectUrl = session.GetString(sess.Data, "backUrl", "/")
	return &Result{
		Session:     newSess,
		ActiveUser:  user,
		RedirectURL: redirectUrl,
	}, nil
}
