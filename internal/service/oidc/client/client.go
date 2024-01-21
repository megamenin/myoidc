package client

import (
	"context"
	"myoidc/internal/domain"
	"myoidc/internal/service/oidc/pkce"
	"myoidc/pkg/errors"
	"net/url"
)

type Token struct {
	Access  string
	Refresh *string
}

func (t Token) Valid() bool {
	return t.Access != "" && (t.Refresh == nil || *t.Refresh != "")
}

type UrlParam struct {
	Key   string
	Value string
}

func NewUrlParam(key, value string) UrlParam {
	return UrlParam{
		Key:   key,
		Value: value,
	}
}

// Client for Authorization Code Flow.
type Client interface {
	BuildAuthURL(ctx context.Context, state string, scopes []string, tempSessId string, params ...UrlParam) (*url.URL, error)
	FetchTokenByCode(ctx context.Context, code string, tempSessId string, params ...UrlParam) (*Token, error)
	FetchUserByToken(ctx context.Context, token *Token) (*domain.User, error)
	RefreshToken(ctx context.Context, token *Token) (*Token, error)
}

// SecurityClient supports State and Proof Key of Code Exchange (PKCE).
type SecurityClient interface {
	SupportsState(ctx context.Context) bool
	SupportsPKCE(ctx context.Context) bool
	GetPKCEGenerator(ctx context.Context) pkce.PKCEGenerator
}

// ClientRegistry contains prepared clients for supported OIDC providers.
type ClientRegistry interface {
	// GetClient returns client by provider name.
	GetClient(ctx context.Context, providerName string) (Client, error)
}

type MapClientRegistry map[string]Client

func (mcr MapClientRegistry) GetClient(ctx context.Context, providerName string) (Client, error) {
	client, exists := mcr[providerName]
	if !exists {
		return nil, errors.Errorf("no such client \"%s\"", providerName)
	}
	return client, nil
}
