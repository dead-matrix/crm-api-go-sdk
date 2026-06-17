package crmapi

type ServerRestartResult struct {
	Message string `json:"message"`
}

// ServerStatusResult reports the readiness of a user's bot worker, used to
// poll the actual completion of a restart (rather than the moment the open
// command was accepted).
//
//	Bound — a process is listening on the worker port (worker started).
//	Up    — the worker answers /ping/ (FastAPI is up and serving). Once Up is
//	        true after a restart, the old process (and its active tasks) has
//	        been killed and replaced by the new one.
type ServerStatusResult struct {
	Bound bool `json:"bound"`
	Up    bool `json:"up"`
}
