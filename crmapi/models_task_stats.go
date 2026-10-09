package crmapi

// Даты в ответах CRM - наивные ISO-строки в МСК без зоны, поэтому *string,
// а не time.Time (анмаршал time.Time требует RFC3339 с зоной).

// TaskStatMetric - именованная метрика (итог задачи или запуска расписания).
type TaskStatMetric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// TaskStatAccount - строка таблицы «По аккаунтам» подробной статистики задачи.
type TaskStatAccount struct {
	Session   string  `json:"session"`
	Phone     string  `json:"phone"`
	Username  string  `json:"username"`
	Valid     bool    `json:"valid"`
	SpamBlock bool    `json:"spam_block"`
	InTask    bool    `json:"in_task"`
	Disabled  bool    `json:"disabled"`
	Today     int64   `json:"today"`
	Total     int64   `json:"total"`
	Cut       int64   `json:"cut"`
	Errors    int64   `json:"errors"`
	Hourly    *int64  `json:"hourly"`
	LastAt    *string `json:"last_at"`
	Note      string  `json:"note"`
}

// TaskStatProblem - категория проблем задачи (разбор лога и account_events).
type TaskStatProblem struct {
	Code     string  `json:"code"`
	Label    string  `json:"label"`
	Count    int64   `json:"count"`
	Accounts int64   `json:"accounts"`
	LastAt   *string `json:"last_at"`
	Sample   string  `json:"sample"`
}

// TaskStatProblemChat - проблемный канал или чат задачи.
type TaskStatProblemChat struct {
	Title  string  `json:"title"`
	URL    string  `json:"url"`
	Reason string  `json:"reason"`
	Count  int64   `json:"count"`
	LastAt *string `json:"last_at"`
}

// TaskStatDay - значение за день (МСК).
type TaskStatDay struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
	Cut   int64  `json:"cut"`
}

// TaskStatSources - какие источники данных были доступны при сборке ответа.
type TaskStatSources struct {
	ClickHouse bool `json:"clickhouse"`
	Log        bool `json:"log"`
}

// TaskStats - подробная статистика задачи лайкера или чаттинга.
type TaskStats struct {
	TaskType     string                `json:"task_type"`
	TaskID       int64                 `json:"task_id"`
	UserID       int64                 `json:"user_id"`
	Title        *string               `json:"title"`
	Status       int                   `json:"status"`
	StatusText   string                `json:"status_text"`
	Paused       bool                  `json:"paused"`
	DateStart    *string               `json:"date_start"`
	DateEnd      *string               `json:"date_end"`
	TZ           string                `json:"tz"`
	Totals       []TaskStatMetric      `json:"totals"`
	Accounts     []TaskStatAccount     `json:"accounts"`
	Problems     []TaskStatProblem     `json:"problems"`
	ProblemChats []TaskStatProblemChat `json:"problem_chats"`
	Daily        []TaskStatDay         `json:"daily"`
	Sources      TaskStatSources       `json:"sources"`
	Notes        []string              `json:"notes"`
}

// ActivityDay - активность клиента за день по всем задачам (МСК).
type ActivityDay struct {
	Date     string `json:"date"`
	Likes    int64  `json:"likes"`
	LikesCut int64  `json:"likes_cut"`
	Chatting int64  `json:"chatting"`
}

// ActivityDailySources - доступность источников для ActivityDaily.
type ActivityDailySources struct {
	ClickHouse bool `json:"clickhouse"`
}

// ActivityDaily - динамика лайков и сообщений чаттинга клиента по дням.
type ActivityDaily struct {
	Days    []ActivityDay        `json:"days"`
	Sources ActivityDailySources `json:"sources"`
}

// ScheduleAutoStop - автоостановка запусков расписания.
type ScheduleAutoStop struct {
	Enabled bool  `json:"enabled"`
	Hours   int64 `json:"hours"`
	Minutes int64 `json:"minutes"`
}

// ScheduleRun - один запуск расписания (задача, созданная планировщиком).
// В Schedule.LastRun заполнены только task_id, status, status_text, даты.
type ScheduleRun struct {
	TaskID     int64            `json:"task_id"`
	Status     int              `json:"status"`
	StatusText string           `json:"status_text"`
	Paused     bool             `json:"paused"`
	DateStart  *string          `json:"date_start"`
	DateEnd    *string          `json:"date_end"`
	Metrics    []TaskStatMetric `json:"metrics"`
	// GapHours - часы от предыдущего запуска (nil у первого).
	GapHours *float64 `json:"gap_hours"`
	// Late - интервал больше period_days*24 + 2 часа (пропуск запуска).
	Late bool `json:"late"`
}

// Schedule - расписание (планировщик) задач клиента.
type Schedule struct {
	ID           int64            `json:"id"`
	Title        *string          `json:"title"`
	TaskType     string           `json:"task_type"`
	TaskLabel    string           `json:"task_label"`
	DonorTaskID  int64            `json:"donor_task_id"`
	OnPause      bool             `json:"on_pause"`
	RunsDone     int64            `json:"runs_done"`
	MaxRunsLimit *int64           `json:"max_runs_limit"`
	PeriodDays   int64            `json:"period_days"`
	RunTime      string           `json:"run_time"`
	Timezone     string           `json:"timezone"`
	FirstRunMode string           `json:"first_run_mode"`
	AutoStop     ScheduleAutoStop `json:"auto_stop"`
	LastRunAt    *string          `json:"last_run_at"`
	// NextRunAt - оценка CRM; nil на паузе или при исчерпанном лимите.
	NextRunAt *string      `json:"next_run_at"`
	CreatedAt *string      `json:"created_at"`
	UpdatedAt *string      `json:"updated_at"`
	LastRun   *ScheduleRun `json:"last_run"`
}

// ScheduleRuns - история запусков расписания.
type ScheduleRuns struct {
	ScheduleID int64         `json:"schedule_id"`
	TaskType   string        `json:"task_type"`
	PeriodDays int64         `json:"period_days"`
	Runs       []ScheduleRun `json:"runs"`
}
