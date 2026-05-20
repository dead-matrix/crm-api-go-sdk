package crmapi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestReplyTemplateListItem_MarshalUsesCamelCase гарантирует, что
// re-marshal публичной структуры даёт camelCase ключи (паритет с
// CRM-сервером), а не Go-имена полей вроде "LastUsedAt".
func TestReplyTemplateListItem_MarshalUsesCamelCase(t *testing.T) {
	ts := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	item := ReplyTemplateListItem{
		ID:         1,
		PublicID:   "00000000-0000-0000-0000-000000000001",
		Title:      "Шаблон",
		Kind:       "single",
		UsageCount: 3,
		LastUsedAt: &ts,
		CreatedAt:  &ts,
		UpdatedAt:  &ts,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	wire := string(data)
	for _, key := range []string{"lastUsedAt", "createdAt", "updatedAt", "publicId", "usageCount"} {
		if !strings.Contains(wire, `"`+key+`"`) {
			t.Fatalf("expected key %q in wire %s", key, wire)
		}
	}
	for _, badKey := range []string{"LastUsedAt", "CreatedAt", "UpdatedAt", "PublicID", "UsageCount"} {
		if strings.Contains(wire, `"`+badKey+`"`) {
			t.Fatalf("unexpected Go-name key %q in wire %s", badKey, wire)
		}
	}
}

// TestReplyTemplateListItem_NilTimestampsOmitted: nil-указатели на даты
// не должны порождать "lastUsedAt":null. omitempty гарантирует чистый
// wire-format при re-marshal.
func TestReplyTemplateListItem_NilTimestampsOmitted(t *testing.T) {
	item := ReplyTemplateListItem{
		ID:       1,
		PublicID: "00000000-0000-0000-0000-000000000001",
		Title:    "Шаблон",
		Kind:     "single",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	wire := string(data)
	for _, key := range []string{"lastUsedAt", "createdAt", "updatedAt"} {
		if strings.Contains(wire, `"`+key+`"`) {
			t.Fatalf("expected key %q to be omitted, got %s", key, wire)
		}
	}
}

// TestReplyTemplateFull_MarshalUsesCamelCase — то же для ReplyTemplateFull.
func TestReplyTemplateFull_MarshalUsesCamelCase(t *testing.T) {
	ts := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	full := ReplyTemplateFull{
		ID:        2,
		PublicID:  "00000000-0000-0000-0000-000000000002",
		Title:     "Альбом",
		Kind:      "album",
		Items:     []ReplyTemplateItem{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}

	data, err := json.Marshal(full)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	wire := string(data)
	for _, key := range []string{"createdAt", "updatedAt", "publicId"} {
		if !strings.Contains(wire, `"`+key+`"`) {
			t.Fatalf("expected key %q in wire %s", key, wire)
		}
	}
	for _, badKey := range []string{"CreatedAt", "UpdatedAt", "PublicID"} {
		if strings.Contains(wire, `"`+badKey+`"`) {
			t.Fatalf("unexpected Go-name key %q in wire %s", badKey, wire)
		}
	}
}
