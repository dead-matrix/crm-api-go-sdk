package crmapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newTaskStatsServer поднимает тестовый CRM: авторизация + handler на path.
func newTaskStatsServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/staff/123/auth" {
			expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
			fmt.Fprintf(w, `{"status":"success","data":{"token":"jwt-1","expires_at":"%s"}}`, expiresAt)
			return
		}
		h, ok := handlers[r.URL.Path]
		if !ok {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusTeapot)
			return
		}
		h(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func TestClient_TaskStats(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/tasks/stats": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":{
				"task_type":"liker","task_id":496,"user_id":8542238123,
				"title":null,"status":0,"status_text":"Активна","paused":false,
				"date_start":"2026-10-09T10:40:12","date_end":null,"tz":"Europe/Moscow",
				"totals":[{"key":"success","label":"Лайков поставлено","value":1993}],
				"accounts":[{"session":"79990001122","phone":"+7999","username":"u",
					"valid":true,"spam_block":false,"in_task":true,"disabled":true,
					"today":12,"total":340,"cut":3,"errors":0,"hourly":null,
					"last_at":"2026-10-09T17:10:02","note":"исключён: AuthKeyUnregistered"}],
				"problems":[{"code":"flood_wait","label":"Анти-флуд","count":14,"accounts":3,"last_at":null,"sample":"строка"}],
				"problem_chats":[{"title":"Канал","url":"https://t.me/x","reason":"нет комментариев","count":2,"last_at":"2026-10-09T11:00:00"}],
				"daily":[{"date":"2026-10-01","value":120,"cut":3}],
				"sources":{"clickhouse":true,"log":false},
				"notes":["n1"]}}`))
		},
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.TaskStats(context.Background(), " liker ", 496, 8542238123, 30)
	if err != nil {
		t.Fatalf("TaskStats error: %v", err)
	}
	if got.Get("task_type") != "liker" || got.Get("task_id") != "496" || got.Get("user_id") != "8542238123" || got.Get("days") != "30" {
		t.Fatalf("query = %v", got)
	}
	if res.TaskID != 496 || res.StatusText != "Активна" || res.Title != nil || res.DateEnd != nil || res.DateStart == nil {
		t.Fatalf("header fields = %+v", res)
	}
	if len(res.Totals) != 1 || res.Totals[0].Value != 1993 {
		t.Fatalf("totals = %+v", res.Totals)
	}
	if len(res.Accounts) != 1 {
		t.Fatalf("accounts = %+v", res.Accounts)
	}
	a := res.Accounts[0]
	if !a.InTask || !a.Disabled || a.Total != 340 || a.Hourly != nil || a.LastAt == nil || a.Note == "" {
		t.Fatalf("account = %+v", a)
	}
	if len(res.Problems) != 1 || res.Problems[0].Accounts != 3 || res.Problems[0].LastAt != nil {
		t.Fatalf("problems = %+v", res.Problems)
	}
	if len(res.ProblemChats) != 1 || res.ProblemChats[0].URL != "https://t.me/x" {
		t.Fatalf("problem_chats = %+v", res.ProblemChats)
	}
	if len(res.Daily) != 1 || res.Daily[0].Cut != 3 {
		t.Fatalf("daily = %+v", res.Daily)
	}
	if !res.Sources.ClickHouse || res.Sources.Log || len(res.Notes) != 1 {
		t.Fatalf("sources/notes = %+v %+v", res.Sources, res.Notes)
	}
}

func TestClient_TaskStats_NoDaysAndValidation(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/tasks/stats": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":{"task_type":"chatting","task_id":1,"user_id":2}}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	if _, err := client.TaskStats(context.Background(), "chatting", 1, 2, 0); err != nil {
		t.Fatalf("TaskStats error: %v", err)
	}
	if got.Has("days") {
		t.Fatalf("days must be omitted when <=0, query = %v", got)
	}

	for name, call := range map[string]func() error{
		"user_id":   func() error { _, err := client.TaskStats(context.Background(), "liker", 1, 0, 0); return err },
		"task_id":   func() error { _, err := client.TaskStats(context.Background(), "liker", 0, 1, 0); return err },
		"task_type": func() error { _, err := client.TaskStats(context.Background(), " ", 1, 1, 0); return err },
	} {
		var ve *ValidationError
		if err := call(); !errors.As(err, &ve) || !strings.Contains(err.Error(), name) {
			t.Fatalf("%s: expected ValidationError, got %v", name, err)
		}
	}
}

func TestClient_TaskStats_NotFound(t *testing.T) {
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/tasks/stats": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":"error","message":"Task not found","code":"not_found"}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	_, err := client.TaskStats(context.Background(), "liker", 1, 2, 0)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected APIError 404, got %v", err)
	}
}

func TestClient_ActivityDaily(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/profile/activity-daily": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":{"days":[
				{"date":"2026-10-08","likes":0,"likes_cut":0,"chatting":0},
				{"date":"2026-10-09","likes":15,"likes_cut":2,"chatting":7}],
				"sources":{"clickhouse":true}}}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ActivityDaily(context.Background(), 100, 14)
	if err != nil {
		t.Fatalf("ActivityDaily error: %v", err)
	}
	if got.Get("user_id") != "100" || got.Get("days") != "14" {
		t.Fatalf("query = %v", got)
	}
	if len(res.Days) != 2 || res.Days[1].Likes != 15 || res.Days[1].LikesCut != 2 || res.Days[1].Chatting != 7 || !res.Sources.ClickHouse {
		t.Fatalf("res = %+v", res)
	}
	if _, err := client.ActivityDaily(context.Background(), 0, 0); err == nil {
		t.Fatalf("expected ValidationError for user_id=0")
	}
}

func TestClient_AccountsList_LikeChattingFields(t *testing.T) {
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/accounts/list": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"status":"success","data":[{"session_name":"s1","valid":true,
				"day_like":5,"all_like":50,"day_chatting":3,"all_chatting":30}]}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	items, err := client.AccountsList(context.Background(), 100, false, false, 0)
	if err != nil {
		t.Fatalf("AccountsList error: %v", err)
	}
	if len(items) != 1 || items[0].DayLike != 5 || items[0].AllLike != 50 || items[0].DayChatting != 3 || items[0].AllChatting != 30 {
		t.Fatalf("items = %+v", items)
	}
}

func TestClient_Schedules(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/schedules": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":[{"id":3039,"title":null,"task_type":"chatting",
				"task_label":"Чаттинг","donor_task_id":2027,"on_pause":false,"runs_done":1,
				"max_runs_limit":null,"period_days":1,"run_time":"12:00","timezone":"Europe/Moscow",
				"first_run_mode":"now","auto_stop":{"enabled":true,"hours":2,"minutes":30},
				"last_run_at":"2026-10-09T14:30:54","next_run_at":"2026-10-10T12:00:00",
				"created_at":"2026-10-01T10:00:00","updated_at":null,
				"last_run":{"task_id":2031,"status":0,"status_text":"Активна","date_start":"2026-10-09T14:30:54","date_end":null}}]}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.Schedules(context.Background(), 100)
	if err != nil {
		t.Fatalf("Schedules error: %v", err)
	}
	if got.Get("user_id") != "100" {
		t.Fatalf("query = %v", got)
	}
	if len(res) != 1 {
		t.Fatalf("res = %+v", res)
	}
	s := res[0]
	if s.ID != 3039 || s.Title != nil || s.DonorTaskID != 2027 || s.MaxRunsLimit != nil || s.RunTime != "12:00" ||
		!s.AutoStop.Enabled || s.AutoStop.Minutes != 30 || s.NextRunAt == nil || s.UpdatedAt != nil {
		t.Fatalf("schedule = %+v", s)
	}
	if s.LastRun == nil || s.LastRun.TaskID != 2031 || s.LastRun.DateEnd != nil {
		t.Fatalf("last_run = %+v", s.LastRun)
	}
}

func TestClient_Schedules_EmptyAndValidation(t *testing.T) {
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/schedules": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"status":"success","data":null}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.Schedules(context.Background(), 100)
	if err != nil || res == nil || len(res) != 0 {
		t.Fatalf("res = %v, err = %v; want empty non-nil slice", res, err)
	}
	if _, err := client.Schedules(context.Background(), 0); err == nil {
		t.Fatalf("expected ValidationError for user_id=0")
	}
}

func TestClient_ScheduleRuns(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/schedules/3039/runs": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":{"schedule_id":3039,"task_type":"chatting","period_days":1,
				"runs":[{"task_id":2031,"status":0,"status_text":"Активна","paused":false,
					"date_start":"2026-10-09T14:30:54","date_end":null,
					"metrics":[{"key":"answers","label":"Сообщений","value":111}],
					"gap_hours":26.5,"late":true},
				{"task_id":2030,"status":1,"status_text":"Завершена","paused":false,
					"date_start":"2026-10-08T12:00:00","date_end":"2026-10-08T20:00:00",
					"metrics":[],"gap_hours":null,"late":false}]}}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.ScheduleRuns(context.Background(), 3039, 100, 50)
	if err != nil {
		t.Fatalf("ScheduleRuns error: %v", err)
	}
	if got.Get("user_id") != "100" || got.Get("limit") != "50" {
		t.Fatalf("query = %v", got)
	}
	if res.ScheduleID != 3039 || res.PeriodDays != 1 || len(res.Runs) != 2 {
		t.Fatalf("res = %+v", res)
	}
	r0 := res.Runs[0]
	if r0.GapHours == nil || *r0.GapHours != 26.5 || !r0.Late || len(r0.Metrics) != 1 || r0.Metrics[0].Value != 111 {
		t.Fatalf("run0 = %+v", r0)
	}
	if res.Runs[1].GapHours != nil || res.Runs[1].DateEnd == nil {
		t.Fatalf("run1 = %+v", res.Runs[1])
	}

	if _, err := client.ScheduleRuns(context.Background(), 0, 100, 0); err == nil || !strings.Contains(err.Error(), "schedule_id") {
		t.Fatalf("expected schedule_id ValidationError, got %v", err)
	}
	if _, err := client.ScheduleRuns(context.Background(), 1, 0, 0); err == nil || !strings.Contains(err.Error(), "user_id") {
		t.Fatalf("expected user_id ValidationError, got %v", err)
	}
}

func TestClient_ScheduleLog(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/schedules/3039/log": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="3039.txt"`)
			_, _ = w.Write([]byte("строка 1\nстрока 2\n"))
		},
		"/api/schedules/404/log": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"status":"error","message":"Log not found","code":"not_found"}`))
		},
	})
	client := mustNewClient(t, server.URL, server.Client())

	filename, content, err := client.ScheduleLog(context.Background(), 3039, 100)
	if err != nil {
		t.Fatalf("ScheduleLog error: %v", err)
	}
	if got.Get("user_id") != "100" {
		t.Fatalf("query = %v", got)
	}
	if filename != "3039.txt" || string(content) != "строка 1\nстрока 2\n" {
		t.Fatalf("filename = %q, content = %q", filename, content)
	}

	_, _, err = client.ScheduleLog(context.Background(), 404, 100)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
		t.Fatalf("expected APIError 404, got %v", err)
	}

	if _, _, err := client.ScheduleLog(context.Background(), 0, 100); err == nil {
		t.Fatalf("expected ValidationError for schedule_id=0")
	}
}
