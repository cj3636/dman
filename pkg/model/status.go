package model

// StatusUser summarizes per-user stored file counts and bytes.
type StatusUser struct {
	User  string `json:"user"`
	Files int    `json:"files"`
	Bytes int64  `json:"bytes"`
}

// StatusResponse is returned by /status.
type StatusResponse struct {
	FilesTotal  int64             `json:"files_total"`
	BytesTotal  int64             `json:"bytes_total"`
	Users       []StatusUser      `json:"users"`
	LastPublish string            `json:"last_publish,omitempty"`
	LastInstall string            `json:"last_install,omitempty"`
	Metrics     map[string]uint64 `json:"metrics,omitempty"`
}
