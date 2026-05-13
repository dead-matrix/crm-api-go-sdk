package crmapi

// Sales-decks payload types. ScriptItem / ScriptFull (legacy text quick
// replies) were removed in Phase 5 along with the ScriptsList /
// ScriptsGet / ScriptsCreate client methods. For quick replies use
// reply_templates.go (ReplyTemplateFull, etc.) instead.

type PriceMediaItem struct {
	Text  string   `json:"text"`
	Media []string `json:"media"`
}

type ToolsMediaItem struct {
	VideoURL string `json:"video_url"`
	FileID   string `json:"file_id"`
}

type ToolsMediaResult struct {
	Text  string           `json:"text"`
	Media []ToolsMediaItem `json:"media"`
}
