package oauth0

import (
	"encoding/json"
	"myoidc/internal/domain"
	"strings"
)

// OAuth0Unmarshaler see https://auth0.com/docs
type OAuth0Unmarshaler struct{}

func (um OAuth0Unmarshaler) UnmarshalUserInfo(body []byte, user *domain.User) error {
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}

	sub, _ := m["sub"].(string)
	user.Id = strings.TrimLeft(sub, "google-oauth2|")
	user.Login, _ = m["nickname"].(string)
	user.Email, _ = m["email"].(string)
	user.FirstName, _ = m["given_name"].(string)
	user.LastName, _ = m["family_name"].(string)
	user.FullName, _ = m["name"].(string)

	return nil
}
