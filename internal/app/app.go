package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/sirupsen/logrus"
	"myoidc/internal/config"
	"myoidc/internal/handler/http"
	"myoidc/internal/handler/http/oidc"
	oidccli "myoidc/internal/service/oidc/client"
	"myoidc/internal/service/oidc/client/oauth0"
	"myoidc/internal/service/oidc/pkce"
	"myoidc/internal/service/session/inmemory"
	oidc_callback "myoidc/internal/usecase/oidc/callback"
	oidc_login "myoidc/internal/usecase/oidc/login"
	oidc_userinfo "myoidc/internal/usecase/oidc/userinfo"
	"myoidc/pkg/errors"
	"myoidc/pkg/log"
)

func init() {
	log.SetDefault(log.NewLogrusLogger(
		log.LogrusFormatter(&logrus.TextFormatter{DisableQuote: true}),
	))
	oidccli.RegisterUnmarshaler("oauth0", &oauth0.OAuth0Unmarshaler{})
}

type App struct {
	cfg *config.Config
	r   *fiber.App
	l   log.Logger
}

func (app App) Run() {
	err := app.r.Listen("0.0.0.0:8080")
	if err != nil {
		app.l.WithError(err).Fatal("failed to start http server")
	}
}

func Setup() *App {
	l := log.GetDefault()

	// load configs
	cfg := config.NewConfig()
	err := config.Load(cfg)
	if err != nil {
		l.WithError(err).Fatal("config load error")
	}

	// setup services
	sm := inmemory.NewManager()
	reg, err := buildOidcClientRegistry(cfg)
	if err != nil {
		l.WithError(err).Fatal("oidc setup error")
	}

	// setup use cases
	oidcLoginUseCase := oidc_login.NewUseCase(reg, sm)
	oidcCallbackUseCase := oidc_callback.NewUseCase(reg, sm)
	oidcUserInfoUseCase := oidc_userinfo.NewUseCase(reg, sm)

	// setup http handlers
	r := fiber.New(fiber.Config{AppName: "myoidc"})
	r.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			err := errors.Errorf("%+v", e)
			l.WithError(err).Errorf("unhandled error on %s", c.Path())
		},
	}))
	r.Get("/oauth/login", oidc.NewLoginHandler(oidcLoginUseCase, l).Handler())
	r.Get("/oauth/callback", oidc.NewCallbackHandler(oidcCallbackUseCase, l).Handler())
	r.Get("/oauth/userinfo", oidc.NewUserInfoHandler(oidcUserInfoUseCase, l).Handler())

	r.Get("/*", func(c *fiber.Ctx) error {
		sessId := http.GetCookie(c, http.CookieSessionId)
		if sessId != "" {
			return c.Redirect("/oauth/userinfo")
		} else {
			return c.Redirect("/oauth/login?providerName=myoidc")
		}
	})

	return &App{cfg, r, l}
}

func buildOidcClientRegistry(cfg *config.Config) (oidccli.ClientRegistry, error) {
	oidcClientConfigs := make(map[string]oidccli.ClientConfig)
	for i, conf := range cfg.OidcClients {
		cli := oidccli.ClientConfig{
			ClientId:            conf.ClientId,
			AuthUrl:             conf.AuthUrl,
			TokenUrl:            conf.TokenUrl,
			UserInfoUrl:         conf.UserInfoUrl,
			ClientSecret:        conf.ClientSecret,
			UsePKCE:             conf.UsePKCE,
			UseState:            conf.UseState,
			RedirectUrl:         cfg.Domain + "/oauth/callback?providerName=" + conf.ProviderName,
			DisableSslVerify:    cfg.DisableTLSVerify,
			UserInfoUnmarshaler: conf.UserInfoUnmarshaler,
		}
		var err error
		cli.PKCEGenerator, err = pkce.NewPKCEGenerator(
			conf.PKCEMethod,
			conf.StateLength,
			conf.PKCEChallengeLength,
		)
		if err != nil {
			err = errors.Wrapf(err, "invalid pkce config for oidc provider [%d] \"%s\"", i, conf.ProviderName)
			return nil, err
		}
		oidcClientConfigs[conf.ProviderName] = cli
	}

	if len(oidcClientConfigs) == 0 {
		return nil, errors.Error("no oidc provider is registered")
	}

	return oidccli.NewGenericClientRegistry(oidcClientConfigs)
}
