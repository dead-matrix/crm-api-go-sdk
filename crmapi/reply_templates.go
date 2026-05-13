package crmapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// itemTypeSet lists every valid value for ReplyTemplateItem.Type.
// album items additionally must come from albumItemTypeSet.
var itemTypeSet = map[string]struct{}{
	ReplyTemplateItemTypeText:      {},
	ReplyTemplateItemTypePhoto:     {},
	ReplyTemplateItemTypeVideo:     {},
	ReplyTemplateItemTypeGIF:       {},
	ReplyTemplateItemTypeVoice:     {},
	ReplyTemplateItemTypeVideoNote: {},
	ReplyTemplateItemTypeSticker:   {},
	ReplyTemplateItemTypeFile:      {},
}

var albumItemTypeSet = map[string]struct{}{
	ReplyTemplateItemTypePhoto: {},
	ReplyTemplateItemTypeVideo: {},
	ReplyTemplateItemTypeGIF:   {},
}

var noCaptionTypeSet = map[string]struct{}{
	ReplyTemplateItemTypeVoice:     {},
	ReplyTemplateItemTypeVideoNote: {},
	ReplyTemplateItemTypeSticker:   {},
}

// rawReplyTemplateItem is the over-the-wire shape that mirrors the
// public ReplyTemplateItem. Defined separately so the test fixtures
// can construct JSON without depending on the public struct having
// pointer semantics for every field (which is the current shape).
type rawReplyTemplateItem struct {
	Position        int     `json:"position"`
	Type            string  `json:"type"`
	Caption         *string `json:"caption"`
	MediaObjectKey  *string `json:"mediaObjectKey"`
	Mime            *string `json:"mime"`
	SizeBytes       *int64  `json:"sizeBytes"`
	DurationMs      *int64  `json:"durationMs"`
	Width           *int    `json:"width"`
	Height          *int    `json:"height"`
	FileName        *string `json:"fileName"`
	OriginTenantID  *string `json:"originTenantId"`
	OriginMessageID *string `json:"originMessageId"`
}

func (r rawReplyTemplateItem) toPublic() ReplyTemplateItem {
	return ReplyTemplateItem{
		Position:        r.Position,
		Type:            r.Type,
		Caption:         r.Caption,
		MediaObjectKey:  r.MediaObjectKey,
		Mime:            r.Mime,
		SizeBytes:       r.SizeBytes,
		DurationMs:      r.DurationMs,
		Width:           r.Width,
		Height:          r.Height,
		FileName:        r.FileName,
		OriginTenantID:  r.OriginTenantID,
		OriginMessageID: r.OriginMessageID,
	}
}

// ReplyTemplatesList returns the multimedia quick-reply templates
// visible to the authenticated staff member.
//
// Server-side sort is `usage_count DESC, id ASC` over reply_template_usages
// scoped to the current staff — i.e. the ranking is personal.
//
// limit/offset paginate. limit is clamped server-side; pass 0 to use
// the server default (100 at the time of writing).
func (c *Client) ReplyTemplatesList(ctx context.Context, limit int64, offset int64) ([]ReplyTemplateListItem, error) {
	if limit < 0 {
		return nil, &ValidationError{Message: "limit must be non-negative"}
	}
	if offset < 0 {
		return nil, &ValidationError{Message: "offset must be non-negative"}
	}

	query := map[string]string{}
	if limit > 0 {
		query["limit"] = fmt.Sprintf("%d", limit)
	}
	if offset > 0 {
		query["offset"] = fmt.Sprintf("%d", offset)
	}

	var raw []struct {
		ID         int64                `json:"id"`
		PublicID   string               `json:"publicId"`
		Title      string               `json:"title"`
		Kind       string               `json:"kind"`
		Creator    ReplyTemplateCreator `json:"creator"`
		Preview    ReplyTemplatePreview `json:"preview"`
		UsageCount int64                `json:"usageCount"`
		LastUsedAt *string              `json:"lastUsedAt"`
		CreatedAt  *string              `json:"createdAt"`
		UpdatedAt  *string              `json:"updatedAt"`
	}

	if err := c.get(ctx, "/api/reply-templates", query, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ReplyTemplateListItem, 0, len(raw))
	for _, row := range raw {
		items = append(items, ReplyTemplateListItem{
			ID:         row.ID,
			PublicID:   row.PublicID,
			Title:      row.Title,
			Kind:       row.Kind,
			Creator:    row.Creator,
			Preview:    row.Preview,
			UsageCount: row.UsageCount,
			LastUsedAt: parseOptionalTime(row.LastUsedAt),
			CreatedAt:  parseOptionalTime(row.CreatedAt),
			UpdatedAt:  parseOptionalTime(row.UpdatedAt),
		})
	}
	return items, nil
}

// ReplyTemplatesGet returns the full template payload (including items
// sorted by Position) and inkrements the personal usage counter on the
// CRM side for the authenticated staff. The increment is intentional —
// it powers the personal ranking in ReplyTemplatesList.
func (c *Client) ReplyTemplatesGet(ctx context.Context, templateID int64) (*ReplyTemplateFull, error) {
	if templateID <= 0 {
		return nil, &ValidationError{Message: "template_id must be a positive integer"}
	}

	var raw struct {
		ID        int64                  `json:"id"`
		PublicID  string                 `json:"publicId"`
		Title     string                 `json:"title"`
		Kind      string                 `json:"kind"`
		Creator   ReplyTemplateCreator   `json:"creator"`
		Items     []rawReplyTemplateItem `json:"items"`
		CreatedAt *string                `json:"createdAt"`
		UpdatedAt *string                `json:"updatedAt"`
	}

	if err := c.get(ctx, fmt.Sprintf("/api/reply-templates/%d", templateID), nil, true, &raw); err != nil {
		return nil, err
	}

	items := make([]ReplyTemplateItem, 0, len(raw.Items))
	for _, it := range raw.Items {
		items = append(items, it.toPublic())
	}

	return &ReplyTemplateFull{
		ID:        raw.ID,
		PublicID:  raw.PublicID,
		Title:     raw.Title,
		Kind:      raw.Kind,
		Creator:   raw.Creator,
		Items:     items,
		CreatedAt: parseOptionalTime(raw.CreatedAt),
		UpdatedAt: parseOptionalTime(raw.UpdatedAt),
	}, nil
}

// ReplyTemplatesCreate persists a new template under the current staff
// member as its creator. Server-side validation mirrors validateCreate;
// pre-validating here gives the caller a fast, well-typed error path.
func (c *Client) ReplyTemplatesCreate(ctx context.Context, input CreateReplyTemplateInput) (*ReplyTemplateFull, error) {
	input = input.normalized()
	if err := input.validate(); err != nil {
		return nil, err
	}

	var raw struct {
		ID        int64                  `json:"id"`
		PublicID  string                 `json:"publicId"`
		Title     string                 `json:"title"`
		Kind      string                 `json:"kind"`
		Creator   ReplyTemplateCreator   `json:"creator"`
		Items     []rawReplyTemplateItem `json:"items"`
		CreatedAt *string                `json:"createdAt"`
		UpdatedAt *string                `json:"updatedAt"`
	}

	if err := c.post(ctx, "/api/reply-templates", nil, true, input, &raw); err != nil {
		return nil, err
	}

	items := make([]ReplyTemplateItem, 0, len(raw.Items))
	for _, it := range raw.Items {
		items = append(items, it.toPublic())
	}

	return &ReplyTemplateFull{
		ID:        raw.ID,
		PublicID:  raw.PublicID,
		Title:     raw.Title,
		Kind:      raw.Kind,
		Creator:   raw.Creator,
		Items:     items,
		CreatedAt: parseOptionalTime(raw.CreatedAt),
		UpdatedAt: parseOptionalTime(raw.UpdatedAt),
	}, nil
}

// ReplyTemplatesDelete removes a template owned by the current staff
// member. CRM enforces creator-only deletion — a non-creator request
// surfaces as AuthError (HTTP 403). Cascading rows in items/usages
// are cleaned up by the database.
func (c *Client) ReplyTemplatesDelete(ctx context.Context, templateID int64) (*DeleteReplyTemplateResult, error) {
	if templateID <= 0 {
		return nil, &ValidationError{Message: "template_id must be a positive integer"}
	}

	var raw struct {
		ID       int64  `json:"id"`
		PublicID string `json:"publicId"`
	}

	if err := c.delete(ctx, fmt.Sprintf("/api/reply-templates/%d", templateID), nil, true, &raw); err != nil {
		return nil, err
	}

	return &DeleteReplyTemplateResult{
		ID:       raw.ID,
		PublicID: raw.PublicID,
	}, nil
}

// ───────── validation helpers ─────────

func (in CreateReplyTemplateInput) normalized() CreateReplyTemplateInput {
	in.Title = strings.TrimSpace(in.Title)
	in.Kind = strings.TrimSpace(in.Kind)
	// Item caption is normalised on per-item basis below; pointer
	// nilness is preserved (the API distinguishes nil caption from
	// empty string for some flows on the future PATCH endpoint).
	for i := range in.Items {
		it := &in.Items[i]
		it.Type = strings.TrimSpace(it.Type)
		if it.Caption != nil {
			trimmed := strings.TrimSpace(*it.Caption)
			it.Caption = &trimmed
		}
		if it.MediaObjectKey != nil {
			trimmed := strings.TrimSpace(*it.MediaObjectKey)
			it.MediaObjectKey = &trimmed
		}
		if it.FileName != nil {
			trimmed := strings.TrimSpace(*it.FileName)
			it.FileName = &trimmed
		}
	}
	return in
}

func (in CreateReplyTemplateInput) validate() error {
	if in.Title == "" {
		return &ValidationError{Message: "title must not be empty"}
	}
	if len(in.Title) > ReplyTemplateTitleMaxLength {
		return &ValidationError{Message: fmt.Sprintf("title must be at most %d characters", ReplyTemplateTitleMaxLength)}
	}
	switch in.Kind {
	case ReplyTemplateKindSingle, ReplyTemplateKindAlbum:
	default:
		return &ValidationError{Message: fmt.Sprintf("kind must be %q or %q", ReplyTemplateKindSingle, ReplyTemplateKindAlbum)}
	}
	if len(in.Items) == 0 {
		return &ValidationError{Message: "items must contain at least one element"}
	}
	if len(in.Items) > ReplyTemplateAlbumMaxItems {
		return &ValidationError{Message: fmt.Sprintf("items must contain at most %d elements", ReplyTemplateAlbumMaxItems)}
	}

	// Validate per-item shape.
	for _, it := range in.Items {
		if _, ok := itemTypeSet[it.Type]; !ok {
			return &ValidationError{Message: fmt.Sprintf("item type %q is not supported", it.Type)}
		}
		if it.Position < 0 || it.Position > ReplyTemplateItemMaxPos {
			return &ValidationError{Message: fmt.Sprintf("item position must be in 0..%d (got %d)", ReplyTemplateItemMaxPos, it.Position)}
		}
		if it.Caption != nil && len(*it.Caption) > ReplyTemplateCaptionMax {
			return &ValidationError{Message: fmt.Sprintf("item caption must be at most %d characters", ReplyTemplateCaptionMax)}
		}
		if it.Type == ReplyTemplateItemTypeText {
			if it.Caption == nil || *it.Caption == "" {
				return &ValidationError{Message: "text item requires non-empty caption"}
			}
			if it.MediaObjectKey != nil {
				return &ValidationError{Message: "text item must not carry media_object_key"}
			}
			if it.Mime != nil || it.SizeBytes != nil || it.DurationMs != nil ||
				it.Width != nil || it.Height != nil || it.FileName != nil {
				return &ValidationError{Message: "text item must not carry media metadata"}
			}
		} else {
			if it.MediaObjectKey == nil || *it.MediaObjectKey == "" {
				return &ValidationError{Message: fmt.Sprintf("%s item requires media_object_key", it.Type)}
			}
			if _, isNoCaption := noCaptionTypeSet[it.Type]; isNoCaption {
				if it.Caption != nil && *it.Caption != "" {
					return &ValidationError{Message: fmt.Sprintf("%s item must not carry caption", it.Type)}
				}
			}
		}
	}

	// Validate positions: contiguous 0..N-1, no duplicates.
	seen := make(map[int]struct{}, len(in.Items))
	for _, it := range in.Items {
		if _, dup := seen[it.Position]; dup {
			return &ValidationError{Message: fmt.Sprintf("duplicate position %d in items", it.Position)}
		}
		seen[it.Position] = struct{}{}
	}
	for i := 0; i < len(in.Items); i++ {
		if _, ok := seen[i]; !ok {
			return &ValidationError{Message: fmt.Sprintf("positions must be contiguous 0..%d", len(in.Items)-1)}
		}
	}

	// Kind-specific rules.
	switch in.Kind {
	case ReplyTemplateKindSingle:
		if len(in.Items) != 1 {
			return &ValidationError{Message: "single template requires exactly 1 item"}
		}
	case ReplyTemplateKindAlbum:
		if len(in.Items) < ReplyTemplateAlbumMinItems {
			return &ValidationError{Message: fmt.Sprintf("album requires at least %d items", ReplyTemplateAlbumMinItems)}
		}
		for _, it := range in.Items {
			if _, ok := albumItemTypeSet[it.Type]; !ok {
				return &ValidationError{Message: fmt.Sprintf("album item type must be one of photo/video/gif (got %q at position %d)", it.Type, it.Position)}
			}
			if it.Position != 0 && it.Caption != nil && *it.Caption != "" {
				return &ValidationError{Message: fmt.Sprintf("caption only allowed at position=0 in album (found at position %d)", it.Position)}
			}
		}
	}

	return nil
}

func parseOptionalTime(s *string) *time.Time {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return utils.ParseTime(trimmed)
}
