package model

// HealthResponse represents server liveness + build metadata.
type HealthResponse struct {
	OK         bool   `json:"ok"`
	Version    string `json:"version,omitempty"`
	BuildTime  string `json:"build_time,omitempty"`
	Commit     string `json:"commit,omitempty"`
	ServerTime string `json:"server_time"`
}
