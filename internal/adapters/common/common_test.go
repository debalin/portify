package common

import (
	"context"
	"net/http"
	"os"
	"testing"

	"golang.org/x/oauth2"
)

func TestGetRedirectURL_Default(t *testing.T) {
	os.Unsetenv("FRONTEND_URL")
	url := GetRedirectURL()
	if url != "http://localhost:5175/" {
		t.Errorf("Expected default redirect URL, got '%s'", url)
	}
}

func TestGetRedirectURL_FromEnv(t *testing.T) {
	os.Setenv("FRONTEND_URL", "https://portify.example.com/")
	defer os.Unsetenv("FRONTEND_URL")

	url := GetRedirectURL()
	if url != "https://portify.example.com/" {
		t.Errorf("Expected env redirect URL, got '%s'", url)
	}
}

func TestBaseAdapter_GetOAuth2Config(t *testing.T) {
	os.Setenv("TEST_CLIENT_ID", "my-client-id")
	os.Setenv("TEST_CLIENT_SECRET", "my-secret")
	defer os.Unsetenv("TEST_CLIENT_ID")
	defer os.Unsetenv("TEST_CLIENT_SECRET")

	b := &BaseAdapter{
		OAuthCfg: OAuthConfig{
			ProviderID:   "test",
			ClientIDEnv:  "TEST_CLIENT_ID",
			ClientSecEnv: "TEST_CLIENT_SECRET",
			Scopes:       []string{"scope1", "scope2"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://auth.example.com",
				TokenURL: "https://token.example.com",
			},
		},
	}

	cfg := b.GetOAuth2Config()
	if cfg.ClientID != "my-client-id" {
		t.Errorf("Expected client ID 'my-client-id', got '%s'", cfg.ClientID)
	}
	if cfg.ClientSecret != "my-secret" {
		t.Errorf("Expected client secret 'my-secret', got '%s'", cfg.ClientSecret)
	}
	if len(cfg.Scopes) != 2 {
		t.Errorf("Expected 2 scopes, got %d", len(cfg.Scopes))
	}
}

func TestBaseAdapter_GetAuthURL(t *testing.T) {
	b := &BaseAdapter{
		OAuthCfg: OAuthConfig{
			ProviderID: "test",
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://auth.example.com/authorize",
			},
		},
	}

	url := b.GetAuthURL()
	if url == "" {
		t.Error("Expected non-empty auth URL")
	}
}

func TestBaseAdapter_GetHTTPClient_WithInjected(t *testing.T) {
	injected := &http.Client{}
	b := &BaseAdapter{HTTPClient: injected}

	got := b.GetHTTPClient(context.Background(), "any-token")
	if got != injected {
		t.Error("Expected injected client to be returned")
	}
}

func TestBaseAdapter_GetHTTPClient_CreatesFromToken(t *testing.T) {
	b := &BaseAdapter{}

	got := b.GetHTTPClient(context.Background(), "test-token")
	if got == nil {
		t.Error("Expected non-nil client")
	}
}

func TestWithHTTPClient(t *testing.T) {
	c := &http.Client{}
	opt := WithHTTPClient(c)

	b := &BaseAdapter{}
	opt(b)

	if b.HTTPClient != c {
		t.Error("Expected WithHTTPClient to set HTTPClient")
	}
}

func TestWithBaseURL(t *testing.T) {
	opt := WithBaseURL("http://test:9090")

	b := &BaseAdapter{}
	opt(b)

	if b.BaseURL != "http://test:9090" {
		t.Errorf("Expected BaseURL 'http://test:9090', got '%s'", b.BaseURL)
	}
}

func TestApplyOptions(t *testing.T) {
	c := &http.Client{}
	b := &BaseAdapter{}

	b.ApplyOptions([]Option{
		WithHTTPClient(c),
		WithBaseURL("http://test:8080"),
	})

	if b.HTTPClient != c {
		t.Error("Expected ApplyOptions to set HTTPClient")
	}
	if b.BaseURL != "http://test:8080" {
		t.Errorf("Expected ApplyOptions to set BaseURL, got '%s'", b.BaseURL)
	}
}
