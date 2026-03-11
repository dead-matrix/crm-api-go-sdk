package crmapi

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout           = 10 * time.Second
	defaultRequestRetries    = 3
	defaultRetryBaseDelay    = 200 * time.Millisecond
	defaultRetryMaxDelay     = 1 * time.Second
	defaultAuthRefreshLeeway = 30 * time.Second
	defaultUserAgent         = "crm-api-go-sdk"
)

var defaultRetryStatusCodes = []int{
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
}

// Doer executes HTTP requests.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// TokenCache stores JWTs keyed by CRM base URL and staff ID.
type TokenCache interface {
	Get(TokenCacheKey) (CachedToken, bool)
	Set(TokenCacheKey, CachedToken)
	GetLock(TokenCacheKey) sync.Locker
}

// Config configures CRM API client.
type Config struct {
	BaseURL      string
	StaffID      int64
	ServiceToken string

	HTTPClient Doer
	Logger     *slog.Logger
	TokenCache TokenCache
	UserAgent  string

	Timeout            time.Duration
	RequestRetries     int
	RetryBaseDelay     time.Duration
	RetryMaxDelay      time.Duration
	RetryStatusCodes   []int
	RetryNonIdempotent bool
	AuthRefreshLeeway  time.Duration
}

func (c Config) validateAndNormalize() (Config, error) {
	cfg := c

	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	cfg.ServiceToken = strings.TrimSpace(cfg.ServiceToken)
	cfg.UserAgent = strings.TrimSpace(cfg.UserAgent)

	if cfg.BaseURL == "" {
		return Config{}, &ConfigError{Message: "base URL must be provided"}
	}
	if cfg.StaffID <= 0 {
		return Config{}, &ConfigError{Message: "staff ID must be a positive integer"}
	}
	if cfg.ServiceToken == "" {
		return Config{}, &ConfigError{Message: "service token must be provided"}
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	if cfg.RequestRetries <= 0 {
		cfg.RequestRetries = defaultRequestRetries
	}
	if cfg.RetryBaseDelay <= 0 {
		cfg.RetryBaseDelay = defaultRetryBaseDelay
	}
	if cfg.RetryMaxDelay <= 0 {
		cfg.RetryMaxDelay = defaultRetryMaxDelay
	}
	if cfg.AuthRefreshLeeway <= 0 {
		cfg.AuthRefreshLeeway = defaultAuthRefreshLeeway
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = defaultUserAgent
	}

	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.TokenCache == nil {
		cfg.TokenCache = defaultTokenCache
	}
	if len(cfg.RetryStatusCodes) == 0 {
		cfg.RetryStatusCodes = append([]int(nil), defaultRetryStatusCodes...)
	} else {
		cfg.RetryStatusCodes = normalizeRetryStatusCodes(cfg.RetryStatusCodes)
	}

	return cfg, nil
}

func normalizeRetryStatusCodes(codes []int) []int {
	seen := make(map[int]struct{}, len(codes))
	normalized := make([]int, 0, len(codes))

	for _, code := range codes {
		if code < 100 || code > 599 {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, code)
	}

	return normalized
}
