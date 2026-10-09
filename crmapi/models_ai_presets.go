package crmapi

// AIFunctionSettings - «ИИ-настройки» профиля клиента по функции.
// MaxChars - лимит длины ответа в символах (0 - дефолт воркера).
type AIFunctionSettings struct {
	Model           *string `json:"model"`
	SystemPrompt    *string `json:"system_prompt"`
	ForbiddenTopics *string `json:"forbidden_topics"`
	Reasoning       bool    `json:"reasoning"`
	MaxChars        int64   `json:"max_chars"`
	UpdatedAt       *string `json:"updated_at"`
}

// AIPreset - именованный пресет ИИ-настроек клиента. Секрет BYOK-ключа CRM
// не отдаёт, только HasBYOK.
type AIPreset struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	Model           *string `json:"model"`
	SystemPrompt    *string `json:"system_prompt"`
	ForbiddenTopics *string `json:"forbidden_topics"`
	Reasoning       bool    `json:"reasoning"`
	MaxChars        int64   `json:"max_chars"`
	KeyMode         string  `json:"key_mode"`
	HasBYOK         bool    `json:"has_byok"`
	CreatedAt       *string `json:"created_at"`
	TasksTotal      int64   `json:"tasks_total"`
	TasksActive     int64   `json:"tasks_active"`
}

// AIEffectiveSettings - настройки, с которыми воркер генерирует в задаче.
// *Source: preset | settings | default; Model nil при ModelSource=default
// (модель сервиса по умолчанию). KeySource: preset | account.
type AIEffectiveSettings struct {
	Model           *string `json:"model"`
	ModelSource     string  `json:"model_source"`
	PromptSource    string  `json:"prompt_source"`
	ForbiddenTopics *string `json:"forbidden_topics"`
	ForbiddenSource string  `json:"forbidden_source"`
	Reasoning       bool    `json:"reasoning"`
	ReasoningSource string  `json:"reasoning_source"`
	MaxChars        int64   `json:"max_chars"`
	MaxCharsSource  string  `json:"max_chars_source"`
	KeyMode         string  `json:"key_mode"`
	KeySource       string  `json:"key_source"`
}

// AITask - ИИ-задача клиента (комментинг с нейро-комментингом или чаттинг в
// режиме ИИ) с пресетом и действующими настройками.
type AITask struct {
	TaskID     int64   `json:"task_id"`
	Title      *string `json:"title"`
	Status     int     `json:"status"`
	StatusText string  `json:"status_text"`
	Active     bool    `json:"active"`
	Paused     bool    `json:"paused"`
	DateStart  *string `json:"date_start"`
	DateEnd    *string `json:"date_end"`
	PresetID   *int64  `json:"preset_id"`
	PresetName *string `json:"preset_name"`
	// PresetMissing - у задачи есть ai_preset_id, но пресета уже нет.
	PresetMissing bool                `json:"preset_missing"`
	Effective     AIEffectiveSettings `json:"effective"`
}

// AIUsedModel - фактические генерации модели за окно (леджер CRM).
type AIUsedModel struct {
	Model  *string `json:"model"`
	Count  int64   `json:"count"`
	LastAt *string `json:"last_at"`
}

// AIPresetsFunction - ИИ-настройки одной функции (comment | chatting).
type AIPresetsFunction struct {
	Function   string              `json:"function"`
	KeyMode    string              `json:"key_mode"`
	Settings   *AIFunctionSettings `json:"settings"`
	Presets    []AIPreset          `json:"presets"`
	Tasks      []AITask            `json:"tasks"`
	TasksTotal int64               `json:"tasks_total"`
	UsedModels []AIUsedModel       `json:"used_models"`
}

// AIPresetsOverview - ИИ-пресеты и настройки клиента по функциям.
type AIPresetsOverview struct {
	Functions []AIPresetsFunction `json:"functions"`
	// Days - окно UsedModels в днях.
	Days int `json:"days"`
}
