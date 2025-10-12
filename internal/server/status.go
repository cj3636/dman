package server

import (
	"strings"

	"git.tyss.io/cj3636/dman/internal/storage"
	"git.tyss.io/cj3636/dman/pkg/model"
)

// buildStatus assembles a StatusResponse by traversing stored files.
func buildStatus(store storage.Backend, meta *Meta) (model.StatusResponse, error) {
	files, err := store.List()
	if err != nil {
		return model.StatusResponse{}, err
	}
	perUserFiles := map[string]int{}
	perUserBytes := map[string]int64{}
	var totalBytes int64
	for _, rel := range files { // rel = user/path
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) != 2 {
			continue
		}
		user, p := parts[0], parts[1]
		f, err := store.Open(user, p)
		if err != nil {
			continue
		}
		fi, _ := f.Stat()
		f.Close()
		perUserFiles[user]++
		perUserBytes[user] += fi.Size()
		totalBytes += fi.Size()
	}
	var users []model.StatusUser
	for u, fc := range perUserFiles {
		users = append(users, model.StatusUser{User: u, Files: fc, Bytes: perUserBytes[u]})
	}
	lastPub, lastInst, metrics := meta.snapshot()
	resp := model.StatusResponse{
		FilesTotal:  int64(len(files)),
		BytesTotal:  totalBytes,
		Users:       users,
		LastPublish: lastPub,
		LastInstall: lastInst,
		Metrics:     metrics,
	}
	return resp, nil
}
