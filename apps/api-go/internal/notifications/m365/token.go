package m365

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// TokenResponse represents the OAuth2 token response from Microsoft
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// TokenCache holds a cached access token with expiry information
type TokenCache struct {
	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
}

// FetchToken retrieves an access token using client credentials flow
// It caches the token and reuses it until expiry
func FetchToken(ctx context.Context, tenantID, clientID, clientSecret string, cache *TokenCache) (string, time.Time, error) {
	// Check cache first
	if cache != nil {
		cache.mu.RLock()
		if cache.accessToken != "" && time.Now().Before(cache.expiresAt) {
			token := cache.accessToken
			expiresAt := cache.expiresAt
			cache.mu.RUnlock()
			return token, expiresAt, nil
		}
		cache.mu.RUnlock()
	}

	// Fetch new token
	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenantID)

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("token request returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", time.Time{}, fmt.Errorf("empty access token in response")
	}

	// Calculate expiry time with 5-minute buffer
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	// Update cache
	if cache != nil {
		cache.mu.Lock()
		cache.accessToken = tokenResp.AccessToken
		cache.expiresAt = expiresAt
		cache.mu.Unlock()
	}

	return tokenResp.AccessToken, expiresAt, nil
}

// NewTokenCache creates a new empty token cache
func NewTokenCache() *TokenCache {
	return &TokenCache{}
}
