package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func changelogServer(t *testing.T, body string, gotQuery *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
		case "/api/servers/changelog":
			if gotQuery != nil {
				*gotQuery = r.URL.RawQuery
			}
			_, _ = w.Write([]byte(body))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
}

// Клиент отстал: приходят «не загруженные обновления» по возрастанию версии.
func TestClient_ServersChangelog_Pending(t *testing.T) {
	var gotQuery string
	server := changelogServer(t, `{"status":"success","data":{
		"latest":"1.8.4","known":true,"up_to_date":false,
		"versions":[
			{"version":"1.8.3","released_at":"2026-07-01T10:00:00","items":[
				{"text":"Исправлено это","kind":"fixed"},
				{"text":"Добавлено то","kind":"added"}
			]},
			{"version":"1.8.4","released_at":null,"items":[
				{"text":"Изменён текст","kind":"changed"}
			]}
		]
	}}`, &gotQuery)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ServersChangelog(context.Background(), "1.8.2")
	if err != nil {
		t.Fatalf("ServersChangelog error: %v", err)
	}
	if gotQuery != "version=1.8.2" {
		t.Fatalf("query = %q, want version=1.8.2", gotQuery)
	}
	if res.Latest == nil || *res.Latest != "1.8.4" {
		t.Fatalf("latest = %v, want 1.8.4", res.Latest)
	}
	if !res.Known || res.UpToDate {
		t.Fatalf("known=%v upToDate=%v, want known=true upToDate=false", res.Known, res.UpToDate)
	}
	if len(res.Versions) != 2 {
		t.Fatalf("versions = %d, want 2", len(res.Versions))
	}
	if res.Versions[0].Version != "1.8.3" || len(res.Versions[0].Items) != 2 {
		t.Fatalf("unexpected first version: %+v", res.Versions[0])
	}
	if res.Versions[0].Items[0].Text != "Исправлено это" || res.Versions[0].Items[0].Kind != "fixed" {
		t.Fatalf("unexpected first item: %+v", res.Versions[0].Items[0])
	}
	if res.Versions[0].ReleasedAt == nil || *res.Versions[0].ReleasedAt != "2026-07-01T10:00:00" {
		t.Fatalf("released_at = %v", res.Versions[0].ReleasedAt)
	}
	if res.Versions[1].ReleasedAt != nil {
		t.Fatalf("released_at must stay nil when absent, got %v", res.Versions[1].ReleasedAt)
	}
}

// Клиент на актуальной версии: пустой список + up_to_date.
func TestClient_ServersChangelog_UpToDate(t *testing.T) {
	server := changelogServer(t, `{"status":"success","data":{
		"latest":"1.8.4","known":true,"up_to_date":true,"versions":[]
	}}`, nil)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ServersChangelog(context.Background(), "1.8.4")
	if err != nil {
		t.Fatalf("ServersChangelog error: %v", err)
	}
	if !res.UpToDate {
		t.Fatalf("upToDate = false, want true")
	}
	// Пустой ответ обязан дать пустой срез, а не nil (иначе JSON наверх уедет null).
	if res.Versions == nil {
		t.Fatalf("versions must be an empty slice, got nil")
	}
	if len(res.Versions) != 0 {
		t.Fatalf("versions = %d, want 0", len(res.Versions))
	}
}

// Версия воркера неизвестна (сервер выключен): version в query не уходит,
// known=false — «обновлений нет» утверждать нельзя.
func TestClient_ServersChangelog_UnknownVersionOmitsQuery(t *testing.T) {
	var gotQuery string
	server := changelogServer(t, `{"status":"success","data":{
		"latest":"1.8.4","known":false,"up_to_date":false,"versions":[]
	}}`, &gotQuery)
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ServersChangelog(context.Background(), "   ")
	if err != nil {
		t.Fatalf("ServersChangelog error: %v", err)
	}
	if gotQuery != "" {
		t.Fatalf("query = %q, want empty (no version param)", gotQuery)
	}
	if res.Known || res.UpToDate {
		t.Fatalf("known=%v upToDate=%v, want both false", res.Known, res.UpToDate)
	}
}
