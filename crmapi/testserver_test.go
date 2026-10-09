package crmapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newCRMTestServer поднимает тестовый сервер CRM: выдачу JWT он обслуживает
// сам, остальные запросы передаёт handler. Вынесен, чтобы тесты методов
// проверяли свой маршрут и не повторяли обвязку авторизации.
func newCRMTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/staff/123/auth" {
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
			return
		}
		handler(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

// writeSuccess отвечает успешным конвертом CRM с готовым JSON в data.
func writeSuccess(w http.ResponseWriter, data string) {
	fmt.Fprintf(w, `{"status":"success","data":%s}`, data)
}

// writeAPIError отвечает конвертом ошибки CRM с HTTP-статусом и кодом.
func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"status":"error","code":%q,"message":%q,"data":null}`, code, message)
}

// readJSONBody разбирает тело запроса в map: по ней видно и значения, и то,
// какие ключи клиент не отправил вовсе (для omitempty это важнее значений).
func readJSONBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("bad JSON body %q: %v", raw, err)
	}
	return body
}

// wantRequest падает, если запрос пришёл не тем методом или не на тот путь.
func wantRequest(t *testing.T, r *http.Request, method, path string) {
	t.Helper()
	if r.Method != method || r.URL.Path != path {
		t.Fatalf("request = %s %s, want %s %s", r.Method, r.URL.Path, method, path)
	}
}
