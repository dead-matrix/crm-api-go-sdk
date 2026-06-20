package crmapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dead-matrix/crm-api-go-sdk/crmapi/internal/utils"
)

// publicIDPattern accepts canonical 36-char UUID (8-4-4-4-12 hex,
// case-insensitive). Matches what CRM-API regex enforces — pre-validating
// here gives a fast, type-safe error path before the HTTP round-trip.
var publicIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
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
	// ID is sent by the server in read responses (GET full template);
	// absent on create requests (server assigns). Zero is "unknown",
	// which `toPublic()` propagates to the public struct as the
	// `omitempty` int64 zero value.
	ID              int64   `json:"id"`
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
		ID:              r.ID,
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
		Command    *string              `json:"command"`
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
			Command:    row.Command,
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
		Command   *string                `json:"command"`
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
		Command:   raw.Command,
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
		Command   *string                `json:"command"`
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
		Command:   raw.Command,
		Creator:   raw.Creator,
		Items:     items,
		CreatedAt: parseOptionalTime(raw.CreatedAt),
		UpdatedAt: parseOptionalTime(raw.UpdatedAt),
	}, nil
}

// replyTemplateCommandPattern mirrors the CRM regex: letters (any script,
// so Cyrillic too), digits, underscore. Applied to the already-cleaned
// token (leading "/" stripped, trimmed, lower-cased).
var replyTemplateCommandPattern = regexp.MustCompile(`^[\p{L}\p{N}_]+$`)

// normalizeReplyTemplateCommand cleans + validates a command for SetCommand.
// nil / empty / "/" → nil (meaning "clear the command"). Otherwise strips a
// leading "/", trims, lower-cases, and validates length + charset — mirroring
// the CRM's _normalize_command_input so a malformed value fails fast before
// the round-trip. The keyboard-layout canonicalisation used for uniqueness
// lives server-side (and in the messenger frontend for matching); the SDK
// only validates shape.
func normalizeReplyTemplateCommand(command *string) (*string, error) {
	if command == nil {
		return nil, nil
	}
	s := strings.TrimSpace(*command)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	s = strings.ToLower(s)
	if utf8.RuneCountInString(s) > ReplyTemplateCommandMaxLength {
		return nil, &ValidationError{Message: fmt.Sprintf("command must be at most %d characters", ReplyTemplateCommandMaxLength)}
	}
	if !replyTemplateCommandPattern.MatchString(s) {
		return nil, &ValidationError{Message: "command may contain only letters, digits and underscore"}
	}
	return &s, nil
}

// ReplyTemplatesSetCommand sets or clears the slash-command trigger of a
// template. command=nil (or a pointer to an empty / slash-only string)
// CLEARS the command; a non-empty value sets it. The value is normalised
// here (trim, strip leading "/", lower-case) and shape-validated before the
// PATCH. The server additionally enforces creator-only authorisation
// (HTTP 403 → AuthError) and global uniqueness by the layout-canonical form
// (HTTP 409 → APIError). Returns the updated full template.
func (c *Client) ReplyTemplatesSetCommand(ctx context.Context, templateID int64, command *string) (*ReplyTemplateFull, error) {
	if templateID <= 0 {
		return nil, &ValidationError{Message: "template_id must be a positive integer"}
	}

	normalized, err := normalizeReplyTemplateCommand(command)
	if err != nil {
		return nil, err
	}

	body := struct {
		Command *string `json:"command"`
	}{Command: normalized}

	var raw struct {
		ID        int64                  `json:"id"`
		PublicID  string                 `json:"publicId"`
		Title     string                 `json:"title"`
		Kind      string                 `json:"kind"`
		Command   *string                `json:"command"`
		Creator   ReplyTemplateCreator   `json:"creator"`
		Items     []rawReplyTemplateItem `json:"items"`
		CreatedAt *string                `json:"createdAt"`
		UpdatedAt *string                `json:"updatedAt"`
	}

	if err := c.patch(ctx, fmt.Sprintf("/api/reply-templates/%d", templateID), nil, true, body, &raw); err != nil {
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
		Command:   raw.Command,
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
	in.PublicID = strings.ToLower(strings.TrimSpace(in.PublicID))
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
	if in.PublicID != "" && !publicIDPattern.MatchString(in.PublicID) {
		return &ValidationError{Message: "public_id must be a 36-char UUID (8-4-4-4-12 hex)"}
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

// ───────── Delivery refs (file-id reuse cache) ─────────
//
// Two endpoints, both scoped to one template_id:
//
//   GET /api/reply-templates/{template_id}/delivery-refs
//       ?provider=...&providerScope=...
//   PUT /api/reply-templates/{template_id}/delivery-refs
//
// The messenger reads at PrepareSend time, falls back to S3 URLs when
// a ref is missing, and PUTs back the freshly-extracted Telegram
// file_ids after a successful dispatch.

// rawDeliveryRef хранит ISO timestamps как сырые строки до парсинга в
// public DeliveryRef. Нужен потому что CRM возвращает даты в ISO 8601, а
// public-API SDK выдаёт *time.Time (паритет с Python Optional[datetime]).
type rawDeliveryRef struct {
	ID             int64   `json:"id"`
	ItemID         int64   `json:"itemId"`
	Provider       string  `json:"provider"`
	ProviderScope  string  `json:"providerScope"`
	MediaRef       string  `json:"mediaRef"`
	MediaUniqueRef *string `json:"mediaUniqueRef"`
	MediaType      *string `json:"mediaType"`
	FailCount      int     `json:"failCount"`
	LastUsedAt     *string `json:"lastUsedAt"`
	CreatedAt      *string `json:"createdAt"`
	UpdatedAt      *string `json:"updatedAt"`
}

func (r rawDeliveryRef) toPublic() DeliveryRef {
	return DeliveryRef{
		ID:             r.ID,
		ItemID:         r.ItemID,
		Provider:       r.Provider,
		ProviderScope:  r.ProviderScope,
		MediaRef:       r.MediaRef,
		MediaUniqueRef: r.MediaUniqueRef,
		MediaType:      r.MediaType,
		FailCount:      r.FailCount,
		LastUsedAt:     parseOptionalTime(r.LastUsedAt),
		CreatedAt:      parseOptionalTime(r.CreatedAt),
		UpdatedAt:      parseOptionalTime(r.UpdatedAt),
	}
}

// listDeliveryRefsRaw mirrors the server response envelope.
type listDeliveryRefsRaw struct {
	Refs []rawDeliveryRef `json:"refs"`
}

// upsertDeliveryRefsRaw mirrors the server response envelope.
type upsertDeliveryRefsRaw struct {
	Refs []rawDeliveryRef `json:"refs"`
}

// ReplyTemplatesDeliveryRefsList returns every delivery ref persisted
// for the given template under (provider, providerScope). An empty
// slice (no error) is the normal "first send, nothing cached" state -
// the messenger uses an empty result to mean "fall back to S3 URL for
// every item".
//
// Validation:
//   - templateID > 0
//   - provider non-empty
//   - providerScope non-empty
//
// HTTP mapping:
//   - 404 → APIError (template not found)
//   - 422 → ValidationError (unknown provider on server side)
func (c *Client) ReplyTemplatesDeliveryRefsList(
	ctx context.Context,
	templateID int64,
	provider string,
	providerScope string,
) ([]DeliveryRef, error) {
	if templateID <= 0 {
		return nil, &ValidationError{Message: "template_id must be a positive integer"}
	}
	provider = strings.TrimSpace(provider)
	providerScope = strings.TrimSpace(providerScope)
	if provider == "" {
		return nil, &ValidationError{Message: "provider must not be empty"}
	}
	if providerScope == "" {
		return nil, &ValidationError{Message: "providerScope must not be empty"}
	}

	query := map[string]string{
		"provider":      provider,
		"providerScope": providerScope,
	}
	var raw listDeliveryRefsRaw
	path := fmt.Sprintf("/api/reply-templates/%d/delivery-refs", templateID)
	if err := c.get(ctx, path, query, true, &raw); err != nil {
		return nil, err
	}
	refs := make([]DeliveryRef, 0, len(raw.Refs))
	for _, r := range raw.Refs {
		refs = append(refs, r.toPublic())
	}
	return refs, nil
}

// ReplyTemplatesDeliveryRefsUpsert idempotently writes (or refreshes)
// a batch of delivery refs for one template. Every entry in `input.Refs`
// must address an item that belongs to `templateID`; otherwise the
// server returns 422.
//
// Validation (client-side, pre-flight):
//   - templateID > 0
//   - provider non-empty (server further restricts to known providers)
//   - providerScope non-empty
//   - refs len ∈ [1, 10]
//   - each ItemID > 0
//   - each MediaRef non-empty
//
// Returns the post-upsert rows in their final shape so the caller can
// log/diagnose without a second round-trip.
func (c *Client) ReplyTemplatesDeliveryRefsUpsert(
	ctx context.Context,
	templateID int64,
	input UpsertDeliveryRefsInput,
) ([]DeliveryRef, error) {
	if templateID <= 0 {
		return nil, &ValidationError{Message: "template_id must be a positive integer"}
	}
	in := input.normalized()
	if err := in.validate(); err != nil {
		return nil, err
	}

	var raw upsertDeliveryRefsRaw
	path := fmt.Sprintf("/api/reply-templates/%d/delivery-refs", templateID)
	if err := c.put(ctx, path, nil, true, in, &raw); err != nil {
		return nil, err
	}
	refs := make([]DeliveryRef, 0, len(raw.Refs))
	for _, r := range raw.Refs {
		refs = append(refs, r.toPublic())
	}
	return refs, nil
}

func (in UpsertDeliveryRefsInput) normalized() UpsertDeliveryRefsInput {
	in.Provider = strings.TrimSpace(in.Provider)
	in.ProviderScope = strings.TrimSpace(in.ProviderScope)
	for i := range in.Refs {
		r := &in.Refs[i]
		r.MediaRef = strings.TrimSpace(r.MediaRef)
		if r.MediaUniqueRef != nil {
			t := strings.TrimSpace(*r.MediaUniqueRef)
			r.MediaUniqueRef = &t
		}
		if r.MediaType != nil {
			t := strings.TrimSpace(*r.MediaType)
			r.MediaType = &t
		}
	}
	return in
}

func (in UpsertDeliveryRefsInput) validate() error {
	if in.Provider == "" {
		return &ValidationError{Message: "provider must not be empty"}
	}
	if in.ProviderScope == "" {
		return &ValidationError{Message: "providerScope must not be empty"}
	}
	if len(in.Refs) == 0 {
		return &ValidationError{Message: "refs must contain at least one element"}
	}
	if len(in.Refs) > ReplyTemplateAlbumMaxItems {
		return &ValidationError{Message: fmt.Sprintf("refs must contain at most %d elements", ReplyTemplateAlbumMaxItems)}
	}
	for i, r := range in.Refs {
		if r.ItemID <= 0 {
			return &ValidationError{Message: fmt.Sprintf("refs[%d].itemId must be a positive integer", i)}
		}
		if r.MediaRef == "" {
			return &ValidationError{Message: fmt.Sprintf("refs[%d].mediaRef must not be empty", i)}
		}
	}
	return nil
}
