package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultAPIBase = "https://api.spotify.com/v1"

type Client struct {
	apiBase      string
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	auth         TokenRefresher
	userID       int64
	onRefresh    func(accessToken, refreshToken string, expiresAt time.Time) error
}

func NewClient(accessToken, refreshToken string, expiresAt time.Time, auth TokenRefresher, userID int64, onRefresh func(string, string, time.Time) error) *Client {
	return &Client{
		apiBase:      defaultAPIBase,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		auth:         auth,
		userID:       userID,
		onRefresh:    onRefresh,
	}
}

// NewTestClient returns a Client that targets baseURL instead of the real Spotify API.
// The token is set to never expire so tests do not trigger a refresh.
func NewTestClient(baseURL, accessToken string) *Client {
	return &Client{
		apiBase:     baseURL,
		accessToken: accessToken,
		expiresAt:   time.Now().Add(24 * time.Hour),
	}
}

func (c *Client) Get(ctx context.Context, path string, out any) error {
	if time.Now().After(c.expiresAt.Add(-30 * time.Second)) {
		if err := c.refresh(ctx); err != nil {
			return fmt.Errorf("refresh token: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBase+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			retryAfter := time.Second * 1
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryAfter):
			}
			continue
		}
		break
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("spotify API %s: status %d: %s", path, resp.StatusCode, string(body))
	}

	return json.Unmarshal(body, out)
}

func (c *Client) refresh(ctx context.Context) error {
	tok, err := c.auth.RefreshToken(ctx, c.refreshToken)
	if err != nil {
		return err
	}
	c.accessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		c.refreshToken = tok.RefreshToken
	}
	c.expiresAt = TokenExpiresAt(tok.ExpiresIn)
	if c.onRefresh != nil {
		return c.onRefresh(c.accessToken, c.refreshToken, c.expiresAt)
	}
	return nil
}
