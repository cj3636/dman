package transfer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.tyss.io/cj3636/dman/pkg/model"
)

// Client defines high-level HTTP operations for dman.
type Client interface {
	Compare(ctx context.Context, req model.CompareRequest, includeSame bool) ([]model.Change, error)
	BulkPublish(ctx context.Context, tar io.Reader, contentEncoding string) error
	BulkInstall(ctx context.Context, req model.CompareRequest, acceptEncoding string) (io.ReadCloser, error)
	UploadFile(ctx context.Context, user, rel string, r io.Reader) error
	DownloadFile(ctx context.Context, user, rel string) (io.ReadCloser, error)
	Status(ctx context.Context) (*model.StatusResponse, error)
	Health(ctx context.Context) (*model.HealthResponse, error)
	Prune(ctx context.Context, deletes []model.Change) (int, error) // returns number deleted
}

type httpClient struct {
	baseURL string
	token   string
	h       *http.Client
}

// New creates a new transfer client.
func New(baseURL, token string) Client {
	// no global client timeout; rely on caller context WithTimeout for precise control
	return &httpClient{baseURL: baseURL, token: token, h: &http.Client{Timeout: 0}}
}

func (c *httpClient) addAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func (c *httpClient) Compare(ctx context.Context, req model.CompareRequest, includeSame bool) ([]model.Change, error) {
	b, _ := json.Marshal(req)
	url := c.baseURL + "/compare"
	if includeSame {
		url += "?include_same=1"
	}
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	hreq.Header.Set("Content-Type", "application/json")
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("compare failed: %d", resp.StatusCode)
	}
	var changes []model.Change
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return nil, err
	}
	return changes, nil
}

func (c *httpClient) BulkPublish(ctx context.Context, tar io.Reader, contentEncoding string) error {
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/publish", tar)
	hreq.Header.Set("Content-Type", "application/x-tar")
	if contentEncoding != "" {
		hreq.Header.Set("Content-Encoding", contentEncoding)
	}
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("publish failed: %d", resp.StatusCode)
	}
	return nil
}

func (c *httpClient) BulkInstall(ctx context.Context, req model.CompareRequest, acceptEncoding string) (io.ReadCloser, error) {
	b, _ := json.Marshal(req)
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/install", bytes.NewReader(b))
	hreq.Header.Set("Content-Type", "application/json")
	if acceptEncoding != "" {
		hreq.Header.Set("Accept-Encoding", acceptEncoding)
	}
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("install failed: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (c *httpClient) UploadFile(ctx context.Context, user, rel string, r io.Reader) error {
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/upload?user=%s&path=%s", c.baseURL, user, rel), r)
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed: %d", resp.StatusCode)
	}
	return nil
}

func (c *httpClient) DownloadFile(ctx context.Context, user, rel string) (io.ReadCloser, error) {
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/download?user=%s&path=%s", c.baseURL, user, rel), nil)
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("download failed: %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (c *httpClient) Status(ctx context.Context) (*model.StatusResponse, error) {
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/status", nil)
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("status failed: %d", resp.StatusCode)
	}
	var st model.StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (c *httpClient) Health(ctx context.Context) (*model.HealthResponse, error) {
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("health failed: %d", resp.StatusCode)
	}
	var h model.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (c *httpClient) Prune(ctx context.Context, deletes []model.Change) (int, error) {
	// build delete list filtering to ChangeDelete types
	var dels []map[string]string
	for _, ch := range deletes {
		if ch.Type == model.ChangeDelete {
			dels = append(dels, map[string]string{"user": ch.User, "path": ch.Path})
		}
	}
	if len(dels) == 0 {
		return 0, nil
	}
	body, _ := json.Marshal(map[string]any{"deletes": dels})
	hreq, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/prune", bytes.NewReader(body))
	hreq.Header.Set("Content-Type", "application/json")
	c.addAuth(hreq)
	resp, err := c.h.Do(hreq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("prune failed: %d", resp.StatusCode)
	}
	// optional: parse deleted count; server returns {"deleted":N}
	var res struct {
		Deleted int `json:"deleted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		// ignore decode error; treat as success with unknown count
		return 0, nil
	}
	return res.Deleted, nil
}
