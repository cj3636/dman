package model

type UserSpec struct {
	Name  string   `json:"name"`
	Home  string   `json:"home"`
	Track []string `json:"track"`
}

type InventoryItem struct {
	User  string `json:"user"`
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime_unix"`
	Hash  string `json:"sha256"`
	IsDir bool   `json:"is_dir"`
}

type CompareRequest struct {
	Users     []string        `json:"users"`
	Inventory []InventoryItem `json:"inventory"`
}

type ChangeType string

const (
	ChangeAdd    ChangeType = "add"
	ChangeModify ChangeType = "modify"
	ChangeDelete ChangeType = "delete"
	ChangeSame   ChangeType = "same"
)

type Change struct {
	User string     `json:"user"`
	Path string     `json:"path"`
	Type ChangeType `json:"type"`
}
