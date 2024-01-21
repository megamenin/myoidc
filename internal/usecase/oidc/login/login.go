package login

import (
	"context"
	oidccli "myoidc/internal/service/oidc/client"
	"myoidc/internal/service/session"
	"myoidc/internal/usecase"
	"myoidc/pkg/errors"
	"net/url"
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

func (uc UseCase) Execute(ctx context.Context, providerName string, scopes []string, tempSessionData map[string]interface{}) (*url.URL, error) {
	client, err := uc.reg.GetClient(ctx, providerName)
	if err != nil {
		err = errors.Wrapf(err, "client not found for provider %s", providerName)
		return nil, errors.WithCode(err, usecase.ErrCodeEntityNotFound)
	}

	state := ""
	params := make([]oidccli.UrlParam, 0)
	if security, ok := client.(oidccli.SecurityClient); ok {
		if security.SupportsState(ctx) {
			state = security.GetPKCEGenerator(ctx).State()
			tempSessionData["oidcState"] = state
		}
		if security.SupportsPKCE(ctx) {
			codeChallenge, codeChallengeMethod, codeVerifier := security.GetPKCEGenerator(ctx).CodeChallengeVerifier()
			tempSessionData["oidcCodeVerifier"] = codeVerifier
			params = append(
				params,
				oidccli.NewUrlParam("code_challenge", codeChallenge),
				oidccli.NewUrlParam("code_challenge_method", codeChallengeMethod),
			)
		}
	}

	sess, err := uc.sm.CreateTemp(ctx, tempSessionData)
	if err != nil {
		err = errors.Wrap(err, "failed to create temp session")
		return nil, err
	}

	authURL, err := client.BuildAuthURL(ctx, state, scopes, sess.Id, params...)
	if err != nil {
		err = errors.Wrapf(err, "failed to build auth url for temp session %s", sess.Id)
		return nil, err
	}

	return authURL, nil
}
