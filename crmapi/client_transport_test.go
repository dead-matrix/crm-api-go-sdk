package crmapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func successResponse(r *http.Request) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"status":"success","data":null}`)),
		Header:     make(http.Header),
		Request:    r,
	}
}

func TestDoRequestWithRetryRetriesTransportErrors(t *testing.T) {
	var attempts int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"x":1}` {
			t.Fatalf("attempt %d body = %q", attempts, string(body))
		}
		if attempts < 3 {
			return nil, errors.New("temporary transport failure")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"status":"success","data":null}`)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	client := mustNewClient(t, "https://example.test", &http.Client{Transport: transport})
	resp, err := client.doRequestWithRetry(context.Background(), http.MethodPost, "https://example.test/echo", nil, nil, []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("doRequestWithRetry() error = %v", err)
	}
	defer resp.Body.Close()
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestDoRequestWithRetryRetriesConfiguredStatusCodesForIdempotentRequests(t *testing.T) {
	var attempts int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader(`temporary outage`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}
		return successResponse(r), nil
	})

	client := mustNewClient(t, "https://example.test", &http.Client{Transport: transport})
	resp, err := client.doRequestWithRetry(context.Background(), http.MethodGet, "https://example.test/ping", nil, nil, nil)
	if err != nil {
		t.Fatalf("doRequestWithRetry() error = %v", err)
	}
	defer resp.Body.Close()
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestDoRequestWithRetryDoesNotRetryStatusCodesForNonIdempotentRequestsByDefault(t *testing.T) {
	var attempts int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader(`temporary outage`)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	client := mustNewClient(t, "https://example.test", &http.Client{Transport: transport})
	resp, err := client.doRequestWithRetry(context.Background(), http.MethodPost, "https://example.test/ping", nil, nil, []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("doRequestWithRetry() error = %v", err)
	}
	defer resp.Body.Close()
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestDoRequestWithRetryRetriesNonIdempotentRequestsWhenEnabled(t *testing.T) {
	var attempts int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return &http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Body:       io.NopCloser(strings.NewReader(`temporary outage`)),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}
		return successResponse(r), nil
	})

	client, err := NewClient(Config{
		BaseURL:            "https://example.test",
		StaffID:            123,
		ServiceToken:       "svc-token",
		HTTPClient:         &http.Client{Transport: transport},
		TokenCache:         NewInMemoryTokenCache(),
		RequestRetries:     3,
		RetryBaseDelay:     time.Millisecond,
		RetryMaxDelay:      time.Millisecond,
		RetryNonIdempotent: true,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.doRequestWithRetry(context.Background(), http.MethodPost, "https://example.test/ping", nil, nil, []byte(`{"x":1}`))
	if err != nil {
		t.Fatalf("doRequestWithRetry() error = %v", err)
	}
	defer resp.Body.Close()
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestClientSendsDefaultUserAgent(t *testing.T) {
	var gotUserAgent string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotUserAgent = r.Header.Get("User-Agent")
		return successResponse(r), nil
	})

	client := mustNewClient(t, "https://example.test", &http.Client{Transport: transport})
	if err := client.doJSON(context.Background(), http.MethodGet, "https://example.test/ping", nil, nil, false, nil, nil); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if gotUserAgent != defaultUserAgent {
		t.Fatalf("User-Agent = %q, want %q", gotUserAgent, defaultUserAgent)
	}
}

func TestClientSendsConfiguredUserAgent(t *testing.T) {
	var gotUserAgent string
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotUserAgent = r.Header.Get("User-Agent")
		return successResponse(r), nil
	})

	client, err := NewClient(Config{
		BaseURL:        "https://example.test",
		StaffID:        123,
		ServiceToken:   "svc-token",
		HTTPClient:     &http.Client{Transport: transport},
		TokenCache:     NewInMemoryTokenCache(),
		UserAgent:      "traffsoft-crm-sdk/1.0",
		RequestRetries: 1,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if err := client.doJSON(context.Background(), http.MethodGet, "https://example.test/ping", nil, nil, false, nil, nil); err != nil {
		t.Fatalf("doJSON() error = %v", err)
	}
	if gotUserAgent != "traffsoft-crm-sdk/1.0" {
		t.Fatalf("User-Agent = %q, want configured value", gotUserAgent)
	}
}

func TestDecodeEnvelopeMapsTypedErrors(t *testing.T) {
	client := mustNewClient(t, "https://example.test", &http.Client{})
	cases := []struct {
		name   string
		status int
		body   string
		check  func(error) bool
	}{
		{"auth", http.StatusUnauthorized, `{"status":"error","message":"nope","code":"AUTH"}`, func(err error) bool { var target *AuthError; return errors.As(err, &target) }},
		{"validation", http.StatusBadRequest, `{"status":"error","message":"bad","code":"VALIDATION_ERROR"}`, func(err error) bool { var target *ValidationError; return errors.As(err, &target) }},
		{"api", http.StatusInternalServerError, `{"status":"error","message":"boom","code":"SERVER"}`, func(err error) bool { var target *APIError; return errors.As(err, &target) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.decodeEnvelope(tc.status, []byte(tc.body), nil)
			if !tc.check(err) {
				t.Fatalf("error %T did not match expected type", err)
			}
		})
	}
}

func TestTasksLogDownloadsFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/tasks/log":
			if got := r.Header.Get("Authorization"); got != "Bearer jwt-1" {
				t.Fatalf("unexpected authorization: %q", got)
			}
			w.Header().Set("Content-Disposition", `attachment; filename="task.log"`)
			_, _ = io.Copy(w, bytes.NewBufferString("hello log"))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.TasksLog(context.Background(), 1, "mailing", 2, 3)
	if err != nil {
		t.Fatalf("TasksLog() error = %v", err)
	}
	if string(res.Content) != "hello log" {
		t.Fatalf("content = %q", string(res.Content))
	}
	if res.Filename == nil || *res.Filename != "task.log" {
		t.Fatalf("filename = %v, want task.log", res.Filename)
	}
}
