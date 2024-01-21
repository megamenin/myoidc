package surf

import (
	"encoding/json"
	"myoidc/internal/domain"
)

// SurfUnmarshaler see https://wiki.surfnet.nl/display/surfconextdev/OpenID+Connect+features
type SurfUnmarshaler struct{}

func (um SurfUnmarshaler) UnmarshalUserInfo(body []byte, user *domain.User) error {
	var m map[string]interface{}
	err := json.Unmarshal(body, &m)
	if err != nil {
		return err
	}

	user.Id, _ = m["eduid"].(string)
	user.Login, _ = m["eduperson_principal_name"].(string)
	user.Email, _ = m["email"].(string)
	user.FirstName, _ = m["given_name"].(string)
	user.LastName, _ = m["family_name"].(string)
	user.FullName, _ = m["name"].(string)

	return nil
}
