package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"ariadm/internal/domain/task"
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
	defer resp.Body.Close()

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
func (c *Aria2Client) AddURI(url string, downloadPath string) (string, error) {
	// Options specific to this download task
	options := map[string]string{
		"dir": downloadPath,
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

// TellStatus fetches a live progress snapshot from aria2c for a single GID
func (c *Aria2Client) TellStatus(gid string) (*task.Aria2Status, error) {
	// Request only the fields we need to avoid oversized payloads
	keys := []string{"gid", "status", "totalLength", "completedLength", "downloadSpeed", "files"}
	rawResult, err := c.call("aria2.tellStatus", []interface{}{gid, keys})
	if err != nil {
		return nil, err
	}

	// aria2 returns all numeric fields as strings — we decode into a raw map first
	var raw struct {
		GID             string `json:"gid"`
		Status          string `json:"status"`
		TotalLength     string `json:"totalLength"`
		CompletedLength string `json:"completedLength"`
		DownloadSpeed   string `json:"downloadSpeed"`
		Files           []struct {
			Path string `json:"path"`
		} `json:"files"`
	}
	if err := json.Unmarshal(rawResult, &raw); err != nil {
		return nil, fmt.Errorf("tellStatus: failed to decode aria2 response: %w", err)
	}

	parseI64 := func(s string) int64 {
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}

	// Extract just the base filename from the full filesystem path
	fileName := ""
	var filePaths []string
	if len(raw.Files) > 0 {
		for _, f := range raw.Files {
			if f.Path != "" {
				filePaths = append(filePaths, f.Path)
			}
		}
		if len(filePaths) > 0 {
			fileName = filepath.Base(filePaths[0])
		}
	}

	return &task.Aria2Status{
		GID:             raw.GID,
		Status:          raw.Status,
		TotalLength:     parseI64(raw.TotalLength),
		CompletedLength: parseI64(raw.CompletedLength),
		DownloadSpeed:   parseI64(raw.DownloadSpeed),
		FileName:        fileName,
		Files:           filePaths,
	}, nil
}

