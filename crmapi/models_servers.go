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

// ServerVersionResult carries the user's bot worker version reported by the
// worker's own GET /version/ endpoint. Version is nil when the worker is
// down, unreachable or has no server at all — the CRM answers fast (its
// worker probe runs on a short timeout) instead of failing the request.
type ServerVersionResult struct {
	Version *string `json:"version"`
}
