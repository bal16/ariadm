package httpserver

import (
	"ariadm/internal/domain/task"
	"encoding/json"
	"net"
	"net/http"
)

type HTTPServer struct {
	server      *http.Server
	taskService *task.TaskService
	port        string
}

func NewHTTPServer(port string, ts *task.TaskService) *HTTPServer {
	return &HTTPServer{
		port:        port,
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

// Start binds to the local address and launches the HTTP service listener
func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/download", s.HandleDownload)

	s.server = &http.Server{
		Addr:    net.JoinHostPort("127.0.0.1", s.port),
		Handler: mux,
	}

	// Run the listener inside a goroutine so it doesn't block the Wails UI runtime thread
	go func() {
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			// In production, wire this to a proper logger system
			println("Local HTTP Server crash details:", err.Error())
		}
	}()

	return nil
}

// Stop gracefully shuts down the local listener port
func (s *HTTPServer) Stop() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}
