package crmapi

type ScriptItem struct {
	ID      int64   `json:"id"`
	Title   string  `json:"title"`
	Creator *string `json:"creator,omitempty"`
}

type ScriptFull struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

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
