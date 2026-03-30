// Package common provides shared infrastructure for all provider adapters.
package common

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// OAuthConfig holds the provider-specific OAuth parameters.
type OAuthConfig struct {
	ProviderID   string          // e.g. "spotify", "youtube"
	ClientIDEnv  string          // env var name for client ID
	ClientSecEnv string          // env var name for client secret
	Scopes       []string        // OAuth scopes
	Endpoint     oauth2.Endpoint // OAuth endpoint (auth + token URL)
}

// BaseAdapter contains shared fields and methods for all provider adapters.
type BaseAdapter struct {
	HTTPClient *http.Client // If set, used instead of creating from token (for testing)
	OAuthCfg   OAuthConfig  // Provider-specific OAuth configuration
}

// GetRedirectURL returns the frontend redirect URL from env or the default.
func GetRedirectURL() string {
	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:5175/"
	}
	return redirectURL
}

// GetOAuth2Config builds an oauth2.Config from the adapter's OAuthConfig.
func (b *BaseAdapter) GetOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv(b.OAuthCfg.ClientIDEnv),
		ClientSecret: os.Getenv(b.OAuthCfg.ClientSecEnv),
		RedirectURL:  GetRedirectURL(),
		Scopes:       b.OAuthCfg.Scopes,
		Endpoint:     b.OAuthCfg.Endpoint,
	}
}

// GetAuthURL generates the OAuth authorization URL.
func (b *BaseAdapter) GetAuthURL() string {
	return b.GetOAuth2Config().AuthCodeURL(b.OAuthCfg.ProviderID, oauth2.AccessTypeOffline)
}

// ExchangeAuthCode exchanges an authorization code for an access token.
func (b *BaseAdapter) ExchangeAuthCode(ctx context.Context, code string) (string, error) {
	token, err := b.GetOAuth2Config().Exchange(ctx, code)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

// GetHTTPClient returns the injected HTTP client or creates one from the auth token.
func (b *BaseAdapter) GetHTTPClient(ctx context.Context, authToken string) *http.Client {
	if b.HTTPClient != nil {
		return b.HTTPClient
	}
	token := &oauth2.Token{
		AccessToken: authToken,
		TokenType:   "Bearer",
	}
	return oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
}
