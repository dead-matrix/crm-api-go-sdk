package crmapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a reusable, concurrency-safe CRM API client.
type Client struct {
	baseURL      string
	staffID      int64
	serviceToken string

	httpClient Doer
	logger     *slog.Logger
	userAgent  string

	requestRetries     int
	retryBaseDelay     time.Duration
	retryMaxDelay      time.Duration
	retryStatusCodes   map[int]struct{}
	retryNonIdempotent bool
	authRefreshLeeway  time.Duration

	tokenCache TokenCache
}

// NewClient constructs a CRM API client from explicit configuration.
func NewClient(cfg Config) (*Client, error) {
	normalized, err := cfg.validateAndNormalize()
	if err != nil {
		return nil, err
	}

	return &Client{
		baseURL:            normalized.BaseURL,
		staffID:            normalized.StaffID,
		serviceToken:       normalized.ServiceToken,
		httpClient:         normalized.HTTPClient,
		logger:             normalized.Logger,
		userAgent:          normalized.UserAgent,
		requestRetries:     normalized.RequestRetries,
		retryBaseDelay:     normalized.RetryBaseDelay,
		retryMaxDelay:      normalized.RetryMaxDelay,
		retryStatusCodes:   makeRetryStatusSet(normalized.RetryStatusCodes),
		retryNonIdempotent: normalized.RetryNonIdempotent,
		authRefreshLeeway:  normalized.AuthRefreshLeeway,
		tokenCache:         normalized.TokenCache,
	}, nil
}

// Close releases idle transport resources when the underlying doer supports it.
func (c *Client) Close() error {
	if closer, ok := c.httpClient.(interface{ CloseIdleConnections() }); ok {
		closer.CloseIdleConnections()
	}
	if closer, ok := c.httpClient.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

func (c *Client) ensureJWT(ctx context.Context) (string, error) {
	key := TokenCacheKey{
		BaseURL: c.baseURL,
		StaffID: c.staffID,
	}

	if cached, ok := c.tokenCache.Get(key); ok && cached.IsValid(c.authRefreshLeeway) {
		return cached.Token, nil
	}

	lock := c.tokenCache.GetLock(key)
	lock.Lock()
	defer lock.Unlock()

	if cached, ok := c.tokenCache.Get(key); ok && cached.IsValid(c.authRefreshLeeway) {
		return cached.Token, nil
	}

	token, exp, err := c.issueJWT(ctx)
	if err != nil {
		return "", err
	}

	c.tokenCache.Set(key, CachedToken{
		Token:     token,
		ExpiresAt: exp,
	})

	return token, nil
}

func (c *Client) refreshJWT(ctx context.Context) (string, error) {
	key := TokenCacheKey{
		BaseURL: c.baseURL,
		StaffID: c.staffID,
	}

	lock := c.tokenCache.GetLock(key)
	lock.Lock()
	defer lock.Unlock()

	token, exp, err := c.issueJWT(ctx)
	if err != nil {
		return "", err
	}

	c.tokenCache.Set(key, CachedToken{
		Token:     token,
		ExpiresAt: exp,
	})

	return token, nil
}

func (c *Client) issueJWT(ctx context.Context) (string, time.Time, error) {
	var out jwtAuthResponse

	err := c.doJSON(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/staff/%d/auth", c.baseURL, c.staffID),
		map[string]string{
			"X-Service-Token": c.serviceToken,
			"Content-Type":    "application/json",
		},
		nil,
		false,
		nil,
		&out,
	)
	if err != nil {
		return "", time.Time{}, err
	}

	if strings.TrimSpace(out.Token) == "" || out.ExpiresAt == nil {
		return "", time.Time{}, &APIError{
			Message: "invalid auth response: missing token or expiry",
		}
	}

	return out.Token, out.ExpiresAt.UTC(), nil
}

func (c *Client) get(ctx context.Context, path string, query map[string]string, needAuth bool, out any) error {
	return c.doJSON(ctx, http.MethodGet, c.baseURL+path, nil, query, needAuth, nil, out)
}

func (c *Client) post(ctx context.Context, path string, query map[string]string, needAuth bool, body any, out any) error {
	return c.doJSON(ctx, http.MethodPost, c.baseURL+path, nil, query, needAuth, body, out)
}

func (c *Client) put(ctx context.Context, path string, query map[string]string, needAuth bool, body any, out any) error {
	return c.doJSON(ctx, http.MethodPut, c.baseURL+path, nil, query, needAuth, body, out)
}

func (c *Client) getFile(ctx context.Context, path string, query map[string]string, needAuth bool) ([]byte, http.Header, error) {
	headers := map[string]string{
		"User-Agent": c.userAgent,
	}
	if needAuth {
		token, err := c.ensureJWT(ctx)
		if err != nil {
			return nil, nil, err
		}
		headers["Authorization"] = "Bearer " + token
		headers["X-Staff-ID"] = fmt.Sprintf("%d", c.staffID)
	}

	fullURL := c.baseURL + path

	resp, err := c.doRequestWithRetry(ctx, http.MethodGet, fullURL, headers, query, nil)
	if err != nil {
		return nil, nil, &HTTPError{
			Message: "file request failed",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && needAuth {
		token, refreshErr := c.refreshJWT(ctx)
		if refreshErr != nil {
			return nil, nil, refreshErr
		}

		headers["Authorization"] = "Bearer " + token

		resp, err = c.doRequestWithRetry(ctx, http.MethodGet, fullURL, headers, query, nil)
		if err != nil {
			return nil, nil, &HTTPError{
				Message: "file request retry after auth refresh failed",
				Cause:   err,
			}
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusOK {
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, &HTTPError{
				Message: "failed to read file response body",
				Status:  resp.StatusCode,
				Cause:   err,
			}
		}
		return content, resp.Header.Clone(), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, &HTTPError{
			Message: "failed to read error response body",
			Status:  resp.StatusCode,
			Cause:   err,
		}
	}

	if err := c.decodeEnvelope(resp.StatusCode, body, nil); err != nil {
		return nil, nil, err
	}

	return nil, nil, &HTTPError{
		Message: "unexpected non-200 response without valid API error envelope",
		Status:  resp.StatusCode,
	}
}

func (c *Client) doJSON(
	ctx context.Context,
	method string,
	fullURL string,
	extraHeaders map[string]string,
	query map[string]string,
	needAuth bool,
	requestBody any,
	out any,
) error {
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   c.userAgent,
	}
	for k, v := range extraHeaders {
		headers[k] = v
	}

	if needAuth {
		token, err := c.ensureJWT(ctx)
		if err != nil {
			return err
		}
		headers["Authorization"] = "Bearer " + token
		headers["X-Staff-ID"] = fmt.Sprintf("%d", c.staffID)
	}

	var bodyBytes []byte
	if requestBody != nil {
		encoded, err := json.Marshal(requestBody)
		if err != nil {
			return &ValidationError{
				Message: fmt.Sprintf("failed to encode request body: %v", err),
			}
		}
		bodyBytes = encoded
	}

	resp, err := c.doRequestWithRetry(ctx, method, fullURL, headers, query, bodyBytes)
	if err != nil {
		return &HTTPError{
			Message: fmt.Sprintf("%s request failed", method),
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && needAuth {
		token, refreshErr := c.refreshJWT(ctx)
		if refreshErr != nil {
			return refreshErr
		}

		headers["Authorization"] = "Bearer " + token

		resp, err = c.doRequestWithRetry(ctx, method, fullURL, headers, query, bodyBytes)
		if err != nil {
			return &HTTPError{
				Message: fmt.Sprintf("%s retry after auth refresh failed", method),
				Cause:   err,
			}
		}
		defer resp.Body.Close()
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPError{
			Message: "failed to read response body",
			Status:  resp.StatusCode,
			Cause:   err,
		}
	}

	return c.decodeEnvelope(resp.StatusCode, rawBody, out)
}

func (c *Client) doRequestWithRetry(
	ctx context.Context,
	method string,
	fullURL string,
	headers map[string]string,
	query map[string]string,
	bodyBytes []byte,
) (*http.Response, error) {
	attempts := c.requestRetries
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		req, err := c.newRequest(ctx, method, fullURL, headers, query, bodyBytes)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err == nil {
			if c.shouldRetryResponse(method, resp.StatusCode) && attempt < attempts {
				delay := c.backoffDelay(attempt)
				c.logRetry("status", method, fullURL, attempt, attempts, delay, resp.StatusCode, nil)
				drainAndClose(resp.Body)
				if err := sleepWithContext(ctx, delay); err != nil {
					return nil, err
				}
				continue
			}
			return resp, nil
		}

		lastErr = err

		if !isRetryableTransportError(err) || attempt == attempts {
			break
		}

		delay := c.backoffDelay(attempt)
		c.logRetry("transport", method, fullURL, attempt, attempts, delay, 0, err)
		if err := sleepWithContext(ctx, delay); err != nil {
			return nil, err
		}
	}

	return nil, lastErr
}

func (c *Client) newRequest(
	ctx context.Context,
	method string,
	fullURL string,
	headers map[string]string,
	query map[string]string,
	bodyBytes []byte,
) (*http.Request, error) {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, &ConfigError{
			Message: fmt.Sprintf("invalid request URL: %v", err),
		}
	}

	q := parsedURL.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	parsedURL.RawQuery = q.Encode()

	var body io.Reader
	if bodyBytes != nil {
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), body)
	if err != nil {
		return nil, &HTTPError{
			Message: "failed to create request",
			Cause:   err,
		}
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func (c *Client) backoffDelay(attempt int) time.Duration {
	base := c.retryBaseDelay
	maxDelay := c.retryMaxDelay

	if base <= 0 {
		base = defaultRetryBaseDelay
	}
	if maxDelay <= 0 {
		maxDelay = defaultRetryMaxDelay
	}

	delay := base * time.Duration(1<<(attempt-1))
	if delay > maxDelay {
		delay = maxDelay
	}

	jitter := time.Duration(rand.Int63n(int64(base)))
	delay += jitter

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

func (c *Client) shouldRetryResponse(method string, status int) bool {
	if _, ok := c.retryStatusCodes[status]; !ok {
		return false
	}
	if c.retryNonIdempotent {
		return true
	}
	return isIdempotentMethod(method)
}

func (c *Client) logRetry(kind, method, fullURL string, attempt, maxAttempts int, delay time.Duration, status int, err error) {
	attrs := []any{
		"kind", kind,
		"method", method,
		"url", sanitizeURLForLog(fullURL),
		"attempt", attempt,
		"max_attempts", maxAttempts,
		"delay", delay,
	}
	if status > 0 {
		attrs = append(attrs, "status", status)
	}
	if err != nil {
		attrs = append(attrs, "error", err)
	}
	c.logger.Warn("retrying CRM API request", attrs...)
}

func isRetryableTransportError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return true
}

func isIdempotentMethod(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 1024))
	_ = body.Close()
}

func sanitizeURLForLog(fullURL string) string {
	parsed, err := url.Parse(fullURL)
	if err != nil {
		return fullURL
	}
	parsed.RawQuery = ""
	return parsed.String()
}

func (c *Client) decodeEnvelope(status int, body []byte, out any) error {
	var raw struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Code    string          `json:"code"`
		Data    json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		return &HTTPError{
			Message: fmt.Sprintf("invalid JSON response: %s", truncateForError(body, 200)),
			Status:  status,
			Cause:   err,
		}
	}

	if raw.Status == "success" {
		if out == nil {
			return nil
		}
		if len(raw.Data) == 0 || string(raw.Data) == "null" {
			return nil
		}
		if err := json.Unmarshal(raw.Data, out); err != nil {
			return &HTTPError{
				Message: "failed to decode API success payload",
				Status:  status,
				Cause:   err,
			}
		}
		return nil
	}

	return c.mapAPIError(status, raw.Message, raw.Code)
}

func (c *Client) mapAPIError(status int, message, code string) error {
	if strings.TrimSpace(message) == "" {
		message = "API error"
	}

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &AuthError{
			Message: message,
			Code:    code,
			Status:  status,
		}
	case status == http.StatusBadRequest || status == http.StatusUnprocessableEntity || strings.EqualFold(code, "VALIDATION_ERROR"):
		return &ValidationError{
			Message: message,
		}
	default:
		return &APIError{
			Message: message,
			Code:    code,
			Status:  status,
		}
	}
}

func makeRetryStatusSet(codes []int) map[int]struct{} {
	set := make(map[int]struct{}, len(codes))
	for _, code := range codes {
		set[code] = struct{}{}
	}
	return set
}

func truncateForError(body []byte, max int) string {
	s := strings.TrimSpace(string(body))
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
