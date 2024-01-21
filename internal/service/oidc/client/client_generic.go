package client

import (
	"context"
	"crypto/tls"
	"golang.org/x/oauth2"
	"io"
	"myoidc/internal/domain"
	"myoidc/internal/service/oidc/pkce"
	"myoidc/pkg/errors"
	"net/http"
	"net/url"
)

type GenericClient struct {
	cfg           *oauth2.Config
	userInfoUrl   *url.URL
	useState      bool
	usePKCE       bool
	pkceGenerator pkce.PKCEGenerator
	um            UserInfoUnmarshaler

	HttpClient *http.Client
}

func (cli GenericClient) BuildAuthURL(ctx context.Context, state string, scopes []string, tempSessId string, params ...UrlParam) (*url.URL, error) {
	_, config := cli.prepareOAuth2Client(ctx)
	config.Scopes = scopes

	redirectUrl, err := cli.buildRedirectUrl(config.RedirectURL, tempSessId)
	if err != nil {
		return nil, err
	}
	config.RedirectURL = redirectUrl

	options := make([]oauth2.AuthCodeOption, 0)
	for _, param := range params {
		options = append(options, oauth2.SetAuthURLParam(param.Key, param.Value))
	}
	loginUrlStr := config.AuthCodeURL(state, options...)
	return url.Parse(loginUrlStr)
}

func (cli GenericClient) FetchTokenByCode(ctx context.Context, code string, tempSessId string, params ...UrlParam) (*Token, error) {
	ctx, config := cli.prepareOAuth2Client(ctx)

	redirectUrl, err := cli.buildRedirectUrl(config.RedirectURL, tempSessId)
	if err != nil {
		return nil, err
	}
	config.RedirectURL = redirectUrl

	options := make([]oauth2.AuthCodeOption, 0)
	for _, param := range params {
		options = append(options, oauth2.SetAuthURLParam(param.Key, param.Value))
	}
	return cli.parseToken(config.Exchange(ctx, code, options...))
}

func (cli GenericClient) FetchUserByToken(ctx context.Context, token *Token) (*domain.User, error) {
	if token == nil || !token.Valid() {
		return nil, errors.Error("token is invalid")
	}

	req, err := http.NewRequest("POST", cli.userInfoUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token.Access)
	res, err := cli.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, errors.Errorf("invalid response from oidc server: %s", body)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = cli.um.UnmarshalUserInfo(body, &user)
	//err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (cli GenericClient) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil || !token.Valid() {
		return nil, errors.Error("token is invalid")
	} else if token.Refresh == nil {
		return nil, errors.Error("refresh token is empty")
	}
	ctx, client := cli.prepareOAuth2Client(ctx)
	res := client.TokenSource(ctx, &oauth2.Token{RefreshToken: *token.Refresh})
	return cli.parseToken(res.Token())
}

func (cli GenericClient) SupportsState(ctx context.Context) bool {
	return cli.useState
}

func (cli GenericClient) SupportsPKCE(ctx context.Context) bool {
	return cli.usePKCE
}

func (cli GenericClient) GetPKCEGenerator(ctx context.Context) pkce.PKCEGenerator {
	return cli.pkceGenerator
}

func (cli GenericClient) prepareOAuth2Client(ctx context.Context) (context.Context, *oauth2.Config) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, cli.HttpClient)
	return ctx, &oauth2.Config{
		ClientID:     cli.cfg.ClientID,
		ClientSecret: cli.cfg.ClientSecret,
		Endpoint:     cli.cfg.Endpoint,
		RedirectURL:  cli.cfg.RedirectURL,
		Scopes:       cli.cfg.Scopes,
	}
}

func (cli GenericClient) parseToken(tokenData *oauth2.Token, err error) (*Token, error) {
	if err != nil {
		return nil, err
	}
	var token Token
	token.Access = tokenData.AccessToken
	if tokenData.RefreshToken != "" {
		token.Refresh = &tokenData.RefreshToken
	}
	if !token.Valid() {
		return nil, errors.Error("token data is invalid")
	}
	return &token, err
}

type ClientConfig struct {
	ClientId            string
	AuthUrl             string
	TokenUrl            string
	UserInfoUrl         string
	RedirectUrl         string
	ClientSecret        string
	UseState            bool
	UsePKCE             bool
	PKCEGenerator       pkce.PKCEGenerator
	DisableSslVerify    bool
	UserInfoUnmarshaler string
}

func NewGenericClient(cfg ClientConfig) (*GenericClient, error) {
	authUrl, err := url.Parse(cfg.AuthUrl)
	if err != nil {
		return nil, err
	}
	tokenUrl, err := url.Parse(cfg.TokenUrl)
	if err != nil {
		return nil, err
	}
	userInfoUrl, err := url.Parse(cfg.UserInfoUrl)
	if err != nil {
		return nil, err
	}
	um := unmarshalers[cfg.UserInfoUnmarshaler]
	if um == nil {
		return nil, errors.Errorf("unknown user unmarshaler \"%s\". Use generic.RegisterUnmarshaller to register a new unmarshaller type", cfg.UserInfoUnmarshaler)
	}

	var httpClient *http.Client

	if cfg.DisableSslVerify == true {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		httpClient = http.DefaultClient
	}

	return &GenericClient{
		cfg: &oauth2.Config{
			ClientID:     cfg.ClientId,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{},
			Endpoint: oauth2.Endpoint{
				TokenURL:  tokenUrl.String(),
				AuthURL:   authUrl.String(),
				AuthStyle: oauth2.AuthStyleInParams,
			},
			RedirectURL: cfg.RedirectUrl,
		},
		userInfoUrl:   userInfoUrl,
		useState:      cfg.UseState,
		usePKCE:       cfg.UsePKCE,
		pkceGenerator: cfg.PKCEGenerator,
		um:            um,

		HttpClient: httpClient,
	}, err
}

func (cli GenericClient) buildRedirectUrl(baseUrl, tempSessId string) (string, error) {
	redirectUri, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}
	query := redirectUri.Query()
	query.Add("token", tempSessId)
	redirectUri.RawQuery = query.Encode()

	return redirectUri.String(), nil
}

func NewGenericClientRegistry(configs map[string]ClientConfig) (ClientRegistry, error) {
	reg := make(map[string]Client)
	for name, cfg := range configs {
		cli, err := NewGenericClient(cfg)
		if err != nil {
			return nil, errors.Wrapf(err, "oidc client generation error for provider %s", name)
		}
		reg[name] = cli
	}
	return MapClientRegistry(reg), nil
}
