package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

