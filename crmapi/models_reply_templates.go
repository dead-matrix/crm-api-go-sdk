package crmapi

import "time"

// ReplyTemplateCreator describes the staff member who created a reply
// template. Name is nil when the staff record was deleted; for the
// system creator (EmployeeID = 0) the CRM substitutes the literal
// string "TraffSoft" — clients can rely on Name being non-nil in that
// special case.
type ReplyTemplateCreator struct {
	EmployeeID int64   `json:"employeeId"`
	Name       *string `json:"name"`
}

// ReplyTemplatePreview is a short summary shown by the messenger UI in
// the templates list. FirstItemType is nil when the template has zero
// items (defensive — shouldn't happen for a saved template).
type ReplyTemplatePreview struct {
	FirstItemType  *string `json:"firstItemType"`
	CaptionExcerpt *string `json:"captionExcerpt"`
	ItemsCount     int64   `json:"itemsCount"`
}

// ReplyTemplateListItem is the per-row payload returned by
// ReplyTemplatesList. It deliberately omits items[] to keep the list
// endpoint cheap — fetch ReplyTemplatesGet for the full content.
type ReplyTemplateListItem struct {
	ID         int64                 `json:"id"`
	PublicID   string                `json:"publicId"`
	Title      string                `json:"title"`
	Kind       string                `json:"kind"`
	Creator    ReplyTemplateCreator  `json:"creator"`
	Preview    ReplyTemplatePreview  `json:"preview"`
	UsageCount int64                 `json:"usageCount"`
	LastUsedAt *time.Time
	CreatedAt  *time.Time
	UpdatedAt  *time.Time
}

// ReplyTemplateItem is one element of a template (a single piece of
// content). The same shape is accepted in CreateReplyTemplateInput and
// returned in ReplyTemplateFull.Items, sorted by Position.
//
// For Type="text": Caption holds the body, all media-related fields
// must be nil.
// For non-text types: MediaObjectKey is required and points to a
// stable S3 key (the messenger resolves it to a presigned URL at send
// time). Telegram file_id MUST NOT be persisted in MediaObjectKey —
// it's bot-scoped and breaks when bots rotate.
// For Type ∈ {"voice","video_note","sticker"}: Caption must be nil
// (Telegram strips it).
type ReplyTemplateItem struct {
	// ID is the per-row database identifier. Returned by the server
	// in `GET /reply-templates/{id}`; ignored in create requests
	// (the server assigns one). Clients that cache derived state
	// per-item (e.g. messenger's Telegram file_id reuse) MUST key
	// by this id - `position` alone is fragile under future
	// reorder operations.
	ID              int64   `json:"id,omitempty"`
	Position        int     `json:"position"`
	Type            string  `json:"type"`
	Caption         *string `json:"caption,omitempty"`
	MediaObjectKey  *string `json:"mediaObjectKey,omitempty"`
	Mime            *string `json:"mime,omitempty"`
	SizeBytes       *int64  `json:"sizeBytes,omitempty"`
	DurationMs      *int64  `json:"durationMs,omitempty"`
	Width           *int    `json:"width,omitempty"`
	Height          *int    `json:"height,omitempty"`
	FileName        *string `json:"fileName,omitempty"`
	OriginTenantID  *string `json:"originTenantId,omitempty"`
	OriginMessageID *string `json:"originMessageId,omitempty"`
}

// ─── Delivery refs (reusable file-id cache) ────────────────────────────
//
// `provider`/`provider_scope` pin a Telegram-like ID to a specific
// bot/account so it doesn't leak across providers or bots.
// Currently `provider` is always `"telegram"`. `provider_scope` for
// Telegram is the bot username (file_ids are bot-specific).
//
// The cache is a pure accelerator: when the messenger sends a media
// reply-template item via the URL it persists the resulting Telegram
// file_id here; subsequent sends to the same provider+scope reuse
// the file_id instead of re-downloading + re-presigning. Stale refs
// (Telegram error "wrong file identifier") cause the messenger to
// fall back to the URL path and upsert a fresh ref.

const (
	DeliveryProviderTelegram = "telegram"
)

// DeliveryRef is one persisted reusable handle for one
// (template_item_id, provider, provider_scope) tuple.
type DeliveryRef struct {
	ID             int64   `json:"id"`
	ItemID         int64   `json:"itemId"`
	Provider       string  `json:"provider"`
	ProviderScope  string  `json:"providerScope"`
	MediaRef       string  `json:"mediaRef"`
	MediaUniqueRef *string `json:"mediaUniqueRef,omitempty"`
	MediaType      *string `json:"mediaType,omitempty"`
	FailCount      int     `json:"failCount"`
	LastUsedAt     *string `json:"lastUsedAt,omitempty"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	UpdatedAt      *string `json:"updatedAt,omitempty"`
}

// UpsertDeliveryRefInput is one ref entry in the PUT batch.
type UpsertDeliveryRefInput struct {
	ItemID         int64   `json:"itemId"`
	MediaRef       string  `json:"mediaRef"`
	MediaUniqueRef *string `json:"mediaUniqueRef,omitempty"`
	MediaType      *string `json:"mediaType,omitempty"`
}

// UpsertDeliveryRefsInput is the body of
// PUT /reply-templates/{template_id}/delivery-refs.
//
// Server contract: provider non-empty, provider_scope non-empty,
// refs 1..10 items, every ItemID belongs to the addressed template.
// Idempotent: upsert keyed by (template_item_id, provider, scope).
// Successful upsert always resets server-side fail_count to 0.
type UpsertDeliveryRefsInput struct {
	Provider      string                   `json:"provider"`
	ProviderScope string                   `json:"providerScope"`
	Refs          []UpsertDeliveryRefInput `json:"refs"`
}

// ReplyTemplateFull is returned by ReplyTemplatesGet. Items are sorted
// by Position ascending.
type ReplyTemplateFull struct {
	ID        int64                 `json:"id"`
	PublicID  string                `json:"publicId"`
	Title     string                `json:"title"`
	Kind      string                `json:"kind"`
	Creator   ReplyTemplateCreator  `json:"creator"`
	Items     []ReplyTemplateItem   `json:"items"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// CreateReplyTemplateInput is the body of ReplyTemplatesCreate. Server-
// side validation duplicates the client-side rules enforced here.
//
// kind="single" requires exactly 1 item; kind="album" requires 2..10
// items, all of Type ∈ {"photo","video","gif"}, with Caption allowed
// only on Position=0.
//
// PublicID is optional. Trusted clients (e.g. the messenger's
// save-as-template flow) mint a UUID locally so the same identifier
// can drive the S3 object-key path (`shared/reply-templates/{public_id}/...`)
// before the CRM POST happens — that turns "copy media then create
// template" into an atomic, single-round-trip sequence. Empty string
// means "let the CRM generate one"; a duplicate value surfaces as an
// APIError with HTTP 409 from upstream.
type CreateReplyTemplateInput struct {
	Title    string              `json:"title"`
	Kind     string              `json:"kind"`
	PublicID string              `json:"publicId,omitempty"`
	Items    []ReplyTemplateItem `json:"items"`
}

// DeleteReplyTemplateResult is what ReplyTemplatesDelete returns on
// successful removal — the ids of the just-deleted template, useful
// for cache invalidation on the messenger side.
type DeleteReplyTemplateResult struct {
	ID       int64  `json:"id"`
	PublicID string `json:"publicId"`
}

// Allowed values for ReplyTemplateItem.Type. Exported as a set for
// callers that need to render UI selectors without reaching into the
// SDK.
const (
	ReplyTemplateItemTypeText      = "text"
	ReplyTemplateItemTypePhoto     = "photo"
	ReplyTemplateItemTypeVideo     = "video"
	ReplyTemplateItemTypeGIF       = "gif"
	ReplyTemplateItemTypeVoice     = "voice"
	ReplyTemplateItemTypeVideoNote = "video_note"
	ReplyTemplateItemTypeSticker   = "sticker"
	ReplyTemplateItemTypeFile      = "file"
)

// Allowed values for CreateReplyTemplateInput.Kind.
const (
	ReplyTemplateKindSingle = "single"
	ReplyTemplateKindAlbum  = "album"
)

// Hard limits enforced by the server (kept in sync with CRM Pydantic).
const (
	ReplyTemplateTitleMaxLength = 255
	ReplyTemplateCaptionMax     = 4096
	ReplyTemplateAlbumMinItems  = 2
	ReplyTemplateAlbumMaxItems  = 10
	ReplyTemplateItemMaxPos     = 9
)
