package crmapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetUser_FreezeFields: GetUser декодирует через приватный raw-struct и
// маппит в публичный UserBotInfo вручную. Регресс: поля заморозки (frozen/
// frozen_at/frozen_expiry) ДОЛЖНЫ присутствовать в raw-struct и переноситься —
// иначе они молча теряются (так и было: панель показывала «не заморожено» при
// реально замороженной подписке). frozen_at приходит наивным ISO без зоны —
// проверяем, что ParseTime его принимает (а не роняет весь анмаршал).
func TestGetUser_FreezeFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users/7020095937":
			fmt.Fprint(w, `{"status":"success","data":{
				"user_id":7020095937,
				"full_name":"Admin",
				"bots_info":[
					{"bot_id":1,"bot_name":"Main","access":null,"access_end":null,
					 "frozen":true,"frozen_at":"2026-06-26T14:52:13",
					 "frozen_expiry":{"classic_comment":"2026-07-02T22:55:05.159906","viewer":"2026-07-05T13:13:08.717352"}},
					{"bot_id":2,"bot_name":"Support","access":null,"access_end":null,"frozen":false}
				]
			}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.GetUser(context.Background(), 7020095937)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if len(res.BotsInfo) != 2 {
		t.Fatalf("BotsInfo len = %d, want 2", len(res.BotsInfo))
	}

	main := res.BotsInfo[0]
	if !main.Frozen {
		t.Fatalf("bot 1 Frozen = false, want true")
	}
	if main.FrozenAt == nil {
		t.Fatalf("bot 1 FrozenAt = nil, want parsed time (naive ISO must parse via ParseTime)")
	}
	if got := len(main.FrozenExpiry); got != 2 {
		t.Fatalf("bot 1 FrozenExpiry len = %d, want 2", got)
	}
	if main.FrozenExpiry["classic_comment"] == "" {
		t.Fatalf("bot 1 FrozenExpiry[classic_comment] empty, want end-date string")
	}

	support := res.BotsInfo[1]
	if support.Frozen {
		t.Fatalf("bot 2 Frozen = true, want false")
	}
	if support.FrozenAt != nil {
		t.Fatalf("bot 2 FrozenAt != nil, want nil (not frozen)")
	}
}
