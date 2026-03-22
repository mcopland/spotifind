package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authURL  = "https://accounts.spotify.com/authorize"
	tokenURL = "https://accounts.spotify.com/api/token"
)

var scopes = []string{
	"user-library-read",
	"user-follow-read",
	"playlist-read-private",
	"playlist-read-collaborative",
	"user-read-recently-played",
	"user-top-read",
	"user-read-email",
	"user-read-private",
}

type AuthClient struct {
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	tokenEndpoint string
}

func NewAuthClient(clientID, clientSecret, redirectURI string) *AuthClient {
	return &AuthClient{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		tokenEndpoint: tokenURL,
	}
}

// NewTestAuthClient creates an AuthClient that sends token requests to tokenEndpoint instead of the real Spotify API. Intended for use in tests.
func NewTestAuthClient(clientID, clientSecret, redirectURI, tokenEndpoint string) *AuthClient {
	return &AuthClient{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		tokenEndpoint: tokenEndpoint,
	}
}

func (a *AuthClient) AuthorizeURL(state string) string {
	params := url.Values{}
	params.Set("client_id", a.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", a.RedirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)
	return authURL + "?" + params.Encode()
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TokenRefresher is satisfied by *AuthClient and allows NewClient to accept an interface.
type TokenRefresher interface {
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
}

func (a *AuthClient) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", a.RedirectURI)

	return a.requestToken(ctx, data)
}

func (a *AuthClient) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	return a.requestToken(ctx, data)
}

func (a *AuthClient) requestToken(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(a.ClientID, a.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var tok TokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func TokenExpiresAt(expiresIn int) time.Time {
	return time.Now().Add(time.Duration(expiresIn) * time.Second)
}
