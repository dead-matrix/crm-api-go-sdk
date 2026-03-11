package crmapi

import (
	"sync"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/auth"
)

// TokenCacheKey identifies a cached JWT entry.
type TokenCacheKey struct {
	BaseURL string
	StaffID int64
}

// CachedToken stores a JWT and its expiry time.
type CachedToken struct {
	Token     string
	ExpiresAt time.Time
}

// IsValid reports whether the token remains valid after applying the given leeway.
func (c CachedToken) IsValid(leeway time.Duration) bool {
	if c.Token == "" {
		return false
	}
	return time.Now().UTC().Add(leeway).Before(c.ExpiresAt.UTC())
}

// NewInMemoryTokenCache returns a concurrency-safe in-memory token cache.
func NewInMemoryTokenCache() TokenCache {
	return &tokenCacheAdapter{cache: auth.NewTokenCache()}
}

var defaultTokenCache = NewInMemoryTokenCache()

type tokenCacheAdapter struct {
	cache *auth.TokenCache
}

func (c *tokenCacheAdapter) Get(key TokenCacheKey) (CachedToken, bool) {
	token, ok := c.cache.Get(auth.CacheKey{BaseURL: key.BaseURL, StaffID: key.StaffID})
	if !ok {
		return CachedToken{}, false
	}
	return CachedToken{Token: token.Token, ExpiresAt: token.ExpiresAt}, true
}

func (c *tokenCacheAdapter) Set(key TokenCacheKey, token CachedToken) {
	c.cache.Set(auth.CacheKey{BaseURL: key.BaseURL, StaffID: key.StaffID}, auth.CachedJWT{
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	})
}

func (c *tokenCacheAdapter) GetLock(key TokenCacheKey) sync.Locker {
	return c.cache.GetLock(auth.CacheKey{BaseURL: key.BaseURL, StaffID: key.StaffID})
}
