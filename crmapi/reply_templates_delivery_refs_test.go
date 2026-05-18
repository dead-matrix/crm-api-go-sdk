package crmapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// ReplyTemplatesDeliveryRefsList (GET /api/reply-templates/{id}/delivery-refs)
// ----------------------------------------------------------------------------

func TestReplyTemplatesDeliveryRefsList_Empty(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42/delivery-refs":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			capturedQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":{"refs":[]}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	refs, err := client.ReplyTemplatesDeliveryRefsList(
		context.Background(), 42, "telegram", "tg_crm_bot",
	)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(refs) != 0 {
		t.Fatalf("want 0 refs, got %d", len(refs))
	}
	if got := capturedQuery.Get("provider"); got != "telegram" {
		t.Fatalf("provider query = %q", got)
	}
	if got := capturedQuery.Get("providerScope"); got != "tg_crm_bot" {
		t.Fatalf("providerScope query = %q", got)
	}
}

func TestReplyTemplatesDeliveryRefsList_MapsRows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/9/delivery-refs":
			fmt.Fprint(w, `{"status":"success","data":{"refs":[
				{
					"id": 1, "itemId": 10, "provider": "telegram",
					"providerScope": "tg_crm_bot",
					"mediaRef": "BAACAg-photo-id",
					"mediaUniqueRef": "AgADxx", "mediaType": "photo",
					"failCount": 0,
					"lastUsedAt": "2026-05-18T12:00:00",
					"createdAt": "2026-05-18T11:00:00",
					"updatedAt": "2026-05-18T12:00:00"
				},
				{
					"id": 2, "itemId": 11, "provider": "telegram",
					"providerScope": "tg_crm_bot",
					"mediaRef": "BAACAg-video-id",
					"mediaUniqueRef": null, "mediaType": "video",
					"failCount": 1,
					"lastUsedAt": null, "createdAt": null, "updatedAt": null
				}
			]}}`)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	refs, err := client.ReplyTemplatesDeliveryRefsList(
		context.Background(), 9, "telegram", "tg_crm_bot",
	)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("want 2 refs, got %d", len(refs))
	}
	if refs[0].ItemID != 10 || refs[0].MediaRef != "BAACAg-photo-id" {
		t.Fatalf("refs[0] = %+v", refs[0])
	}
	if refs[1].MediaUniqueRef != nil {
		t.Fatalf("refs[1].mediaUniqueRef should be nil for null wire value")
	}
	if refs[1].FailCount != 1 {
		t.Fatalf("refs[1].failCount = %d, want 1", refs[1].FailCount)
	}
}

func TestReplyTemplatesDeliveryRefsList_ValidatesArgsBeforeRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Server should NOT be hit for invalid client args.
		if !strings.HasSuffix(r.URL.Path, "/auth") {
			t.Fatalf("server reached despite invalid client args: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())

	for _, tc := range []struct {
		name              string
		templateID        int64
		provider          string
		providerScope     string
		expectFragment    string
	}{
		{"zero templateID", 0, "telegram", "x", "template_id"},
		{"empty provider", 1, "", "x", "provider"},
		{"whitespace provider", 1, "   ", "x", "provider"},
		{"empty scope", 1, "telegram", "", "providerScope"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ReplyTemplatesDeliveryRefsList(
				context.Background(), tc.templateID, tc.provider, tc.providerScope,
			)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("want ValidationError, got %T: %v", err, err)
			}
			if !strings.Contains(ve.Message, tc.expectFragment) {
				t.Fatalf("ValidationError message %q missing fragment %q", ve.Message, tc.expectFragment)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplatesDeliveryRefsUpsert (PUT)
// ----------------------------------------------------------------------------

func TestReplyTemplatesDeliveryRefsUpsert_SendsCamelCasePayload(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/5/delivery-refs":
			if r.Method != http.MethodPut {
				t.Fatalf("method = %s, want PUT", r.Method)
			}
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &capturedBody); err != nil {
				t.Fatalf("decode body: %v body=%s", err, string(body))
			}
			fmt.Fprint(w, `{"status":"success","data":{"refs":[
				{"id":1,"itemId":11,"provider":"telegram","providerScope":"tg_bot","mediaRef":"BAA","mediaType":"photo","failCount":0,
				 "lastUsedAt":null,"createdAt":null,"updatedAt":null}
			]}}`)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())

	mediaType := "photo"
	refs, err := client.ReplyTemplatesDeliveryRefsUpsert(
		context.Background(),
		5,
		UpsertDeliveryRefsInput{
			Provider:      "telegram",
			ProviderScope: "tg_bot",
			Refs: []UpsertDeliveryRefInput{
				{ItemID: 11, MediaRef: "BAA", MediaType: &mediaType},
			},
		},
	)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Wire body is strictly camelCase.
	if capturedBody["provider"] != "telegram" {
		t.Fatalf("wire provider = %v", capturedBody["provider"])
	}
	if capturedBody["providerScope"] != "tg_bot" {
		t.Fatalf("wire providerScope = %v", capturedBody["providerScope"])
	}
	refsRaw, _ := capturedBody["refs"].([]any)
	if len(refsRaw) != 1 {
		t.Fatalf("wire refs length = %d", len(refsRaw))
	}
	first := refsRaw[0].(map[string]any)
	if first["itemId"].(float64) != 11 || first["mediaRef"] != "BAA" {
		t.Fatalf("wire refs[0] = %+v", first)
	}

	if len(refs) != 1 || refs[0].ID != 1 || refs[0].ItemID != 11 {
		t.Fatalf("response refs = %+v", refs)
	}
}

func TestReplyTemplatesDeliveryRefsUpsert_ValidatesArgsBeforeRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/auth") {
			t.Fatalf("server hit despite invalid input: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
	}))
	defer server.Close()
	client := mustNewClient(t, server.URL, server.Client())

	for _, tc := range []struct {
		name     string
		tid      int64
		in       UpsertDeliveryRefsInput
		fragment string
	}{
		{
			"zero templateID",
			0,
			UpsertDeliveryRefsInput{Provider: "telegram", ProviderScope: "x", Refs: []UpsertDeliveryRefInput{{ItemID: 1, MediaRef: "ok"}}},
			"template_id",
		},
		{
			"empty provider",
			1,
			UpsertDeliveryRefsInput{Provider: " ", ProviderScope: "x", Refs: []UpsertDeliveryRefInput{{ItemID: 1, MediaRef: "ok"}}},
			"provider",
		},
		{
			"empty scope",
			1,
			UpsertDeliveryRefsInput{Provider: "telegram", ProviderScope: "  ", Refs: []UpsertDeliveryRefInput{{ItemID: 1, MediaRef: "ok"}}},
			"providerScope",
		},
		{
			"empty refs",
			1,
			UpsertDeliveryRefsInput{Provider: "telegram", ProviderScope: "x", Refs: nil},
			"refs",
		},
		{
			"too many refs",
			1,
			UpsertDeliveryRefsInput{
				Provider: "telegram", ProviderScope: "x",
				Refs: func() []UpsertDeliveryRefInput {
					out := make([]UpsertDeliveryRefInput, 0, 11)
					for i := 0; i < 11; i++ {
						out = append(out, UpsertDeliveryRefInput{ItemID: int64(i + 1), MediaRef: "ok"})
					}
					return out
				}(),
			},
			"at most",
		},
		{
			"zero itemId",
			1,
			UpsertDeliveryRefsInput{Provider: "telegram", ProviderScope: "x", Refs: []UpsertDeliveryRefInput{{ItemID: 0, MediaRef: "ok"}}},
			"itemId",
		},
		{
			"empty mediaRef",
			1,
			UpsertDeliveryRefsInput{Provider: "telegram", ProviderScope: "x", Refs: []UpsertDeliveryRefInput{{ItemID: 1, MediaRef: "  "}}},
			"mediaRef",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ReplyTemplatesDeliveryRefsUpsert(context.Background(), tc.tid, tc.in)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("want ValidationError, got %T: %v", err, err)
			}
			if !strings.Contains(ve.Message, tc.fragment) {
				t.Fatalf("message %q missing %q", ve.Message, tc.fragment)
			}
		})
	}
}

func TestReplyTemplatesDeliveryRefsUpsert_MapsServerValidationError(t *testing.T) {
	// 422 from the server (cross-template item id) must surface as
	// ValidationError, not APIError - clients differentiate retry
	// strategy by error type.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/5/delivery-refs":
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, `{"status":"error","message":"item ids do not belong to template 5","code":"validation_error"}`)
		}
	}))
	defer server.Close()
	client := mustNewClient(t, server.URL, server.Client())

	_, err := client.ReplyTemplatesDeliveryRefsUpsert(
		context.Background(), 5,
		UpsertDeliveryRefsInput{
			Provider: "telegram", ProviderScope: "tg_bot",
			Refs: []UpsertDeliveryRefInput{{ItemID: 99, MediaRef: "BAA"}},
		},
	)
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplateItem.ID propagation (full GET)
// ----------------------------------------------------------------------------

func TestReplyTemplatesGet_PropagatesItemIDs(t *testing.T) {
	// The messenger keys delivery-ref upserts by item id; the SDK
	// must surface it from the wire envelope. Older response payloads
	// did not include `id` - the new server always does.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 42, "publicId": "uuid-42", "title": "alb", "kind": "album",
				"creator": {"employeeId": 5, "name": "Olga"},
				"items": [
					{"id": 100, "position": 0, "type": "photo",
					 "mediaObjectKey": "shared/reply-templates/uuid-42/0.jpg"},
					{"id": 101, "position": 1, "type": "video",
					 "mediaObjectKey": "shared/reply-templates/uuid-42/1.mp4"},
					{"id": 102, "position": 2, "type": "gif",
					 "mediaObjectKey": "shared/reply-templates/uuid-42/2.gif"}
				]
			}}`)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	full, err := client.ReplyTemplatesGet(context.Background(), 42)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(full.Items) != 3 {
		t.Fatalf("items len: %d", len(full.Items))
	}
	wantIDs := []int64{100, 101, 102}
	for i, want := range wantIDs {
		if full.Items[i].ID != want {
			t.Fatalf("items[%d].id = %d, want %d", i, full.Items[i].ID, want)
		}
	}
}
