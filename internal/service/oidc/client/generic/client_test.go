package generic

import (
	"context"
	"golang.org/x/oauth2"
	"myoidc/internal/domain"
	oidccli "myoidc/internal/service/oidc/client"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericClient_BuildAuthURL(t *testing.T) {
	tests := []struct {
		name          string
		clientId      string
		scopes        []string
		state         string
		params        []oidccli.UrlParam
		expected      string
		expectedError bool
	}{
		{
			name:     "basic",
			clientId: "client_id",
			scopes:   []string{},
			state:    "",
			params:   nil,
			expected: "https://oauth.server.com/api/oidc/authenticate?client_id=client_id&redirect_uri=callback_url%3Ftoken%3DsessId&response_type=code",
		},
		{
			name:     "encoded client id",
			clientId: "client  &id",
			scopes:   []string{},
			state:    "",
			params:   nil,
			expected: "https://oauth.server.com/api/oidc/authenticate?client_id=client++%26id&redirect_uri=callback_url%3Ftoken%3DsessId&response_type=code",
		},
		{
			name:     "with scopes",
			clientId: "client_id",
			scopes:   []string{"openid", "profile"},
			state:    "",
			params:   nil,
			expected: "https://oauth.server.com/api/oidc/authenticate?client_id=client_id&redirect_uri=callback_url%3Ftoken%3DsessId&response_type=code&scope=openid+profile",
		},
		{
			name:     "with scopes and state",
			clientId: "client_id",
			scopes:   []string{"openid", "profile"},
			state:    "my_state",
			params:   nil,
			expected: "https://oauth.server.com/api/oidc/authenticate?client_id=client_id&redirect_uri=callback_url%3Ftoken%3DsessId&response_type=code&scope=openid+profile&state=my_state",
		},
		{
			name:     "with scopes, state and url params",
			clientId: "client_id",
			scopes:   []string{"openid", "profile"},
			state:    "my_state",
			params:   []oidccli.UrlParam{{Key: "challenge", Value: "123"}},
			expected: "https://oauth.server.com/api/oidc/authenticate?challenge=123&client_id=client_id&redirect_uri=callback_url%3Ftoken%3DsessId&response_type=code&scope=openid+profile&state=my_state",
		},
	}

	var client = &GenericClient{
		cfg: &oauth2.Config{
			ClientID:     "",
			ClientSecret: "",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://oauth.server.com/api/oidc/authenticate",
			},
			RedirectURL: "callback_url",
			Scopes:      []string{},
		},
	}

	for _, tt := range tests {
		client.cfg.ClientID = tt.clientId
		res, err := client.BuildAuthURL(context.TODO(), tt.state, tt.scopes, "sessId", tt.params...)
		if tt.expectedError {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.expected, res.String(), tt.name)
		}
	}
}

func TestGenericClient_FetchTokenByCode(t *testing.T) {
	var client = &GenericClient{
		cfg: &oauth2.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  "",
			Scopes:       []string{},
		},
		userInfoUrl: nil,
		HttpClient:  http.DefaultClient,
	}

	// case 1: success without refresh
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"ACCESS_TOKEN"}`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	res, err := client.FetchTokenByCode(context.TODO(), "code", "")
	if assert.NoError(t, err) {
		expected := &oidccli.Token{Access: "ACCESS_TOKEN", Refresh: nil}
		assert.Equal(t, expected, res, "success without refresh")
	}

	server.Close()

	// case 2: success with refresh
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	res, err = client.FetchTokenByCode(context.TODO(), "code", "")
	if assert.NoError(t, err) {
		expected := &oidccli.Token{Access: "ACCESS_TOKEN", Refresh: pointer("REFRESH_TOKEN")}
		assert.Equal(t, expected, res, "success with refresh")
	}

	server.Close()

	// case 3: response with empty refresh token
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "ACCESS_TOKEN", "refresh_token": ""}`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	res, err = client.FetchTokenByCode(context.TODO(), "code", "")
	if assert.NoError(t, err) {
		expected := &oidccli.Token{Access: "ACCESS_TOKEN", Refresh: nil}
		assert.Equal(t, expected, res, "response with empty refresh token")
	}

	server.Close()

	// case 4: invalid json response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	_, err = client.FetchTokenByCode(context.TODO(), "code", "")
	assert.Error(t, err, "invalid json response")

	// case 5: response with undefined access token
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	_, err = client.FetchTokenByCode(context.TODO(), "code", "")
	assert.Error(t, err, "response with undefined access token")

	server.Close()

	// case 6: response with empty access token
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": ""}`))
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	_, err = client.FetchTokenByCode(context.TODO(), "code", "")
	assert.Error(t, err, "response with empty access token")

	server.Close()

	// case 7: resource server error
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	_, err = client.FetchTokenByCode(context.TODO(), "code", "")
	assert.Error(t, err, "resource server error")

	server.Close()

	// case 8: invalid code
	const requestCode = "12345678"
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		code := r.FormValue("code")
		if code == requestCode {
			w.Write([]byte(`{"access_token":"ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
		}
		w.WriteHeader(401)
		return
	}))

	client.cfg.Endpoint.TokenURL = server.URL
	res, err = client.FetchTokenByCode(context.TODO(), requestCode, "")
	if assert.NoError(t, err) {
		expected := &oidccli.Token{Access: "ACCESS_TOKEN", Refresh: pointer("REFRESH_TOKEN")}
		assert.Equal(t, expected, res, "valid code")
	}
	_, err = client.FetchTokenByCode(context.TODO(), "another code", "")
	assert.Error(t, err, "invalid code")

	server.Close()
}

func TestGenericClient_FetchUserByToken(t *testing.T) {
	var client = &GenericClient{
		cfg: &oauth2.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  "",
			Scopes:       []string{},
		},
		userInfoUrl: nil,
		HttpClient:  http.DefaultClient,
	}

	// case 1: success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
          "id": "123",
          "login": "johndow90",
          "email": "johndow90@mail.test",
          "firstName": "John",
          "lastName": "Dow",
          "permissions": ["perm 1", "perm 2", "perm 3"],
          "filed1": 1,
          "filed2": "2"
        }`))
		return
	}))

	client.userInfoUrl, _ = url.Parse(server.URL)
	res, err := client.FetchUserByToken(context.TODO(), &oidccli.Token{
		Access:  "ACCESS_TOKEN",
		Refresh: nil,
	})
	if assert.NoError(t, err) {
		expected := &domain.User{
			Id:          "USER_ID",
			Login:       "johndow90",
			Email:       "johndow90@mail.test",
			FullName:    "John Dow",
			Permissions: []string{"perm 1", "perm 2", "perm 3"},
		}
		assert.Equal(t, expected, res, "success")
	}

	server.Close()

	// case 2: empty json response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
		return
	}))

	client.userInfoUrl, _ = url.Parse(server.URL)
	res, err = client.FetchUserByToken(context.TODO(), &oidccli.Token{
		Access:  "ACCESS_TOKEN",
		Refresh: nil,
	})
	if assert.NoError(t, err) {
		expected := &domain.User{
			Id: "USER_ID",
		}
		assert.Equal(t, expected, res, "empty json response")
	}

	server.Close()

	// case 3: invalid json response
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json`))
		return
	}))

	client.userInfoUrl, _ = url.Parse(server.URL)
	res, err = client.FetchUserByToken(context.TODO(), &oidccli.Token{
		Access:  "ACCESS_TOKEN",
		Refresh: nil,
	})
	assert.Error(t, err, "invalid json response")

	server.Close()

	// case 4: invalid token request
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`invalid json`))
		return
	}))

	client.userInfoUrl, _ = url.Parse(server.URL)
	res, err = client.FetchUserByToken(context.TODO(), &oidccli.Token{
		Access:  "",
		Refresh: nil,
	})
	assert.Error(t, err, "invalid token request")

	server.Close()
}

func TestGenericClient_RefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		token         *oidccli.Token
		handler       http.HandlerFunc
		expected      *oidccli.Token
		expectedError bool
	}{
		{
			name: "with refresh",
			token: &oidccli.Token{
				Access:  "OLD_ACCESS_TOKEN",
				Refresh: pointer("OLD_REFRESH_TOKEN"),
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"user_id": "USER_ID", "access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
				return
			},
			expected: &oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			expectedError: false,
		},
		{
			name: "without refresh",
			token: &oidccli.Token{
				Access:  "OLD_ACCESS_TOKEN",
				Refresh: nil,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"user_id": "USER_ID", "access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
				return
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "invalid request token",
			token: &oidccli.Token{
				Access:  "",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"user_id": "USER_ID", "access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
				return
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "invalid response token",
			token: &oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"user_id": "", "access_token": "ACCESS_TOKEN", "refresh_token": "REFRESH_TOKEN"}`))
				return
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "invalid response json",
			token: &oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`invalid json`))
				return
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "resource server error",
			token: &oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(500)
				return
			},
			expected:      nil,
			expectedError: true,
		},
	}

	var client = &GenericClient{
		cfg: &oauth2.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			Endpoint:     oauth2.Endpoint{},
			RedirectURL:  "",
			Scopes:       []string{},
		},
		userInfoUrl: nil,
		HttpClient:  http.DefaultClient,
	}

	for _, tt := range tests {
		server := httptest.NewServer(tt.handler)

		client.cfg.Endpoint.TokenURL = server.URL
		res, err := client.RefreshToken(context.TODO(), tt.token)
		if tt.expectedError {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.expected, res, tt.name)
		}

		server.Close()
	}
}

func TestToken_Valid(t *testing.T) {
	tests := []struct {
		name     string
		token    oidccli.Token
		expected bool
	}{
		{
			name: "without refresh",
			token: oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: nil,
			},
			expected: true,
		},
		{
			name: "with refresh",
			token: oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer("REFRESH_TOKEN"),
			},
			expected: true,
		},
		{
			name: "with empty access token",
			token: oidccli.Token{
				Access:  "",
				Refresh: nil,
			},
			expected: false,
		},
		{
			name: "with empty refresh token",
			token: oidccli.Token{
				Access:  "ACCESS_TOKEN",
				Refresh: pointer(""),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.token.Valid(), tt.name)
	}
}

func pointer[T any](v T) *T {
	return &v
}
