package rpc_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ariadm/internal/infra/rpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedRequest struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

func newTestClient(t *testing.T, capture *capturedRequest) *rpc.Aria2Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(capture)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"wails-dm","result":"gid-1"}`))
	}))
	t.Cleanup(server.Close)

	return rpc.NewAria2Client(server.URL)
}

func decodeOptions(t *testing.T, request capturedRequest) ([]string, map[string]string) {
	t.Helper()
	require.Len(t, request.Params, 2)

	var urls []string
	require.NoError(t, json.Unmarshal(request.Params[0], &urls))

	var options map[string]string
	require.NoError(t, json.Unmarshal(request.Params[1], &options))

	return urls, options
}

func TestAddURI_DoesNotForceResumeOptions(t *testing.T) {
	var request capturedRequest
	client := newTestClient(t, &request)

	gid, err := client.AddURI("https://example.com/file.zip", "/downloads", "file.zip")

	require.NoError(t, err)
	assert.Equal(t, "gid-1", gid)
	assert.Equal(t, "aria2.addUri", request.Method)

	urls, options := decodeOptions(t, request)
	assert.Equal(t, []string{"https://example.com/file.zip"}, urls)
	assert.Equal(t, "/downloads", options["dir"])
	assert.Equal(t, "file.zip", options["out"])
	assert.NotContains(t, options, "continue")
	assert.NotContains(t, options, "auto-file-renaming")
}

func TestResumeURI_ReusesExistingPartialFile(t *testing.T) {
	var request capturedRequest
	client := newTestClient(t, &request)

	gid, err := client.ResumeURI("https://example.com/file.zip", "/downloads", "file.zip")

	require.NoError(t, err)
	assert.Equal(t, "gid-1", gid)
	assert.Equal(t, "aria2.addUri", request.Method)

	_, options := decodeOptions(t, request)
	assert.Equal(t, "/downloads", options["dir"])
	assert.Equal(t, "file.zip", options["out"])
	assert.Equal(t, "true", options["continue"])
	assert.Equal(t, "false", options["auto-file-renaming"])
}

func TestResumeURI_OmitsEmptyFilename(t *testing.T) {
	var request capturedRequest
	client := newTestClient(t, &request)

	_, err := client.ResumeURI("https://example.com/file.zip", "/downloads", "")

	require.NoError(t, err)
	_, options := decodeOptions(t, request)
	assert.NotContains(t, options, "out")
}
