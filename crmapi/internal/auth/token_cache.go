package auth

import (
	"sync"
	"time"
)

type CacheKey struct {
	BaseURL string
	StaffID int64
}

type CachedJWT struct {
	Token     string
	ExpiresAt time.Time
}

func (c CachedJWT) IsValid(leeway time.Duration) bool {
	if c.Token == "" {
		return false
	}
	return time.Now().UTC().Add(leeway).Before(c.ExpiresAt.UTC())
}

type TokenCache struct {
	mu    sync.RWMutex
	data  map[CacheKey]CachedJWT
	locks map[CacheKey]*sync.Mutex
}

func NewTokenCache() *TokenCache {
	return &TokenCache{
		data:  make(map[CacheKey]CachedJWT),
		locks: make(map[CacheKey]*sync.Mutex),
	}
}

func (c *TokenCache) Get(key CacheKey) (CachedJWT, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	token, ok := c.data[key]
	return token, ok
}

func (c *TokenCache) Set(key CacheKey, token CachedJWT) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = token
}

func (c *TokenCache) GetLock(key CacheKey) *sync.Mutex {
	c.mu.Lock()
	defer c.mu.Unlock()

	if lock, ok := c.locks[key]; ok {
		return lock
	}

	lock := &sync.Mutex{}
	c.locks[key] = lock
	return lock
}
