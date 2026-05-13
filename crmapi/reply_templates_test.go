package crmapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// ReplyTemplatesList (GET /api/reply-templates)
// ----------------------------------------------------------------------------

func TestReplyTemplatesListMapsRows(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			capturedQuery = r.URL.Query()
			fmt.Fprint(w, `{"status":"success","data":[
				{
					"id": 10,
					"publicId": "uuid-a",
					"title": "Приветствие",
					"kind": "single",
					"creator": {"employeeId": 5, "name": "Olga"},
					"preview": {"firstItemType": "text", "captionExcerpt": "Привет!", "itemsCount": 1},
					"usageCount": 7,
					"lastUsedAt": "2026-05-10T12:00:00",
					"createdAt": "2026-05-01T10:00:00",
					"updatedAt": "2026-05-01T10:00:00"
				},
				{
					"id": 11,
					"publicId": "uuid-b",
					"title": "Альбом",
					"kind": "album",
					"creator": {"employeeId": 0, "name": "TraffSoft"},
					"preview": {"firstItemType": "photo", "captionExcerpt": "look", "itemsCount": 3},
					"usageCount": 0,
					"lastUsedAt": null,
					"createdAt": null,
					"updatedAt": null
				}
			]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	items, err := client.ReplyTemplatesList(context.Background(), 50, 10)
	if err != nil {
		t.Fatalf("ReplyTemplatesList() error = %v", err)
	}

	if got := capturedQuery.Get("limit"); got != "50" {
		t.Fatalf("limit query = %q, want 50", got)
	}
	if got := capturedQuery.Get("offset"); got != "10" {
		t.Fatalf("offset query = %q, want 10", got)
	}

	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	a := items[0]
	if a.ID != 10 || a.PublicID != "uuid-a" || a.Title != "Приветствие" || a.Kind != "single" {
		t.Fatalf("row[0] = %+v", a)
	}
	if a.Creator.EmployeeID != 5 || a.Creator.Name == nil || *a.Creator.Name != "Olga" {
		t.Fatalf("row[0].Creator = %+v (name=%v)", a.Creator, a.Creator.Name)
	}
	if a.Preview.FirstItemType == nil || *a.Preview.FirstItemType != "text" {
		t.Fatalf("row[0].Preview.FirstItemType = %v", a.Preview.FirstItemType)
	}
	if a.Preview.ItemsCount != 1 {
		t.Fatalf("row[0].Preview.ItemsCount = %d, want 1", a.Preview.ItemsCount)
	}
	if a.UsageCount != 7 {
		t.Fatalf("row[0].UsageCount = %d, want 7", a.UsageCount)
	}
	if a.LastUsedAt == nil {
		t.Fatalf("row[0].LastUsedAt should be parsed")
	}

	b := items[1]
	if b.Creator.EmployeeID != 0 || b.Creator.Name == nil || *b.Creator.Name != "TraffSoft" {
		t.Fatalf("row[1].Creator = %+v (name=%v) — system creator must surface as TraffSoft", b.Creator, b.Creator.Name)
	}
	if b.LastUsedAt != nil {
		t.Fatalf("row[1].LastUsedAt = %v, want nil", b.LastUsedAt)
	}
}

func TestReplyTemplatesListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates":
			fmt.Fprint(w, `{"status":"success","data":[]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	items, err := client.ReplyTemplatesList(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ReplyTemplatesList() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list, got %d", len(items))
	}
}

func TestReplyTemplatesListNegativeValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called on validation failure: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	for _, tc := range []struct {
		name           string
		limit          int64
		offset         int64
	}{
		{"limit negative", -1, 0},
		{"offset negative", 0, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ReplyTemplatesList(context.Background(), tc.limit, tc.offset)
			var ve *ValidationError
			if !errorsAsValidation(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplatesGet (GET /api/reply-templates/{id})
// ----------------------------------------------------------------------------

func TestReplyTemplatesGetMapsAlbum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 42,
				"publicId": "uuid-42",
				"title": "Альбом",
				"kind": "album",
				"creator": {"employeeId": 5, "name": "Olga"},
				"items": [
					{"position":0,"type":"photo","caption":"header",
					 "mediaObjectKey":"shared/reply-templates/uuid-42/0.jpg",
					 "mime":"image/jpeg","sizeBytes":248123,
					 "width":1280,"height":960,
					 "originTenantId":"t1","originMessageId":"m9"},
					{"position":1,"type":"video","mediaObjectKey":"shared/reply-templates/uuid-42/1.mp4",
					 "mime":"video/mp4","sizeBytes":9999999,"durationMs":15000,
					 "width":1920,"height":1080},
					{"position":2,"type":"gif","mediaObjectKey":"shared/reply-templates/uuid-42/2.gif"}
				],
				"createdAt": "2026-05-13T12:00:00",
				"updatedAt": "2026-05-13T12:00:00"
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	full, err := client.ReplyTemplatesGet(context.Background(), 42)
	if err != nil {
		t.Fatalf("ReplyTemplatesGet() error = %v", err)
	}
	if full.ID != 42 || full.PublicID != "uuid-42" || full.Kind != "album" {
		t.Fatalf("envelope = %+v", full)
	}
	if len(full.Items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(full.Items))
	}
	if full.Items[0].Caption == nil || *full.Items[0].Caption != "header" {
		t.Fatalf("items[0].Caption = %v, want pointer to \"header\"", full.Items[0].Caption)
	}
	if full.Items[1].DurationMs == nil || *full.Items[1].DurationMs != 15000 {
		t.Fatalf("items[1].DurationMs = %v", full.Items[1].DurationMs)
	}
	if full.Items[2].Caption != nil {
		t.Fatalf("items[2].Caption = %v, want nil", full.Items[2].Caption)
	}
	if full.Items[0].OriginMessageID == nil || *full.Items[0].OriginMessageID != "m9" {
		t.Fatalf("items[0].OriginMessageID = %v", full.Items[0].OriginMessageID)
	}
}

func TestReplyTemplatesGetMapsSingleText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/7":
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 7, "publicId": "uuid-7", "title": "Hi",
				"kind": "single",
				"creator": {"employeeId": 0, "name": "TraffSoft"},
				"items": [{"position":0,"type":"text","caption":"Привет!"}],
				"createdAt": null, "updatedAt": null
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	full, err := client.ReplyTemplatesGet(context.Background(), 7)
	if err != nil {
		t.Fatalf("ReplyTemplatesGet() error = %v", err)
	}
	if full.Kind != "single" || len(full.Items) != 1 {
		t.Fatalf("envelope = %+v", full)
	}
	it := full.Items[0]
	if it.Type != "text" || it.Caption == nil || *it.Caption != "Привет!" {
		t.Fatalf("item = %+v", it)
	}
	if it.MediaObjectKey != nil {
		t.Fatalf("text item must have nil MediaObjectKey, got %v", it.MediaObjectKey)
	}
}

func TestReplyTemplatesGetValidatesID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	for _, tid := range []int64{0, -1, -100} {
		_, err := client.ReplyTemplatesGet(context.Background(), tid)
		var ve *ValidationError
		if !errorsAsValidation(err, &ve) {
			t.Fatalf("id=%d: expected ValidationError, got %v", tid, err)
		}
	}
}

func TestReplyTemplatesGet404PropagatesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/999":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"status":"error","message":"Reply template not found"}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.ReplyTemplatesGet(context.Background(), 999)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := err.(*APIError); !ok {
		t.Fatalf("expected *APIError, got %T (%v)", err, err)
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplatesCreate (POST /api/reply-templates)
// ----------------------------------------------------------------------------

func TestReplyTemplatesCreateSendsBodyAndMapsResponse(t *testing.T) {
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			capturedBody, _ = io.ReadAll(r.Body)
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 101,
				"publicId": "new-uuid",
				"title": "hi",
				"kind": "single",
				"creator": {"employeeId": 123, "name": "Tester"},
				"items": [{"position":0,"type":"text","caption":"hello"}],
				"createdAt": "2026-05-13T15:00:00",
				"updatedAt": "2026-05-13T15:00:00"
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	caption := "hello"
	full, err := client.ReplyTemplatesCreate(context.Background(), CreateReplyTemplateInput{
		Title: "hi",
		Kind:  ReplyTemplateKindSingle,
		Items: []ReplyTemplateItem{
			{Position: 0, Type: ReplyTemplateItemTypeText, Caption: &caption},
		},
	})
	if err != nil {
		t.Fatalf("ReplyTemplatesCreate() error = %v", err)
	}

	// Body must be camelCase JSON exactly as CRM expects.
	if !strings.Contains(string(capturedBody), `"title":"hi"`) {
		t.Fatalf("body missing title: %s", capturedBody)
	}
	if !strings.Contains(string(capturedBody), `"kind":"single"`) {
		t.Fatalf("body missing kind: %s", capturedBody)
	}
	if !strings.Contains(string(capturedBody), `"caption":"hello"`) {
		t.Fatalf("body missing caption: %s", capturedBody)
	}
	if !strings.Contains(string(capturedBody), `"type":"text"`) {
		t.Fatalf("body missing item type: %s", capturedBody)
	}

	if full.ID != 101 || full.PublicID != "new-uuid" {
		t.Fatalf("envelope = %+v", full)
	}
	if full.CreatedAt == nil {
		t.Fatalf("CreatedAt should be parsed")
	}
}

func TestReplyTemplatesCreateAlbumPayloadShape(t *testing.T) {
	var rawBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates":
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &rawBody)
			fmt.Fprint(w, `{"status":"success","data":{
				"id": 1, "publicId": "p", "title": "x", "kind": "album",
				"creator": {"employeeId": 1, "name": null},
				"items": [], "createdAt": null, "updatedAt": null
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	caption := "header"
	mok0 := "shared/x/0.jpg"
	mok1 := "shared/x/1.mp4"
	mok2 := "shared/x/2.gif"
	_, err := client.ReplyTemplatesCreate(context.Background(), CreateReplyTemplateInput{
		Title: "x", Kind: ReplyTemplateKindAlbum,
		Items: []ReplyTemplateItem{
			{Position: 0, Type: ReplyTemplateItemTypePhoto, Caption: &caption, MediaObjectKey: &mok0},
			{Position: 1, Type: ReplyTemplateItemTypeVideo, MediaObjectKey: &mok1},
			{Position: 2, Type: ReplyTemplateItemTypeGIF, MediaObjectKey: &mok2},
		},
	})
	if err != nil {
		t.Fatalf("ReplyTemplatesCreate(album) error = %v", err)
	}
	items, ok := rawBody["items"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("body.items = %v", rawBody["items"])
	}
	first := items[0].(map[string]any)
	if first["mediaObjectKey"] != "shared/x/0.jpg" {
		t.Fatalf("items[0].mediaObjectKey = %v", first["mediaObjectKey"])
	}
}

// ----------------------------------------------------------------------------
// Client-side validation for ReplyTemplatesCreate (HTTP must NOT be reached).
// ----------------------------------------------------------------------------

func TestReplyTemplatesCreateValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	caption := "hi"
	mok := "shared/x/0.bin"

	cases := []struct {
		name  string
		input CreateReplyTemplateInput
	}{
		{
			name: "empty title",
			input: CreateReplyTemplateInput{
				Title: "  ", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{{Position: 0, Type: ReplyTemplateItemTypeText, Caption: &caption}},
			},
		},
		{
			name: "unknown kind",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: "compound",
				Items: []ReplyTemplateItem{{Position: 0, Type: ReplyTemplateItemTypeText, Caption: &caption}},
			},
		},
		{
			name: "single with two items",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypeText, Caption: &caption},
					{Position: 1, Type: ReplyTemplateItemTypeText, Caption: &caption},
				},
			},
		},
		{
			name: "album with file",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 1, Type: ReplyTemplateItemTypeFile, MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "album with voice",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 1, Type: ReplyTemplateItemTypeVoice, MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "album caption on position 1",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 1, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok, Caption: &caption},
				},
			},
		},
		{
			name: "text item with media key",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypeText, Caption: &caption, MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "text item missing caption",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypeText},
				},
			},
		},
		{
			name: "photo without media",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto},
				},
			},
		},
		{
			name: "voice with caption",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypeVoice, MediaObjectKey: &mok, Caption: &caption},
				},
			},
		},
		{
			name: "unknown type",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindSingle,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: "screenshot", MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "duplicate positions",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "positions not contiguous",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: []ReplyTemplateItem{
					{Position: 0, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 1, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
					{Position: 3, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok},
				},
			},
		},
		{
			name: "11 items",
			input: CreateReplyTemplateInput{
				Title: "t", Kind: ReplyTemplateKindAlbum,
				Items: func() []ReplyTemplateItem {
					out := make([]ReplyTemplateItem, 11)
					for i := 0; i < 11; i++ {
						out[i] = ReplyTemplateItem{Position: i, Type: ReplyTemplateItemTypePhoto, MediaObjectKey: &mok}
					}
					return out
				}(),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.ReplyTemplatesCreate(context.Background(), tc.input)
			var ve *ValidationError
			if !errorsAsValidation(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ReplyTemplatesDelete (DELETE /api/reply-templates/{id})
// ----------------------------------------------------------------------------

func TestReplyTemplatesDeleteMapsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			if r.Method != http.MethodDelete {
				t.Fatalf("method = %s, want DELETE", r.Method)
			}
			fmt.Fprint(w, `{"status":"success","data":{"id": 42, "publicId": "uuid-42"}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ReplyTemplatesDelete(context.Background(), 42)
	if err != nil {
		t.Fatalf("ReplyTemplatesDelete() error = %v", err)
	}
	if res.ID != 42 || res.PublicID != "uuid-42" {
		t.Fatalf("res = %+v", res)
	}
}

func TestReplyTemplatesDeleteValidatesID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("HTTP must not be called: %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.ReplyTemplatesDelete(context.Background(), 0)
	var ve *ValidationError
	if !errorsAsValidation(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestReplyTemplatesDelete403SurfacesAsAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/reply-templates/42":
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"status":"error","message":"Only the creator may delete this template"}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.ReplyTemplatesDelete(context.Background(), 42)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if _, ok := err.(*AuthError); !ok {
		t.Fatalf("expected *AuthError, got %T (%v)", err, err)
	}
}
