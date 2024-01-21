package generic

import (
	"encoding/json"
	"myoidc/internal/domain"
)

var unmarshalers map[string]UserInfoUnmarshaler

func init() {
	unmarshalers = map[string]UserInfoUnmarshaler{
		"json": &JSONUnmarshaler{},
	}
}

func RegisterUnmarshaler(name string, um UserInfoUnmarshaler) {
	unmarshalers[name] = um
}

type UserInfoUnmarshaler interface {
	UnmarshalUserInfo(body []byte, user *domain.User) error
}

type JSONUnmarshaler struct{}

func (um JSONUnmarshaler) UnmarshalUserInfo(body []byte, user *domain.User) error {
	return json.Unmarshal(body, user)
}
