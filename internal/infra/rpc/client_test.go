package rpc_test

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/bal16/ariadm/internal/infra/rpc"

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

func newJSONClient(t *testing.T, response string, capture *capturedRequest) *rpc.Aria2Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(capture)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	return rpc.NewAria2Client(server.URL)
}

func TestPing_UsesGetVersion(t *testing.T) {
	var request capturedRequest
	client := newJSONClient(t, `{"jsonrpc":"2.0","id":"wails-dm","result":{"version":"1.37.0"}}`, &request)

	require.NoError(t, client.Ping())
	assert.Equal(t, "aria2.getVersion", request.Method)
}

func TestTellStatus_ParsesSourceURL(t *testing.T) {
	var request capturedRequest
	client := newJSONClient(t, `{
		"jsonrpc":"2.0",
		"id":"wails-dm",
		"result":{
			"gid":"abc123",
			"status":"active",
			"totalLength":"2048",
			"completedLength":"1024",
			"downloadSpeed":"512",
			"files":[{
				"path":"/downloads/archive.zip",
				"uris":[{"status":"used","uri":"https://example.com/archive.zip"}]
			}]
		}
	}`, &request)

	status, err := client.TellStatus("abc123")

	require.NoError(t, err)
	assert.Equal(t, "https://example.com/archive.zip", status.URL)
	assert.Equal(t, "archive.zip", status.FileName)
	assert.Equal(t, int64(2048), status.TotalLength)
	assert.Equal(t, int64(1024), status.CompletedLength)
	assert.Equal(t, int64(512), status.DownloadSpeed)
	assert.Equal(t, "aria2.tellStatus", request.Method)
}

func TestListDownloads_CombinesActiveAndWaiting(t *testing.T) {
	var methods []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request capturedRequest
		_ = json.NewDecoder(r.Body).Decode(&request)
		methods = append(methods, request.Method)

		w.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case "aria2.tellActive":
			_, _ = w.Write([]byte(`{
				"jsonrpc":"2.0","id":"wails-dm","result":[{
					"gid":"active_gid","status":"active","totalLength":"4096",
					"completedLength":"2048","downloadSpeed":"256",
					"files":[{"path":"/downloads/a.zip","uris":[{"status":"used","uri":"https://example.com/a.zip"}]}]
				}]
			}`))
		case "aria2.tellWaiting":
			// The paused download must be requested from the waiting queue.
			assert.Len(t, request.Params, 3)
			var offset int
			var limit int
			require.NoError(t, json.Unmarshal(request.Params[0], &offset))
			require.NoError(t, json.Unmarshal(request.Params[1], &limit))
			assert.Equal(t, 0, offset)
			assert.Equal(t, 1000, limit)

			_, _ = w.Write([]byte(`{
				"jsonrpc":"2.0","id":"wails-dm","result":[{
					"gid":"paused_gid","status":"paused","totalLength":"0",
					"completedLength":"0","downloadSpeed":"0",
					"files":[{"path":"/downloads/b.zip","uris":[{"status":"waiting","uri":"https://example.com/b.zip"}]}]
				}]
			}`))
		default:
			t.Errorf("unexpected RPC method: %s", request.Method)
		}
	}))
	defer server.Close()

	client := rpc.NewAria2Client(server.URL)
	downloads, err := client.ListDownloads()

	require.NoError(t, err)
	require.Len(t, downloads, 2)
	assert.Equal(t, []string{"aria2.tellActive", "aria2.tellWaiting"}, methods)

	assert.Equal(t, "active_gid", downloads[0].GID)
	assert.Equal(t, "https://example.com/a.zip", downloads[0].URL)
	assert.Equal(t, "a.zip", downloads[0].FileName)

	assert.Equal(t, "paused_gid", downloads[1].GID)
	assert.Equal(t, "paused", downloads[1].Status)
	assert.Equal(t, "https://example.com/b.zip", downloads[1].URL)
}

func TestSetPort_RedirectsRequests(t *testing.T) {
	var firstHits, secondHits int

	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"wails-dm","result":{"version":"1.37.0"}}`))
	}))
	defer first.Close()

	second := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"wails-dm","result":{"version":"1.37.0"}}`))
	}))
	defer second.Close()

	client := rpc.NewAria2Client(first.URL)
	require.NoError(t, client.Ping())

	_, portValue, err := net.SplitHostPort(strings.TrimPrefix(second.URL, "http://"))
	require.NoError(t, err)
	port, err := strconv.Atoi(portValue)
	require.NoError(t, err)

	client.SetPort(port)
	require.NoError(t, client.Ping())

	assert.Equal(t, 1, firstHits)
	assert.Equal(t, 1, secondHits)
}
