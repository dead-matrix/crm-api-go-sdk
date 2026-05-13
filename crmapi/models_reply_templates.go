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
type CreateReplyTemplateInput struct {
	Title string              `json:"title"`
	Kind  string              `json:"kind"`
	Items []ReplyTemplateItem `json:"items"`
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
