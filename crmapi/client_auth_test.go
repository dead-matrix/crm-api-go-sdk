package crmapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func mustNewClient(t *testing.T, baseURL string, httpClient Doer) *Client {
	t.Helper()
	c, err := NewClient(Config{
		BaseURL:        baseURL,
		StaffID:        123,
		ServiceToken:   "svc-token",
		HTTPClient:     httpClient,
		TokenCache:     NewInMemoryTokenCache(),
		RequestRetries: 3,
		RetryBaseDelay: time.Millisecond,
		RetryMaxDelay:  time.Millisecond,
		Timeout:        time.Second,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c
}

type stubTokenCache struct {
	mu     sync.Mutex
	tokens map[TokenCacheKey]CachedToken
	locks  map[TokenCacheKey]*sync.Mutex
}

func newStubTokenCache() *stubTokenCache {
	return &stubTokenCache{
		tokens: make(map[TokenCacheKey]CachedToken),
		locks:  make(map[TokenCacheKey]*sync.Mutex),
	}
}

func (c *stubTokenCache) Get(key TokenCacheKey) (CachedToken, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	token, ok := c.tokens[key]
	return token, ok
}

func (c *stubTokenCache) Set(key TokenCacheKey, token CachedToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokens[key] = token
}

func (c *stubTokenCache) GetLock(key TokenCacheKey) sync.Locker {
	c.mu.Lock()
	defer c.mu.Unlock()
	if lock, ok := c.locks[key]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	c.locks[key] = lock
	return lock
}

func TestClientIssueJWT(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/staff/123/auth" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("X-Service-Token"); got != "svc-token" {
			t.Fatalf("unexpected service token: %q", got)
		}
		fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	token, exp, err := client.issueJWT(context.Background())
	if err != nil {
		t.Fatalf("issueJWT() error = %v", err)
	}
	if token != "jwt-1" {
		t.Fatalf("token = %q, want jwt-1", token)
	}
	if exp.Format(time.RFC3339) != expiresAt {
		t.Fatalf("expires_at = %s, want %s", exp.Format(time.RFC3339), expiresAt)
	}
}

func TestClientRefreshesJWTOn401(t *testing.T) {
	var authCalls, userCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			authCalls++
			token := fmt.Sprintf("jwt-%d", authCalls)
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"%s","expires_at":"%s"}}`, token, expiresAt)
		case "/api/users/42":
			userCalls++
			if got := r.Header.Get("Authorization"); got == "Bearer jwt-1" {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"status":"error","message":"expired","code":"AUTH_ERROR"}`)
				return
			}
			fmt.Fprint(w, `{"status":"success","data":{"user_id":42,"bots_info":[]}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if user.UserID != 42 {
		t.Fatalf("user_id = %d, want 42", user.UserID)
	}
	if authCalls != 2 {
		t.Fatalf("authCalls = %d, want 2", authCalls)
	}
	if userCalls != 2 {
		t.Fatalf("userCalls = %d, want 2", userCalls)
	}
}

func TestClientUsesInjectedTokenCache(t *testing.T) {
	cache := newStubTokenCache()
	cache.Set(TokenCacheKey{BaseURL: "https://example.test", StaffID: 123}, CachedToken{
		Token:     "cached-jwt",
		ExpiresAt: time.Now().Add(time.Hour),
	})

	client, err := NewClient(Config{
		BaseURL:      "https://example.test",
		StaffID:      123,
		ServiceToken: "svc-token",
		TokenCache:   cache,
		HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/api/staff/123/auth" {
				t.Fatal("auth endpoint should not be called when token cache already has a valid token")
			}
			if got := r.Header.Get("Authorization"); got != "Bearer cached-jwt" {
				t.Fatalf("Authorization = %q, want Bearer cached-jwt", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"status":"success","data":{"user_id":42,"bots_info":[]}}`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		})},
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	user, err := client.GetUser(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if user.UserID != 42 {
		t.Fatalf("user_id = %d, want 42", user.UserID)
	}
}
