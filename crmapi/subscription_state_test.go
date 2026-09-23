package crmapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubscriptionState(t *testing.T) {
	var gotBody map[string][]int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users/subscription-state":
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			raw, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(raw, &gotBody); err != nil {
				t.Fatalf("bad body %q: %v", raw, err)
			}
			fmt.Fprint(w, `{"status":"success","data":[
				{"user_id":5,"has_active_subscription":true,"frozen":false},
				{"user_id":7,"has_active_subscription":false,"frozen":true},
				{"user_id":9,"has_active_subscription":false,"frozen":false}
			]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.SubscriptionState(context.Background(), []int64{5, 7, 9})
	if err != nil {
		t.Fatalf("SubscriptionState error: %v", err)
	}
	if ids := gotBody["user_ids"]; len(ids) != 3 || ids[0] != 5 || ids[2] != 9 {
		t.Fatalf("request user_ids = %v", ids)
	}
	want := []SubscriptionStateItem{
		{UserID: 5, HasActiveSubscription: true},
		{UserID: 7, Frozen: true},
		{UserID: 9},
	}
	if len(res) != len(want) {
		t.Fatalf("len = %d, want %d", len(res), len(want))
	}
	for i := range want {
		if res[i] != want[i] {
			t.Fatalf("item %d = %+v, want %+v", i, res[i], want[i])
		}
	}
}

func TestSubscriptionState_Validation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("no request expected, got %s", r.URL.Path)
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	tooMany := make([]int64, SubscriptionStateMaxIDs+1)
	for i := range tooMany {
		tooMany[i] = int64(i + 1)
	}
	cases := map[string][]int64{
		"empty":    nil,
		"zero id":  {1, 0},
		"negative": {-3},
		"too many": tooMany,
	}
	for name, ids := range cases {
		_, err := client.SubscriptionState(context.Background(), ids)
		var vErr *ValidationError
		if !errors.As(err, &vErr) {
			t.Fatalf("%s: err = %v, want ValidationError", name, err)
		}
	}
}

func TestUsers_SubscriptionFlags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/staff/123/auth":
			fmt.Fprint(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"2030-01-01T00:00:00Z"}}`)
		case "/api/users":
			fmt.Fprint(w, `{"status":"success","data":{"bot_id":1,"limit":10,"offset":0,"count":2,"items":[
				{"user_id":1,"full_name":"A","restricted":false,"has_active_subscription":true,"frozen":false},
				{"user_id":2,"full_name":"B","restricted":false,"has_active_subscription":false,"frozen":true}
			]}}`)
		case "/api/users/1":
			fmt.Fprint(w, `{"status":"success","data":{"user_id":1,"full_name":"A","status":null,
				"has_active_subscription":true,"frozen":true,"bots_info":[]}}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustNewClient(t, server.URL, server.Client())
	list, err := client.ListUsers(context.Background(), 1, 10, 0)
	if err != nil {
		t.Fatalf("ListUsers error: %v", err)
	}
	if !list.Items[0].HasActiveSubscription || list.Items[0].Frozen {
		t.Fatalf("item 0 = %+v", list.Items[0])
	}
	if list.Items[1].HasActiveSubscription || !list.Items[1].Frozen {
		t.Fatalf("item 1 = %+v", list.Items[1])
	}

	user, err := client.GetUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUser error: %v", err)
	}
	if !user.HasActiveSubscription || !user.Frozen {
		t.Fatalf("GetUser flags = %v/%v, want true/true", user.HasActiveSubscription, user.Frozen)
	}
}
