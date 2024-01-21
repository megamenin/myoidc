package config

import (
	"github.com/spf13/viper"
	"strings"
)

func init() {
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetConfigFile("default.toml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func Load(dist interface{}, opts ...viper.DecoderConfigOption) error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(dist, opts...)
	if err != nil {
		return err
	}
	return nil
}

type Config struct {
	Domain           string
	DisableTLSVerify bool
	OidcClients      []OIDCClientConfig
}

func NewConfig() *Config {
	return &Config{
		OidcClients: make([]OIDCClientConfig, 0),
	}
}

type OIDCClientConfig struct {
	ProviderName        string
	ClientId            string
	AuthUrl             string
	TokenUrl            string
	UserInfoUrl         string
	ClientSecret        string
	UseState            bool
	UsePKCE             bool
	StateLength         int
	PKCEMethod          string
	PKCEChallengeLength int `default:"32"`
	UserInfoUnmarshaler string
}
