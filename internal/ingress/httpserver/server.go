package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/bal16/ariadm/internal/domain/task"
)

type HTTPServer struct {
	server      *http.Server
	taskService *task.TaskService
	ports       []int
	port        int
}

func NewHTTPServer(ports []int, ts *task.TaskService) *HTTPServer {
	return &HTTPServer{
		ports:       ports,
		taskService: ts,
	}
}

// Request payload format expected from the browser extension
type downloadRequest struct {
	URL string `json:"url"`
}

// Response payload format returned to the extension
type downloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	TaskID  string `json:"task_id,omitempty"`
}

func (s *HTTPServer) HandleDownload(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 1. Enforce POST requests only
	if r.Method != http.MethodPost {
		s.writeJSON(w, http.StatusMethodNotAllowed, downloadResponse{Success: false, Message: "Method not allowed"})
		return
	}

	// 2. Decode the incoming JSON payload
	var req downloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeJSON(w, http.StatusBadRequest, downloadResponse{Success: false, Message: "Invalid request payload"})
		return
	}

	if req.URL == "" {
		s.writeJSON(w, http.StatusBadRequest, downloadResponse{Success: false, Message: "URL parameter is required"})
		return
	}

	// 3. Hand the URL down to the TDD-validated domain service layer
	t, err := s.taskService.DownloadFile(req.URL)
	if err != nil {
		s.writeJSON(w, http.StatusUnprocessableEntity, downloadResponse{Success: false, Message: err.Error()})
		return
	}

	// 4. Return the successful creation tracking details
	s.writeJSON(w, http.StatusCreated, downloadResponse{
		Success: true,
		Message: "Download successfully added to the engine queue",
		TaskID:  t.ID,
	})
}

// HandleHealth lets the browser extension verify that it found ariadm and not
// some unrelated service that happens to occupy a candidate port.
func (s *HTTPServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"app":    "ariadm",
	})
}

func (s *HTTPServer) enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// Internal utility helper to write JSON responses uniformly
func (s *HTTPServer) writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

// Start binds the first free port in the configured range and serves it in the
// background. Binding is attempted synchronously so callers can report a real
// failure instead of claiming to be listening.
func (s *HTTPServer) Start() (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/download", s.HandleDownload)
	mux.HandleFunc("/health", s.HandleHealth)

	for _, port := range s.ports {
		listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			continue
		}

		s.port = port
		s.server = &http.Server{Handler: mux}

		go func() {
			if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				println("Local HTTP Server crash details:", err.Error())
			}
		}()

		return port, nil
	}

	return 0, fmt.Errorf("no usable HTTP port in range %v", s.ports)
}

// Port returns the port that Start bound, or zero before a successful start.
func (s *HTTPServer) Port() int {
	return s.port
}

// Stop gracefully shuts down the local listener port
func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
