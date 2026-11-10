package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/bal16/ariadm/internal/domain/task"
)

type Aria2Client struct {
	rpcURL string
	client *http.Client
}

func NewAria2Client(url string) *Aria2Client {
	return &Aria2Client{
		rpcURL: url,
		client: &http.Client{},
	}
}

// JSON-RPC specification structures
type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// internal helper to send generic JSON-RPC payloads
func (c *Aria2Client) call(method string, params []interface{}) (json.RawMessage, error) {
	reqPayload := rpcRequest{
		JSONRPC: "2.0",
		ID:      "wails-dm",
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.rpcURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("rpc connection failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// AddURI sends a new download URL to aria2c and returns the engine's GID string
func (c *Aria2Client) AddURI(url string, downloadPath string, fileName string) (string, error) {
	return c.addURI(url, downloadPath, fileName, false)
}

// ResumeURI re-attaches to an existing partial download without letting aria2c
// auto-rename it to a new file (e.g. "file.1.zip").
func (c *Aria2Client) ResumeURI(url string, downloadPath string, fileName string) (string, error) {
	return c.addURI(url, downloadPath, fileName, true)
}

func (c *Aria2Client) addURI(url string, downloadPath string, fileName string, resume bool) (string, error) {
	// Options specific to this download task
	options := map[string]string{
		"dir": downloadPath,
	}
	if fileName != "" {
		options["out"] = fileName
	}
	if resume {
		options["continue"] = "true"
		options["auto-file-renaming"] = "false"
	}

	// aria2.addUri signature: [ [urls], {options} ]
	params := []interface{}{
		[]string{url},
		options,
	}

	rawResult, err := c.call("aria2.addUri", params)
	if err != nil {
		return "", err
	}

	var gid string
	if err := json.Unmarshal(rawResult, &gid); err != nil {
		return "", err
	}

	return gid, nil
}

// Pause stops an active download stream using its GID
func (c *Aria2Client) Pause(gid string) error {
	_, err := c.call("aria2.pause", []interface{}{gid})
	return err
}

// Unpause resumes a suspended download stream using its GID
func (c *Aria2Client) Unpause(gid string) error {
	_, err := c.call("aria2.unpause", []interface{}{gid})
	return err
}

// PauseAll halts every active and waiting download in aria2c's queue
func (c *Aria2Client) PauseAll() error {
	_, err := c.call("aria2.pauseAll", []interface{}{})
	return err
}

// UnpauseAll resumes every paused download in aria2c's queue
func (c *Aria2Client) UnpauseAll() error {
	_, err := c.call("aria2.unpauseAll", []interface{}{})
	return err
}

// ChangeGlobalOption dynamically alters running engine traits (e.g. speed limits limits)
func (c *Aria2Client) ChangeGlobalOption(options map[string]string) error {
	_, err := c.call("aria2.changeGlobalOption", []interface{}{options})
	return err
}

// Remove stops and removes an active, paused, or waiting download from aria2c's queue
func (c *Aria2Client) Remove(gid string) error {
	_, err := c.call("aria2.remove", []interface{}{gid})
	return err
}

// RemoveDownloadResult purges a completed, error, or removed entry from aria2c's in-memory result list
func (c *Aria2Client) RemoveDownloadResult(gid string) error {
	_, err := c.call("aria2.removeDownloadResult", []interface{}{gid})
	return err
}

// Ping verifies that the aria2c RPC endpoint is responsive.
func (c *Aria2Client) Ping() error {
	_, err := c.call("aria2.getVersion", []interface{}{})
	return err
}

// TellStatus fetches a live progress snapshot from aria2c for a single GID
func (c *Aria2Client) TellStatus(gid string) (*task.Aria2Status, error) {
	// Request only the fields we need to avoid oversized payloads
	rawResult, err := c.call("aria2.tellStatus", []interface{}{gid, statusKeys()})
	if err != nil {
		return nil, err
	}

	return decodeStatus(rawResult)
}

// ListDownloads returns every download that aria2c currently holds in its
// active or waiting queue. Paused downloads restored from a session file are
// reported by tellWaiting, so they are included as well.
func (c *Aria2Client) ListDownloads() ([]*task.Aria2Status, error) {
	keys := statusKeys()

	activeRaw, err := c.call("aria2.tellActive", []interface{}{keys})
	if err != nil {
		return nil, err
	}
	active, err := decodeStatusList(activeRaw)
	if err != nil {
		return nil, err
	}

	// A large page size keeps adoption of session-restored downloads in one
	// round trip. Realistic queues are far below this limit.
	waitingRaw, err := c.call("aria2.tellWaiting", []interface{}{0, 1000, keys})
	if err != nil {
		return nil, err
	}
	waiting, err := decodeStatusList(waitingRaw)
	if err != nil {
		return nil, err
	}

	return append(active, waiting...), nil
}

func statusKeys() []string {
	return []string{"gid", "status", "totalLength", "completedLength", "downloadSpeed", "files"}
}

type statusPayload struct {
	GID             string `json:"gid"`
	Status          string `json:"status"`
	TotalLength     string `json:"totalLength"`
	CompletedLength string `json:"completedLength"`
	DownloadSpeed   string `json:"downloadSpeed"`
	Files           []struct {
		Path string `json:"path"`
		URIs []struct {
			URI string `json:"uri"`
		} `json:"uris"`
	} `json:"files"`
}

func decodeStatus(rawResult json.RawMessage) (*task.Aria2Status, error) {
	var raw statusPayload
	if err := json.Unmarshal(rawResult, &raw); err != nil {
		return nil, fmt.Errorf("tellStatus: failed to decode aria2 response: %w", err)
	}
	return toAria2Status(raw), nil
}

func decodeStatusList(rawResult json.RawMessage) ([]*task.Aria2Status, error) {
	var raw []statusPayload
	if err := json.Unmarshal(rawResult, &raw); err != nil {
		return nil, fmt.Errorf("tellStatus: failed to decode aria2 response list: %w", err)
	}

	statuses := make([]*task.Aria2Status, 0, len(raw))
	for _, item := range raw {
		statuses = append(statuses, toAria2Status(item))
	}
	return statuses, nil
}

func toAria2Status(raw statusPayload) *task.Aria2Status {
	parseI64 := func(s string) int64 {
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}

	// Extract just the base filename from the full filesystem path
	fileName := ""
	downloadURL := ""
	var filePaths []string
	if len(raw.Files) > 0 {
		for _, f := range raw.Files {
			if f.Path != "" {
				filePaths = append(filePaths, f.Path)
			}
			if downloadURL == "" && len(f.URIs) > 0 {
				downloadURL = f.URIs[0].URI
			}
		}
		if len(filePaths) > 0 {
			fileName = filepath.Base(filePaths[0])
		}
	}

	return &task.Aria2Status{
		GID:             raw.GID,
		Status:          raw.Status,
		URL:             downloadURL,
		TotalLength:     parseI64(raw.TotalLength),
		CompletedLength: parseI64(raw.CompletedLength),
		DownloadSpeed:   parseI64(raw.DownloadSpeed),
		FileName:        fileName,
		Files:           filePaths,
	}
}
