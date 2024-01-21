package session

import "context"

type Manager interface {
	Create(ctx context.Context, userId string, data map[string]interface{}) (*Session, error)
	CreateTemp(ctx context.Context, data map[string]interface{}) (*Session, error)
	Get(ctx context.Context, sessId string) (*Session, error)
	Destroy(ctx context.Context, sessId string) error
}

type Session struct {
	Id     string
	UserId string
	Data   map[string]interface{}
}

func NewSession(id string, userId string, data map[string]interface{}) *Session {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &Session{
		Id:     id,
		UserId: userId,
		Data:   data,
	}
}

type AuthData map[string]interface{}

func NewAuthData(providerName string, accessToken string, refreshToken *string) AuthData {
	m := AuthData{
		"providerName": providerName,
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}
	return m
}

func (data AuthData) IsValid() bool {
	return GetString(data, "providerName", "") != "" &&
		GetString(data, "accessToken", "") != ""
}

func (data AuthData) GetProviderName() string {
	return GetString(data, "providerName", "")
}

func (data AuthData) GetAccessToken() string {
	return GetString(data, "accessToken", "")
}

func (data AuthData) GetRefreshToken() *string {
	refreshToken := GetString(data, "refreshToken", "")
	if refreshToken == "" {
		return nil
	}
	return &refreshToken
}

func GetString(m map[string]interface{}, key string, def string) string {
	v, ok := m[key].(string)
	if ok {
		return v
	}
	return def
}
