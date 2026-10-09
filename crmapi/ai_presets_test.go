package crmapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func TestClient_AIPresets(t *testing.T) {
	var got url.Values
	server := newTaskStatsServer(t, map[string]http.HandlerFunc{
		"/api/ai-presets": func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query()
			_, _ = w.Write([]byte(`{"status":"success","data":{"days":30,"functions":[
				{"function":"comment","key_mode":"service",
				 "settings":{"model":"openai/gpt-5-mini","system_prompt":"S","forbidden_topics":null,
				   "reasoning":true,"max_chars":300,"updated_at":"2026-10-01T12:00:00"},
				 "presets":[{"id":12,"name":"Мягкий","model":null,"system_prompt":"P","forbidden_topics":"",
				   "reasoning":false,"max_chars":0,"key_mode":"byok","has_byok":true,
				   "created_at":"2026-09-30T10:00:00","tasks_total":3,"tasks_active":1}],
				 "tasks":[{"task_id":501,"title":null,"status":1,"status_text":"Отслеживание",
				   "active":true,"paused":false,"date_start":"2026-10-09T10:00:00","date_end":null,
				   "preset_id":12,"preset_name":"Мягкий","preset_missing":false,
				   "effective":{"model":"openai/gpt-5-mini","model_source":"settings","prompt_source":"preset",
				     "forbidden_topics":"","forbidden_source":"preset","reasoning":false,"reasoning_source":"preset",
				     "max_chars":300,"max_chars_source":"settings","key_mode":"byok","key_source":"preset"}}],
				 "tasks_total":14,
				 "used_models":[{"model":"openai/gpt-5-mini","count":120,"last_at":"2026-10-09T17:00:00"}]},
				{"function":"chatting","key_mode":"byok","settings":null,"presets":[],"tasks":[],
				 "tasks_total":0,"used_models":[]}]}}`))
		},
	})

	client := mustNewClient(t, server.URL, server.Client())
	res, err := client.AIPresets(context.Background(), 8099510570)
	if err != nil {
		t.Fatalf("AIPresets error: %v", err)
	}
	if got.Get("user_id") != "8099510570" {
		t.Fatalf("query = %v", got)
	}
	if res.Days != 30 || len(res.Functions) != 2 {
		t.Fatalf("overview = %+v", res)
	}
	c := res.Functions[0]
	if c.Function != "comment" || c.Settings == nil || c.Settings.MaxChars != 300 || c.TasksTotal != 14 {
		t.Fatalf("comment block = %+v", c)
	}
	if len(c.Presets) != 1 || !c.Presets[0].HasBYOK || c.Presets[0].Model != nil || c.Presets[0].TasksActive != 1 {
		t.Fatalf("presets = %+v", c.Presets)
	}
	if c.Presets[0].ForbiddenTopics == nil || *c.Presets[0].ForbiddenTopics != "" {
		t.Fatalf("empty forbidden topics must stay a non-nil empty string")
	}
	task := c.Tasks[0]
	if task.PresetID == nil || *task.PresetID != 12 || task.Title != nil || task.Effective.KeySource != "preset" ||
		task.Effective.MaxChars != 300 || task.Effective.PromptSource != "preset" {
		t.Fatalf("task = %+v", task)
	}
	if len(c.UsedModels) != 1 || c.UsedModels[0].Count != 120 {
		t.Fatalf("used models = %+v", c.UsedModels)
	}
	ch := res.Functions[1]
	if ch.Settings != nil || ch.KeyMode != "byok" || ch.Presets == nil {
		t.Fatalf("chatting block = %+v", ch)
	}
}

func TestClient_AIPresets_Validation(t *testing.T) {
	client := mustNewClient(t, "http://127.0.0.1:1", nil)
	_, err := client.AIPresets(context.Background(), 0)
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
}
